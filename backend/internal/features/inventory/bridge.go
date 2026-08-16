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

// WarehouseResolver locates the destination warehouse for automated
// inbound stock movements.
type WarehouseResolver interface {
	// DefaultWarehouseID returns the default active warehouse for a
	// company. An error is returned when no warehouse is configured.
	DefaultWarehouseID(ctx context.Context, companyID uuid.UUID) (uuid.UUID, error)
}

// ProductClassifier tells whether a product is a physical good (stock
// tracked) or a service (no stock).
type ProductClassifier interface {
	IsService(ctx context.Context, productID uuid.UUID) (bool, error)
}

// ReserveForSaleInput is the payload of ReserveForSale.
type ReserveForSaleInput struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	Quantity  valueobjects.Quantity
	SaleID    uuid.UUID
}

// ReserveForSale consumes `in.Quantity` from the active batches of the
// product at the company's default warehouse using FIFO ordering (the
// batch with the earliest arrival date is consumed first). For every
// consumed batch an outbound "sale" movement referencing the sale is
// appended to the ledger. The whole operation runs on a single
// transaction with each batch row locked.
//
// It returns the weighted average unit cost of the consumed units; the
// caller uses it as the sale line's cost snapshot.
func (s *InventoryService) ReserveForSale(ctx context.Context, in ReserveForSaleInput) (valueobjects.Money, error) {
	var weighted valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		warehouseID, err := s.warehouses.DefaultWarehouseID(ctx, in.CompanyID)
		if err != nil {
			return err
		}
		page, err := s.batches.List(ctx, InventoryBatchFilter{
			CompanyID:   &in.CompanyID,
			ProductID:   &in.ProductID,
			WarehouseID: &warehouseID,
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
			if take.GreaterThan(locked.CurrentQuantity) {
				take = locked.CurrentQuantity
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
			mv, err := NewInventoryMovement(NewInventoryMovementOptions{
				CompanyID:    in.CompanyID,
				ProductID:    in.ProductID,
				WarehouseID:  warehouseID,
				BatchID:      &locked.ID,
				Type:         enums.MovementTypeSale,
				Quantity:     take.Neg(),
				UnitCost:     locked.UnitCost,
				CurrencyCode: locked.CurrencyCode,
				OccurredAt:   now,
				Reference:    &ref,
				Notes:        "sale issue (FIFO)",
			})
			if err != nil {
				return err
			}
			if err := s.movements.Create(ctx, mv); err != nil {
				return err
			}
			totalCost = totalCost.Add(locked.UnitCost.Decimal().Mul(take.Decimal()))
			totalQty = totalQty.Add(take.Decimal())
			remaining = remaining.Sub(take)
		}
		if !remaining.IsZero() {
			return derrors.Wrap(derrors.ErrInsufficientStock, errField("no enough stock for product "+in.ProductID.String()))
		}
		if totalQty.IsZero() {
			return derrors.Wrap(derrors.ErrInsufficientStock, errField("no stock available for product "+in.ProductID.String()))
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
func (s *InventoryService) ReturnVoidedSale(ctx context.Context, companyID, saleID uuid.UUID) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		page, err := s.movements.List(ctx, InventoryMovementFilter{
			CompanyID:     &companyID,
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
			if m.Type != enums.MovementTypeSale || m.BatchID == nil {
				continue
			}
			locked, err := s.batches.GetByIDForUpdate(ctx, *m.BatchID)
			if err != nil {
				return err
			}
			restore := m.Quantity.Abs()
			if _, err := locked.Receive(restore); err != nil {
				return err
			}
			if err := s.batches.Update(ctx, locked); err != nil {
				return err
			}
			mv, err := NewInventoryMovement(NewInventoryMovementOptions{
				CompanyID:    m.CompanyID,
				ProductID:    m.ProductID,
				WarehouseID:  m.WarehouseID,
				BatchID:      m.BatchID,
				Type:         enums.MovementTypeVoidSale,
				Quantity:     restore,
				UnitCost:     m.UnitCost,
				CurrencyCode: m.CurrencyCode,
				OccurredAt:   now,
				Reference:    &ref,
				Notes:        "sale voided, stock returned",
			})
			if err != nil {
				return err
			}
			if err := s.movements.Create(ctx, mv); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReceiveFromPurchaseInput is the payload of ReceiveFromPurchase.
type ReceiveFromPurchaseInput struct {
	CompanyID      uuid.UUID
	SupplierID     *uuid.UUID
	ProductID      uuid.UUID
	PurchaseLineID uuid.UUID
	LotNumber      valueobjects.LotNumber
	ArrivalDate    valueobjects.Date
	Quantity       valueobjects.Quantity
	UnitCost       valueobjects.Money
	CurrencyCode   valueobjects.CurrencyCode
	ExpiryDate     *valueobjects.Date
}

// ReceiveFromPurchase registers the goods received from a purchase
// order line as a new inventory batch with an inbound
// "purchase_receipt" movement. The operation is idempotent per line: a
// line that already has a batch is skipped, so Create / Approve /
// MarkAsReceived can all trigger the receipt safely.
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
		warehouseID, err := s.warehouses.DefaultWarehouseID(ctx, in.CompanyID)
		if err != nil {
			return err
		}
		batch, err := NewInventoryBatch(time.Now().UTC(), NewInventoryBatchOptions{
			CompanyID:       in.CompanyID,
			ProductID:       in.ProductID,
			WarehouseID:     warehouseID,
			SupplierID:      in.SupplierID,
			PurchaseLineID:  &in.PurchaseLineID,
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
		ref, err := valueobjects.NewReference(enums.ReferenceTypePurchase, in.PurchaseLineID)
		if err != nil {
			return err
		}
		mv, err := NewInventoryMovement(NewInventoryMovementOptions{
			CompanyID:    in.CompanyID,
			ProductID:    in.ProductID,
			WarehouseID:  warehouseID,
			BatchID:      &batch.ID,
			Type:         enums.MovementTypePurchaseReceipt,
			Quantity:     in.Quantity,
			UnitCost:     in.UnitCost,
			CurrencyCode: in.CurrencyCode,
			OccurredAt:   time.Now().UTC(),
			Reference:    &ref,
			Notes:        "purchase receipt",
		})
		if err != nil {
			return err
		}
		if err := s.movements.Create(ctx, mv); err != nil {
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
			"quantity", out.InitialQuantity,
		)
	}
	return out, nil
}

// VoidPurchaseReceipt deducts the remaining stock of every batch
// created from the given purchase order lines and appends an outbound
// "void_purchase" movement. Batches already consumed (zero remaining)
// are left untouched; batch rows are kept for audit and to preserve
// historical sale allocations.
func (s *InventoryService) VoidPurchaseReceipt(ctx context.Context, companyID uuid.UUID, purchaseLineIDs []uuid.UUID) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		now := time.Now().UTC()
		for _, lineID := range purchaseLineIDs {
			page, err := s.batches.List(ctx, InventoryBatchFilter{
				CompanyID:      &companyID,
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
				remaining := locked.CurrentQuantity
				if !remaining.IsPositive() {
					continue
				}
				if err := locked.Consume(remaining); err != nil {
					return err
				}
				if err := s.batches.Update(ctx, locked); err != nil {
					return err
				}
				mv, err := NewInventoryMovement(NewInventoryMovementOptions{
					CompanyID:    locked.CompanyID,
					ProductID:    locked.ProductID,
					WarehouseID:  locked.WarehouseID,
					BatchID:      &locked.ID,
					Type:         enums.MovementTypeVoidPurchase,
					Quantity:     remaining.Neg(),
					UnitCost:     locked.UnitCost,
					CurrencyCode: locked.CurrencyCode,
					OccurredAt:   now,
					Reference:    &ref,
					Notes:        "purchase voided, stock deducted",
				})
				if err != nil {
					return err
				}
				if err := s.movements.Create(ctx, mv); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
