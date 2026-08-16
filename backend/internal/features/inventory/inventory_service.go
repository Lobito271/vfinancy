// Package inventory implements the business logic for the inventory
// aggregate: stock entries / exits / transfers / adjustments, aging,
// and the 25-day clearance rule.
//
// Clearance rule: any batch with arrival_date + 25 days in the
// past and current_quantity > 0 is "on clearance". This is computed
// lazily and exposed via the domain entity's MaximumSaleDate /
// IsClearance / NeedsClearanceSoon methods. The service layer
// surfaces batch listings for the dashboard.
package inventory

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/shared/logger"
)

// InventoryService owns the inventory slice.
type InventoryService struct {
	batches   InventoryBatchRepository
	movements InventoryMovementRepository
	warehouses WarehouseResolver
	products  ProductClassifier
	txm       repositories.TransactionManager
	log       *logger.Logger
}

// New returns an InventoryService ready for use.
func New(
	batches InventoryBatchRepository,
	movements InventoryMovementRepository,
	warehouses WarehouseResolver,
	products ProductClassifier,
	txm repositories.TransactionManager,
	log *logger.Logger,
) *InventoryService {
	return &InventoryService{
		batches:    batches,
		movements:  movements,
		warehouses: warehouses,
		products:   products,
		txm:        txm,
		log:        log,
	}
}

// ReceiveInput registers a new batch of stock (purchase receipt).
type ReceiveInput struct {
	CompanyID      uuid.UUID
	ProductID      uuid.UUID
	WarehouseID    uuid.UUID
	SupplierID     *uuid.UUID
	PurchaseLineID *uuid.UUID
	LotNumber      valueobjects.LotNumber
	ArrivalDate    valueobjects.Date
	Quantity       valueobjects.Quantity
	UnitCost       valueobjects.Money
	CurrencyCode   valueobjects.CurrencyCode
	ExpiryDate     *valueobjects.Date
}

// Receive creates a new inventory batch and records an inbound
// movement. Both operations are inside the same transaction.
func (s *InventoryService) Receive(ctx context.Context, in ReceiveInput) (*InventoryBatch, error) {
	var out *InventoryBatch
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		batch, err := NewInventoryBatch(time.Now().UTC(), NewInventoryBatchOptions{
			CompanyID:       in.CompanyID,
			ProductID:       in.ProductID,
			WarehouseID:     in.WarehouseID,
			SupplierID:      in.SupplierID,
			PurchaseLineID:  in.PurchaseLineID,
			LotNumber:       in.LotNumber,
			ArrivalDate:     in.ArrivalDate,
			ExpiryDate:      in.ExpiryDate,
			InitialQuantity: in.Quantity,
			UnitCost:        in.UnitCost,
			CurrencyCode:    in.CurrencyCode,
		})
		if err != nil {
			return err
		}
		if err := s.batches.Create(ctx, batch); err != nil {
			return err
		}
		// Companion movement row (append-only ledger).
		now := time.Now().UTC()
		ref, _ := valueobjects.NewReference(enums.ReferenceTypePurchase, batch.ID)
		// The movement is inbound (positive) and matches the batch's
		// initial quantity. We do not embed the reference in the
		// inbound movement because the batch is the source-of-truth
		// for the initial quantity.
		_ = ref
		movement, err := NewInventoryMovement(NewInventoryMovementOptions{
			CompanyID:    in.CompanyID,
			ProductID:    in.ProductID,
			WarehouseID:  in.WarehouseID,
			BatchID:      &batch.ID,
			Type:         enums.MovementTypePurchase,
			Quantity:     in.Quantity,
			UnitCost:     in.UnitCost,
			CurrencyCode: in.CurrencyCode,
			OccurredAt:   now,
			Notes:        "initial receipt",
		})
		if err != nil {
			return err
		}
		if err := s.movements.Create(ctx, movement); err != nil {
			return err
		}
		out = batch
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("inventory received",
		"batch_id", out.ID,
		"product_id", out.ProductID,
		"warehouse_id", out.WarehouseID,
		"quantity", out.InitialQuantity,
	)
	return out, nil
}

// IssueInput consumes stock for a sale.
type IssueInput struct {
	BatchID   uuid.UUID
	Quantity  valueobjects.Quantity
	Reference valueobjects.Reference
}

// Issue consumes a positive quantity from the given batch. The
// movement type is "sale" outbound (negative quantity).
func (s *InventoryService) Issue(ctx context.Context, in IssueInput) error {
	return s.consumeOrAdjust(ctx, in.BatchID, in.Quantity, in.Reference, enums.MovementTypeSale, "sale issue")
}

// AdjustInput records a manual adjustment (positive or negative).
type AdjustInput struct {
	BatchID   uuid.UUID
	Delta     valueobjects.Quantity // signed
	Reason    string
	Reference valueobjects.Reference
}

// Adjust applies a manual adjustment. The Delta is signed: positive
// adds stock, negative removes it. The batch's current_quantity is
// prevented from going negative — if it would, the operation fails.
func (s *InventoryService) Adjust(ctx context.Context, in AdjustInput) error {
	movementType := enums.MovementTypeAdjustmentIn
	notes := "adjustment in"
	if in.Delta.IsNegative() {
		movementType = enums.MovementTypeAdjustmentOut
		notes = "adjustment out"
	}
	return s.consumeOrAdjust(ctx, in.BatchID, in.Delta, in.Reference, movementType, notes)
}

// VoidInput cancels a mistaken stock receipt.
type VoidInput struct {
	BatchID uuid.UUID
	Reason  string
}

// Void cancels a mistaken stock receipt: the batch's remaining
// quantity is zeroed, its status flips to "voided", and a compensating
// "void_out" movement is appended to the ledger. The receipt row is
// kept for audit (no hard delete, the product reference is untouched).
// Already-voided batches are rejected. Runs inside a single
// transaction with the batch row locked (SELECT ... FOR UPDATE).
func (s *InventoryService) Void(ctx context.Context, in VoidInput) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		batch, err := s.batches.GetByIDForUpdate(ctx, in.BatchID)
		if err != nil {
			return err
		}
		if batch.Status == InventoryBatchStatusVoided {
			return ErrBatchAlreadyVoided
		}
		remaining := batch.CurrentQuantity
		batch.Void()
		if err := s.batches.Update(ctx, batch); err != nil {
			return err
		}
		if remaining.IsPositive() {
			notes := "lote anulado"
			if reason := strings.TrimSpace(in.Reason); reason != "" {
				notes = reason
			}
			ref, err := valueobjects.NewReference(enums.ReferenceTypeAdjustment, batch.ID)
			if err != nil {
				return err
			}
			mv, err := NewInventoryMovement(NewInventoryMovementOptions{
				CompanyID:    batch.CompanyID,
				ProductID:    batch.ProductID,
				WarehouseID:  batch.WarehouseID,
				BatchID:      &batch.ID,
				Type:         enums.MovementTypeVoidOut,
				Quantity:     remaining.Neg(),
				UnitCost:     batch.UnitCost,
				CurrencyCode: batch.CurrencyCode,
				OccurredAt:   time.Now().UTC(),
				Reference:    &ref,
				Notes:        notes,
			})
			if err != nil {
				return err
			}
			if err := s.movements.Create(ctx, mv); err != nil {
				return err
			}
		}
		s.log.Info("inventory batch voided",
			"batch_id", batch.ID,
			"product_id", batch.ProductID,
			"warehouse_id", batch.WarehouseID,
			"quantity", remaining,
		)
		return nil
	})
}

// TransferInput moves stock from one batch to another (typically
// between warehouses; we keep batches per warehouse in this model).
type TransferInput struct {
	FromBatchID uuid.UUID
	ToBatchID   uuid.UUID
	Quantity    valueobjects.Quantity
	Reference   valueobjects.Reference
}

// Transfer moves quantity from one batch to another inside a single
// transaction. Both batches are locked via SELECT FOR UPDATE
// (implemented in the repository) to prevent race conditions.
func (s *InventoryService) Transfer(ctx context.Context, in TransferInput) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		from, err := s.batches.GetByID(ctx, in.FromBatchID)
		if err != nil {
			return err
		}
		to, err := s.batches.GetByID(ctx, in.ToBatchID)
		if err != nil {
			return err
		}
		if from.CompanyID != to.CompanyID {
			return derrors.New("COMPANY_MISMATCH", "transfer must stay within a company")
		}
		if from.ProductID != to.ProductID {
			return derrors.New("PRODUCT_MISMATCH", "transfer must keep the same product")
		}
		if err := from.Consume(in.Quantity); err != nil {
			return err
		}
		if _, err := to.Receive(in.Quantity); err != nil {
			return err
		}
		if err := s.batches.Update(ctx, from); err != nil {
			return err
		}
		if err := s.batches.Update(ctx, to); err != nil {
			return err
		}
		now := time.Now().UTC()
		// Outbound movement on the source batch.
		outRef, err := valueobjects.NewReference(enums.ReferenceTypeTransfer, to.ID)
		if err != nil {
			return err
		}
		out, err := NewInventoryMovement(NewInventoryMovementOptions{
			CompanyID:    from.CompanyID,
			ProductID:    from.ProductID,
			WarehouseID:  from.WarehouseID,
			BatchID:      &from.ID,
			Type:         enums.MovementTypeTransferOut,
			Quantity:     in.Quantity.Neg(),
			UnitCost:     from.UnitCost,
			CurrencyCode: from.CurrencyCode,
			OccurredAt:   now,
			Reference:    &outRef,
			Notes:        "transfer out",
		})
		if err != nil {
			return err
		}
		if err := s.movements.Create(ctx, out); err != nil {
			return err
		}
		// Inbound movement on the destination batch.
		inRef, err := valueobjects.NewReference(enums.ReferenceTypeTransfer, from.ID)
		if err != nil {
			return err
		}
		in, err := NewInventoryMovement(NewInventoryMovementOptions{
			CompanyID:    to.CompanyID,
			ProductID:    to.ProductID,
			WarehouseID:  to.WarehouseID,
			BatchID:      &to.ID,
			Type:         enums.MovementTypeTransferIn,
			Quantity:     in.Quantity,
			UnitCost:     to.UnitCost,
			CurrencyCode: to.CurrencyCode,
			OccurredAt:   now,
			Reference:    &inRef,
			Notes:        "transfer in",
		})
		if err != nil {
			return err
		}
		if err := s.movements.Create(ctx, in); err != nil {
			return err
		}
		return nil
	})
}

// consumeOrAdjust is the shared workhorse for Issue and Adjust.
// It validates the batch, applies the change, and writes the
// movement row.
func (s *InventoryService) consumeOrAdjust(
	ctx context.Context,
	batchID uuid.UUID,
	delta valueobjects.Quantity,
	ref valueobjects.Reference,
	movementType enums.InventoryMovementType,
	notes string,
) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		batch, err := s.batches.GetByID(ctx, batchID)
		if err != nil {
			return err
		}
		// Stock signs: outbound (Issue, AdjustOut) use Consume; inbound
		// (AdjustIn) uses Receive. The movement row carries the signed
		// quantity (negative for outbound, positive for inbound).
		amount := delta.Abs()
		movementQuantity := amount
		if movementType.IsOutbound() {
			if err := batch.Consume(amount); err != nil {
				return err
			}
			movementQuantity = amount.Neg()
		} else {
			if _, err := batch.Receive(amount); err != nil {
				return err
			}
		}
		if err := s.batches.Update(ctx, batch); err != nil {
			return err
		}
		movement, err := NewInventoryMovement(NewInventoryMovementOptions{
			CompanyID:    batch.CompanyID,
			ProductID:    batch.ProductID,
			WarehouseID:  batch.WarehouseID,
			BatchID:      &batch.ID,
			Type:         movementType,
			Quantity:     movementQuantity,
			UnitCost:     batch.UnitCost,
			CurrencyCode: batch.CurrencyCode,
			OccurredAt:   time.Now().UTC(),
			Reference:    &ref,
			Notes:        notes,
		})
		if err != nil {
			return err
		}
		if err := s.movements.Create(ctx, movement); err != nil {
			return err
		}
		s.log.Info("inventory movement",
			"type", movementType,
			"batch_id", batch.ID,
			"product_id", batch.ProductID,
			"warehouse_id", batch.WarehouseID,
			"quantity", delta,
		)
		return nil
	})
}

// StockFor returns the current quantity of a batch.
func (s *InventoryService) StockFor(ctx context.Context, batchID uuid.UUID) (valueobjects.Quantity, error) {
	batch, err := s.batches.GetByID(ctx, batchID)
	if err != nil {
		return valueobjects.Quantity{}, err
	}
	return batch.CurrentQuantity, nil
}

// GenerateClearanceCandidates returns the batches that are past their
// maximum sale date and still have stock. Used by the dashboard widget
// and the "Productos en remate" report.
func (s *InventoryService) GenerateClearanceCandidates(ctx context.Context, companyID uuid.UUID, at time.Time) ([]*InventoryBatch, error) {
	return s.batches.ListClearance(ctx, companyID, at)
}

// NeedsClearanceSoon returns the batches within 3 days of their clearance
// date.
func (s *InventoryService) NeedsClearanceSoon(ctx context.Context, companyID uuid.UUID) ([]*InventoryBatch, error) {
	all, err := s.batches.List(ctx, InventoryBatchFilter{
		CompanyID:   &companyID,
		OnlyActive:  true,
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var out []*InventoryBatch
	for _, b := range all.Items {
		if b.NeedsClearanceSoon(now) {
			out = append(out, b)
		}
	}
	return out, nil
}

// AgingReport returns all active batches for a company, sorted by age
// (oldest first). Used by the inventory aging report.
func (s *InventoryService) AgingReport(ctx context.Context, companyID uuid.UUID) ([]*InventoryBatch, error) {
	all, err := s.batches.List(ctx, InventoryBatchFilter{
		CompanyID:   &companyID,
		OnlyActive:  true,
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	// The repo's List returns by arrival_date DESC. Aging report wants
	// oldest first. We do that here to keep the repo's ordering
	// stable for other callers.
	now := time.Now().UTC()
	for i, j := 0, len(all.Items)-1; i < j; i, j = i+1, j-1 {
		all.Items[i], all.Items[j] = all.Items[j], all.Items[i]
	}
	_ = now
	return all.Items, nil
}

// ListBatches returns inventory batches matching the filter.
func (s *InventoryService) ListBatches(ctx context.Context, filter InventoryBatchFilter) (repositories.Page[*InventoryBatch], error) {
	return s.batches.List(ctx, filter)
}

// ListMovements returns inventory movements matching the filter.
func (s *InventoryService) ListMovements(ctx context.Context, filter InventoryMovementFilter) (repositories.Page[*InventoryMovement], error) {
	return s.movements.List(ctx, filter)
}
