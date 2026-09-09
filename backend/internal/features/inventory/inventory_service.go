// Package inventory implements the business logic for the inventory
// aggregate: stock receipts, adjustments, voids and the clearance rule.
//
// Stock is NEVER stored on the product. Each InventoryBatch carries a
// denormalized quantity for fast lookup, but every change is
// accompanied by an InventoryMovement row and is always recomputable
// from that append-only ledger.
//
// Clearance rule: any batch with arrival_date + ClearanceDays in the
// past and quantity > 0 is "on clearance". The threshold is honored
// from the user-configured setting via SetClearanceSettings, falling
// back to the default 25-day rule.
package inventory

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// InventoryService owns the inventory slice.
type InventoryService struct {
	batches           InventoryBatchRepository
	movements         InventoryMovementRepository
	txm               repositories.TransactionManager
	log               *logger.Logger
	clearanceSettings func(context.Context) (int, int)
}

// New returns an InventoryService ready for use.
func New(batches InventoryBatchRepository, movements InventoryMovementRepository, txm repositories.TransactionManager, log *logger.Logger) *InventoryService {
	return &InventoryService{
		batches:   batches,
		movements: movements,
		txm:       txm,
		log:       log,
	}
}

// SetClearanceSettings installs the provider of the user-configured
// clearance settings (days, warning days).
func (s *InventoryService) SetClearanceSettings(provider func(context.Context) (int, int)) {
	s.clearanceSettings = provider
}

func (s *InventoryService) clearancePolicy(ctx context.Context) (int, int) {
	if s.clearanceSettings != nil {
		if days, warning := s.clearanceSettings(ctx); days > 0 && warning >= 0 {
			return days, warning
		}
	}
	return ClearanceDays, 3
}

// ClearanceDaysFor returns the effective maximum-stay threshold,
// honoring the user-configured value with a fallback to the default
// clearance period.
func (s *InventoryService) ClearanceDaysFor(ctx context.Context) int {
	days, _ := s.clearancePolicy(ctx)
	return days
}

// ReceiveInput is the payload of Receive: a manual stock intake.
type ReceiveInput struct {
	ProductID   uuid.UUID
	Quantity    valueobjects.Quantity
	ArrivalDate valueobjects.Date
	UnitCost    valueobjects.Money
	Notes       string
}

// Receive creates a new inventory batch and records the inbound
// "adjustment_in" movement inside the same transaction. The quantity
// must be positive and the arrival date cannot be in the future.
func (s *InventoryService) Receive(ctx context.Context, in ReceiveInput) (*InventoryBatch, error) {
	var out *InventoryBatch
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		batch, err := NewInventoryBatch(time.Now().UTC(), NewInventoryBatchOptions{
			ProductID:       in.ProductID,
			ArrivalDate:     in.ArrivalDate,
			InitialQuantity: in.Quantity,
			UnitCost:        in.UnitCost,
			ExchangeRate:    valueobjects.One(),
		})
		if err != nil {
			return err
		}
		if err := s.batches.Create(ctx, batch); err != nil {
			return err
		}
		movement, err := NewInventoryMovement(time.Now().UTC(), NewInventoryMovementOptions{
			BatchID:      batch.ID,
			ProductID:    batch.ProductID,
			Type:         enums.MovementTypeAdjustmentIn,
			Quantity:     in.Quantity,
			BalanceAfter: batch.Quantity,
			UnitCost:     in.UnitCost,
			Notes:        in.Notes,
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
		"quantity", out.Quantity,
	)
	return out, nil
}

// AdjustInput is the payload of Adjust: an absolute stock correction.
type AdjustInput struct {
	BatchID     uuid.UUID
	NewQuantity valueobjects.Quantity
	Notes       string
}

// Adjust sets the batch quantity to an absolute target. The delta
// (target minus current quantity) is applied as a single
// "adjustment_in" or "adjustment_out" movement; when the target equals
// the current quantity the call is a no-op. The target must be
// positive.
func (s *InventoryService) Adjust(ctx context.Context, in AdjustInput) error {
	if !in.NewQuantity.IsPositive() {
		return derrors.Wrap(derrors.ErrNegativeQuantity, errField("new quantity must be positive"))
	}
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		batch, err := s.batches.GetByIDForUpdate(ctx, in.BatchID)
		if err != nil {
			return err
		}
		delta := in.NewQuantity.Sub(batch.Quantity)
		if delta.IsZero() {
			return nil
		}
		if delta.IsNegative() {
			if err := batch.Consume(delta.Abs()); err != nil {
				return err
			}
		} else if _, err := batch.Receive(delta); err != nil {
			return err
		}
		batch.Touch()
		if err := s.batches.Update(ctx, batch); err != nil {
			return err
		}
		movement, err := NewInventoryMovement(time.Now().UTC(), NewInventoryMovementOptions{
			BatchID:      batch.ID,
			ProductID:    batch.ProductID,
			Type:         movementTypeForDelta(delta),
			Quantity:     delta,
			BalanceAfter: batch.Quantity,
			UnitCost:     batch.UnitCost,
			Notes:        in.Notes,
		})
		if err != nil {
			return err
		}
		if err := s.movements.Create(ctx, movement); err != nil {
			return err
		}
		s.log.Info("inventory adjusted",
			"batch_id", batch.ID,
			"product_id", batch.ProductID,
			"new_quantity", batch.Quantity,
		)
		return nil
	})
}

// movementTypeForDelta maps a signed adjustment delta to its movement
// type.
func movementTypeForDelta(delta valueobjects.Quantity) enums.InventoryMovementType {
	if delta.IsNegative() {
		return enums.MovementTypeAdjustmentOut
	}
	return enums.MovementTypeAdjustmentIn
}

// VoidInput is the payload of Void.
type VoidInput struct {
	BatchID uuid.UUID
	Reason  string
}

// Void cancels a mistaken stock receipt: the remaining quantity is
// zeroed, the status flips to "voided" and a compensating outbound
// "adjustment_out" movement is appended to the ledger. The batch row
// is kept for audit. Already-voided batches are rejected. Runs inside
// a single transaction with the batch row locked (SELECT ... FOR
// UPDATE).
func (s *InventoryService) Void(ctx context.Context, in VoidInput) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		batch, err := s.batches.GetByIDForUpdate(ctx, in.BatchID)
		if err != nil {
			return err
		}
		if batch.Status == enums.BatchStatusVoided {
			return ErrBatchAlreadyVoided
		}
		remaining := batch.Quantity
		batch.Quantity = valueobjects.ZeroQuantity()
		batch.Status = enums.BatchStatusVoided
		batch.Touch()
		if err := s.batches.Update(ctx, batch); err != nil {
			return err
		}
		if remaining.IsPositive() {
			notes := strings.TrimSpace(in.Reason)
			if notes == "" {
				notes = "batch voided"
			}
			movement, err := NewInventoryMovement(time.Now().UTC(), NewInventoryMovementOptions{
				BatchID:      batch.ID,
				ProductID:    batch.ProductID,
				Type:         enums.MovementTypeAdjustmentOut,
				Quantity:     remaining.Neg(),
				BalanceAfter: valueobjects.ZeroQuantity(),
				UnitCost:     batch.UnitCost,
				Notes:        notes,
			})
			if err != nil {
				return err
			}
			if err := s.movements.Create(ctx, movement); err != nil {
				return err
			}
		}
		s.log.Info("inventory batch voided",
			"batch_id", batch.ID,
			"product_id", batch.ProductID,
			"quantity", remaining,
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
	return batch.Quantity, nil
}

// StockForProduct returns the total available quantity of a product
// across its active batches (main warehouse).
func (s *InventoryService) StockForProduct(ctx context.Context, productID uuid.UUID) (valueobjects.Quantity, error) {
	page, err := s.batches.List(ctx, InventoryBatchFilter{
		ProductID:   &productID,
		OnlyActive:  true,
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return valueobjects.Quantity{}, err
	}
	total := valueobjects.ZeroQuantity()
	for _, b := range page.Items {
		total = total.Add(b.Quantity)
	}
	return total, nil
}

// ListBatches returns inventory batches matching the filter.
func (s *InventoryService) ListBatches(ctx context.Context, filter InventoryBatchFilter) (repositories.Page[*InventoryBatch], error) {
	return s.batches.List(ctx, filter)
}

// ListMovements returns inventory movements matching the filter.
func (s *InventoryService) ListMovements(ctx context.Context, filter InventoryMovementFilter) (repositories.Page[*InventoryMovement], error) {
	return s.movements.List(ctx, filter)
}

// GenerateClearanceCandidates returns the active batches that are past
// their clearance date and still hold stock. Used by the dashboard
// widget and the notifications worker.
func (s *InventoryService) GenerateClearanceCandidates(ctx context.Context, at time.Time) ([]*InventoryBatch, error) {
	days, _ := s.clearancePolicy(ctx)
	out := []*InventoryBatch{}
	for offset := 0; ; {
		page, err := s.batches.List(ctx, InventoryBatchFilter{
			OnlyActive:  true,
			PageRequest: repositories.PageRequest{Limit: 200, Offset: offset},
		})
		if err != nil {
			return nil, err
		}
		for _, batch := range page.Items {
			if batch.IsClearanceOn(days, at) {
				out = append(out, batch)
			}
		}
		if !page.HasMore() {
			break
		}
		offset += len(page.Items)
	}
	return out, nil
}

// RefreshClearanceFlags reconciles the persisted is_clearance column
// with the current clearance rule. The column is normally refreshed
// only here (periodically), so without this pass newly expired batches
// would stay invisible to the is_clearance reads (dashboard KPI,
// OnlyClearance filter) until the next run. Returns how many rows
// changed.
func (s *InventoryService) RefreshClearanceFlags(ctx context.Context, at time.Time) (int, error) {
	days, _ := s.clearancePolicy(ctx)
	return s.batches.RefreshClearanceFlags(ctx, at, days)
}

// NeedsClearanceSoon returns the active stocked batches within the
// warning window of their clearance date (not yet on clearance).
func (s *InventoryService) NeedsClearanceSoon(ctx context.Context, at time.Time) ([]*InventoryBatch, error) {
	days, warning := s.clearancePolicy(ctx)
	today := valueobjects.NewDateFromTime(at)
	out := []*InventoryBatch{}
	for offset := 0; ; {
		page, err := s.batches.List(ctx, InventoryBatchFilter{
			OnlyActive:  true,
			PageRequest: repositories.PageRequest{Limit: 200, Offset: offset},
		})
		if err != nil {
			return nil, err
		}
		for _, batch := range page.Items {
			until := int(batch.MaxSaleDate(days).Sub(today).Hours() / 24)
			if !batch.Quantity.IsZero() && until <= warning && !batch.IsClearanceOn(days, at) {
				out = append(out, batch)
			}
		}
		if !page.HasMore() {
			break
		}
		offset += len(page.Items)
	}
	return out, nil
}

// AgingReport returns the active stocked batches that have already
// arrived as of `at`, oldest first. Used by the inventory aging
// report.
func (s *InventoryService) AgingReport(ctx context.Context, at time.Time) ([]*InventoryBatch, error) {
	today := valueobjects.NewDateFromTime(at)
	page, err := s.batches.List(ctx, InventoryBatchFilter{
		OnlyActive:  true,
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	out := make([]*InventoryBatch, 0, len(page.Items))
	for _, batch := range page.Items {
		if !batch.ArrivalDate.After(today) {
			out = append(out, batch)
		}
	}
	// The repo's List returns by arrival_date DESC. The aging report
	// wants oldest first; we reverse here to keep the repo's ordering
	// stable for other callers.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
