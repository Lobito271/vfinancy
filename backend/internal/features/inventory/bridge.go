package inventory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// ReserveForSaleInput is the payload of ReserveForSale.
type ReserveForSaleInput struct {
	ProductID uuid.UUID
	Quantity  valueobjects.Quantity
	SaleID    uuid.UUID
}

// ReserveForSale consumes `in.Quantity` from the active batches of the
// product using FIFO ordering (the batch with the earliest arrival
// date is consumed first). For every consumed batch an outbound "sale"
// movement referencing the sale is appended to the ledger. The whole
// operation runs on a single transaction with each batch row locked.
//
// It returns the weighted average unit cost of the consumed units; the
// caller uses it as the sale line's cost snapshot. When the stock is
// short the transaction is rolled back and ErrInsufficientStock is
// returned.
func (s *InventoryService) ReserveForSale(ctx context.Context, in ReserveForSaleInput) (valueobjects.Money, error) {
	var weighted valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		page, err := s.batches.List(ctx, InventoryBatchFilter{
			ProductID:   &in.ProductID,
			OnlyActive:  true,
			PageRequest: repositories.PageRequest{Limit: 1000},
		})
		if err != nil {
			return err
		}
		// List returns arrival_date DESC; FIFO needs oldest first.
		batches := make([]*InventoryBatch, 0, len(page.Items))
		for i := len(page.Items) - 1; i >= 0; i-- {
			batches = append(batches, page.Items[i])
		}

		ref, err := valueobjects.NewReference(enums.ReferenceTypeSale, in.SaleID)
		if err != nil {
			return err
		}
		remaining := in.Quantity
		var totalCost, totalQty decimal.Decimal
		now := time.Now().UTC()
		for _, b := range batches {
			if remaining.IsZero() {
				break
			}
			locked, err := s.batches.GetByIDForUpdate(ctx, b.ID)
			if err != nil {
				return err
			}
			take := remaining
			if take.GreaterThan(locked.Quantity) {
				take = locked.Quantity
			}
			if take.IsZero() {
				continue
			}
			if err := locked.Consume(take); err != nil {
				return err
			}
			if err := s.batches.Update(ctx, locked); err != nil {
				return err
			}
			movement, err := NewInventoryMovement(now, NewInventoryMovementOptions{
				BatchID:      locked.ID,
				ProductID:    in.ProductID,
				Type:         enums.MovementTypeSale,
				Quantity:     take.Neg(),
				BalanceAfter: locked.Quantity,
				UnitCost:     locked.UnitCost,
				Reference:    &ref,
				Notes:        "sale issue (FIFO)",
			})
			if err != nil {
				return err
			}
			if err := s.movements.Create(ctx, movement); err != nil {
				return err
			}
			totalCost = totalCost.Add(locked.UnitCost.Decimal().Mul(take.Decimal()))
			totalQty = totalQty.Add(take.Decimal())
			remaining = remaining.Sub(take)
		}
		if !remaining.IsZero() || totalQty.IsZero() {
			return derrors.Wrap(derrors.ErrInsufficientStock, errField("not enough stock for product "+in.ProductID.String()))
		}
		weighted, err = valueobjects.MoneyFromDecimal(totalCost.Div(totalQty))
		return err
	})
	if err != nil {
		return valueobjects.Zero(), err
	}
	s.log.Info("inventory reserved for sale",
		"sale_id", in.SaleID,
		"product_id", in.ProductID,
		"quantity", in.Quantity,
		"weighted_cost", weighted,
	)
	return weighted, nil
}

// ReturnVoidedSale restores the stock consumed by a cancelled sale. It
// finds every "sale" movement referencing the sale and writes a
// compensating inbound "void_sale" movement back to the same batch,
// inside a single transaction.
func (s *InventoryService) ReturnVoidedSale(ctx context.Context, saleID uuid.UUID) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		page, err := s.movements.List(ctx, InventoryMovementFilter{
			ReferenceType: enums.ReferenceTypeSale.String(),
			ReferenceID:   &saleID,
			PageRequest:   repositories.PageRequest{Limit: 1000},
		})
		if err != nil {
			return err
		}
		ref, err := valueobjects.NewReference(enums.ReferenceTypeSale, saleID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, m := range page.Items {
			if m.Type != enums.MovementTypeSale {
				continue
			}
			locked, err := s.batches.GetByIDForUpdate(ctx, m.BatchID)
			if err != nil {
				return err
			}
			restore := m.QuantityDelta.Abs()
			newQuantity, err := locked.Receive(restore)
			if err != nil {
				return err
			}
			if err := s.batches.Update(ctx, locked); err != nil {
				return err
			}
			movement, err := NewInventoryMovement(now, NewInventoryMovementOptions{
				BatchID:      m.BatchID,
				ProductID:    m.ProductID,
				Type:         enums.MovementTypeVoidSale,
				Quantity:     restore,
				BalanceAfter: newQuantity,
				UnitCost:     m.UnitCost,
				Reference:    &ref,
				Notes:        "sale voided, stock returned",
			})
			if err != nil {
				return err
			}
			if err := s.movements.Create(ctx, movement); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReceiveFromPurchaseInput is the payload of ReceiveFromPurchase.
type ReceiveFromPurchaseInput struct {
	ProductID      uuid.UUID
	PurchaseLineID uuid.UUID
	ArrivalDate    valueobjects.Date
	Quantity       valueobjects.Quantity
	UnitCost       valueobjects.Money
	ExchangeRate   valueobjects.ExchangeRate
}

// ReceiveFromPurchase registers the goods received from a purchase
// order line as a new inventory batch with an inbound
// "purchase_receipt" movement. The operation is idempotent per line: a
// line that already has a batch is skipped, so Create / Approve /
// MarkAsReceived can all trigger the receipt safely. The arrival date
// cannot be in the future.
func (s *InventoryService) ReceiveFromPurchase(ctx context.Context, in ReceiveFromPurchaseInput) (*InventoryBatch, error) {
	var out *InventoryBatch
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		exists, err := s.batches.ExistsByPurchaseLineID(ctx, in.PurchaseLineID)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		batch, err := NewInventoryBatch(time.Now().UTC(), NewInventoryBatchOptions{
			ProductID:           in.ProductID,
			PurchaseOrderItemID: &in.PurchaseLineID,
			ArrivalDate:         in.ArrivalDate,
			InitialQuantity:     in.Quantity,
			UnitCost:            in.UnitCost,
			ExchangeRate:        in.ExchangeRate,
		})
		if err != nil {
			return err
		}
		if err := s.batches.Create(ctx, batch); err != nil {
			return err
		}
		ref, err := valueobjects.NewReference(enums.ReferenceTypePurchase, in.PurchaseLineID)
		if err != nil {
			return err
		}
		movement, err := NewInventoryMovement(time.Now().UTC(), NewInventoryMovementOptions{
			BatchID:      batch.ID,
			ProductID:    batch.ProductID,
			Type:         enums.MovementTypePurchaseReceipt,
			Quantity:     in.Quantity,
			BalanceAfter: batch.Quantity,
			UnitCost:     in.UnitCost,
			Reference:    &ref,
			Notes:        "purchase receipt",
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
	if out != nil {
		s.log.Info("inventory received from purchase",
			"batch_id", out.ID,
			"purchase_line_id", in.PurchaseLineID,
			"product_id", in.ProductID,
			"quantity", out.Quantity,
		)
	}
	return out, nil
}

// VoidPurchaseReceipt deducts the remaining stock of every batch
// created from the given purchase order lines and appends an outbound
// "void_purchase" movement. Batches already consumed (zero remaining)
// are left untouched; batch rows are kept for audit and to preserve
// historical sale allocations.
func (s *InventoryService) VoidPurchaseReceipt(ctx context.Context, purchaseLineIDs []uuid.UUID) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		now := time.Now().UTC()
		for _, lineID := range purchaseLineIDs {
			page, err := s.batches.List(ctx, InventoryBatchFilter{
				PurchaseLineID: &lineID,
				OnlyActive:     true,
				PageRequest:    repositories.PageRequest{Limit: 100},
			})
			if err != nil {
				return err
			}
			ref, err := valueobjects.NewReference(enums.ReferenceTypePurchase, lineID)
			if err != nil {
				return err
			}
			for _, b := range page.Items {
				locked, err := s.batches.GetByIDForUpdate(ctx, b.ID)
				if err != nil {
					return err
				}
				remaining := locked.Quantity
				if !remaining.IsPositive() {
					continue
				}
				if err := locked.Consume(remaining); err != nil {
					return err
				}
				if err := s.batches.Update(ctx, locked); err != nil {
					return err
				}
				movement, err := NewInventoryMovement(now, NewInventoryMovementOptions{
					BatchID:      locked.ID,
					ProductID:    locked.ProductID,
					Type:         enums.MovementTypeVoidPurchase,
					Quantity:     remaining.Neg(),
					BalanceAfter: locked.Quantity,
					UnitCost:     locked.UnitCost,
					Reference:    &ref,
					Notes:        "purchase voided, stock deducted",
				})
				if err != nil {
					return err
				}
				if err := s.movements.Create(ctx, movement); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
