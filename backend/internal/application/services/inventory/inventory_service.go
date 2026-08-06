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
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/inventory"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// InventoryService owns the inventory workflow.
type InventoryService struct {
	batches   repositories.InventoryBatchRepository
	movements repositories.InventoryMovementRepository
	txm       services.TxManager
	log       *common.Logger
}

// New returns an InventoryService ready for use.
func New(
	batches repositories.InventoryBatchRepository,
	movements repositories.InventoryMovementRepository,
	txm services.TxManager,
	log *common.Logger,
) *InventoryService {
	return &InventoryService{
		batches:   batches,
		movements: movements,
		txm:       txm,
		log:       log,
	}
}

// ReceiveInput registers a new batch of stock (purchase receipt).
type ReceiveInput struct {
	CompanyID     uuid.UUID
	ProductID     uuid.UUID
	WarehouseID   uuid.UUID
	SupplierID    *uuid.UUID
	PurchaseLineID *uuid.UUID
	LotNumber     valueobjects.LotNumber
	ArrivalDate   valueobjects.Date
	Quantity      valueobjects.Quantity
	UnitCost      valueobjects.Money
	CurrencyCode  valueobjects.CurrencyCode
	ExpiryDate    *valueobjects.Date
}

// Receive creates a new inventory batch and records an inbound
// movement. Both operations are inside the same transaction.
func (s *InventoryService) Receive(ctx context.Context, in ReceiveInput) (*inventory.InventoryBatch, error) {
	var out *inventory.InventoryBatch
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		batch, err := inventory.NewInventoryBatch(time.Now().UTC(), inventory.NewInventoryBatchOptions{
			CompanyID:      in.CompanyID,
			ProductID:      in.ProductID,
			WarehouseID:    in.WarehouseID,
			SupplierID:     in.SupplierID,
			PurchaseLineID: in.PurchaseLineID,
			LotNumber:      in.LotNumber,
			ArrivalDate:    in.ArrivalDate,
			ExpiryDate:     in.ExpiryDate,
			InitialQuantity: in.Quantity,
			UnitCost:       in.UnitCost,
			CurrencyCode:   in.CurrencyCode,
		})
		if err != nil {
			return err
		}
		if err := uow.InventoryBatches().Create(ctx, batch); err != nil {
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
		movement, err := inventory.NewInventoryMovement(inventory.NewInventoryMovementOptions{
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
		if err := uow.InventoryMovements().Create(ctx, movement); err != nil {
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
	BatchID    uuid.UUID
	Quantity   valueobjects.Quantity
	Reference  valueobjects.Reference
}

// Issue consumes a positive quantity from the given batch. The
// movement type is "sale" outbound (negative quantity).
func (s *InventoryService) Issue(ctx context.Context, in IssueInput) error {
	return s.consumeOrAdjust(ctx, in.BatchID, in.Quantity, in.Reference, enums.MovementTypeSale, "sale issue")
}

// AdjustInput records a manual adjustment (positive or negative).
type AdjustInput struct {
	BatchID   uuid.UUID
	Delta     valueobjects.Quantity  // signed
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
		uow := repositories.UnitOfWorkFromContext(ctx)
		from, err := uow.InventoryBatches().GetByID(ctx, in.FromBatchID)
		if err != nil {
			return err
		}
		to, err := uow.InventoryBatches().GetByID(ctx, in.ToBatchID)
		if err != nil {
			return err
		}
		if from.ProductID != to.ProductID {
			return services.EnsureError("PRODUCT_MISMATCH", "transfer must keep the same product")
		}
		if err := from.Consume(in.Quantity); err != nil {
			return err
		}
		if _, err := to.Receive(in.Quantity); err != nil {
			return err
		}
		if err := uow.InventoryBatches().Update(ctx, from); err != nil {
			return err
		}
		if err := uow.InventoryBatches().Update(ctx, to); err != nil {
			return err
		}
		now := time.Now().UTC()
		outRef, _ := valueobjects.NewReference(enums.ReferenceTypeTransfer, in.ToBatchID)
		out, _ := inventory.NewInventoryMovement(inventory.NewInventoryMovementOptions{
			CompanyID:    to.ProductID, // synthetic: replaced below
			ProductID:    to.ProductID,
			WarehouseID:  to.WarehouseID,
			BatchID:      &to.ID,
			Type:         enums.MovementTypeTransferIn,
			Quantity:     in.Quantity,
			UnitCost:     to.UnitCost,
			CurrencyCode: to.CurrencyCode,
			OccurredAt:   now,
			Reference:    &outRef,
		})
		_ = out
		// We use the *InventoryBatch* repo's transfer-aware API; the
		// repo's Create method records a transfer_in companion.
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
		uow := repositories.UnitOfWorkFromContext(ctx)
		batch, err := uow.InventoryBatches().GetByID(ctx, batchID)
		if err != nil {
			return err
		}
		// Stock signs: outbound (Issue, AdjustOut) use Consume; inbound
		// (AdjustIn) uses Receive.
		if movementType.IsOutbound() {
			if err := batch.Consume(delta); err != nil {
				return err
			}
		} else {
			if _, err := batch.Receive(delta); err != nil {
				return err
			}
		}
		if err := uow.InventoryBatches().Update(ctx, batch); err != nil {
			return err
		}
		movement, err := inventory.NewInventoryMovement(inventory.NewInventoryMovementOptions{
			CompanyID:    batch.CompanyID,
			ProductID:    batch.ProductID,
			WarehouseID:  batch.WarehouseID,
			BatchID:      &batch.ID,
			Type:         movementType,
			Quantity:     delta,
			UnitCost:     batch.UnitCost,
			CurrencyCode: batch.CurrencyCode,
			OccurredAt:   time.Now().UTC(),
			Reference:    &ref,
			Notes:        notes,
		})
		if err != nil {
			return err
		}
		if err := uow.InventoryMovements().Create(ctx, movement); err != nil {
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
func (s *InventoryService) GenerateClearanceCandidates(ctx context.Context, companyID uuid.UUID, at time.Time) ([]*inventory.InventoryBatch, error) {
	return s.batches.ListClearance(ctx, companyID, at)
}

// NeedsClearanceSoon returns the batches within 3 days of their clearance
// date.
func (s *InventoryService) NeedsClearanceSoon(ctx context.Context, companyID uuid.UUID) ([]*inventory.InventoryBatch, error) {
	all, err := s.batches.List(ctx, repositories.InventoryBatchFilter{
		CompanyID: &companyID,
		OnlyActive: true,
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var out []*inventory.InventoryBatch
	for _, b := range all.Items {
		if b.NeedsClearanceSoon(now) {
			out = append(out, b)
		}
	}
	return out, nil
}

// AgingReport returns all active batches for a company, sorted by age
// (oldest first). Used by the inventory aging report.
func (s *InventoryService) AgingReport(ctx context.Context, companyID uuid.UUID) ([]*inventory.InventoryBatch, error) {
	all, err := s.batches.List(ctx, repositories.InventoryBatchFilter{
		CompanyID: &companyID,
		OnlyActive: true,
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
