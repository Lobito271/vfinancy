package sales

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// SaleItem is one line of a sale. The cost snapshot fixes the unit
// cost at the moment of sale and drives the profit calculation. A
// nil InventoryBatchID means the units were drawn FIFO across several
// lots, so there is no single origin to report.
type SaleItem struct {
	ID               uuid.UUID
	SaleID           uuid.UUID
	ProductID        uuid.UUID
	InventoryBatchID *uuid.UUID
	LineNumber       int
	Quantity         valueobjects.Quantity
	UnitPrice        valueobjects.Money
	LineTotal        valueobjects.Money
	CostSnapshot     valueobjects.Money
	Description      string
	CreatedAt        time.Time

	// OriginLotArrivalDate, OriginPurchaseID and
	// OriginPurchaseNumber are read-only views of the lot the line
	// was sold from, filled by the repository for the sale detail.
	OriginLotArrivalDate *time.Time
	OriginPurchaseID     *uuid.UUID
	OriginPurchaseNumber string
}

// NewSaleItemOptions is the input to NewSaleItem.
type NewSaleItemOptions struct {
	SaleID           uuid.UUID
	ProductID        uuid.UUID
	InventoryBatchID *uuid.UUID
	LineNumber       int
	Quantity         valueobjects.Quantity
	UnitPrice        valueobjects.Money
	CostSnapshot     valueobjects.Money
	Description      string
}

// NewSaleItem validates and constructs a line. The line total is
// computed as unit_price * quantity.
func NewSaleItem(now time.Time, opts NewSaleItemOptions) (*SaleItem, error) {
	if opts.ProductID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("product is required"))
	}
	if !opts.Quantity.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity must be greater than zero"))
	}
	if opts.UnitPrice.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("unit price cannot be negative"))
	}
	if opts.CostSnapshot.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("cost snapshot cannot be negative"))
	}
	return &SaleItem{
		ID:               uuid.New(),
		SaleID:           opts.SaleID,
		ProductID:        opts.ProductID,
		InventoryBatchID: opts.InventoryBatchID,
		LineNumber:       opts.LineNumber,
		Quantity:         opts.Quantity,
		UnitPrice:        opts.UnitPrice,
		LineTotal:        opts.UnitPrice.MulByDecimal(opts.Quantity.Decimal()).RoundToCurrencyPrecision(),
		CostSnapshot:     opts.CostSnapshot,
		Description:      opts.Description,
		CreatedAt:        now,
	}, nil
}
