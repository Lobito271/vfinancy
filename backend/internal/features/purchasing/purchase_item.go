package purchasing

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// DefaultUnitCode is used when a line carries no unit.
const DefaultUnitCode = "Unidad"

// PurchaseItem is a single line of a purchase. A line with
// a nil ProductID auto-creates the product from its description on
// order creation.
type PurchaseItem struct {
	ID               uuid.UUID
	PurchaseID  uuid.UUID
	ProductID        *uuid.UUID
	LineNumber       int
	Description      string
	UnitCode         string
	QuantityOrdered  valueobjects.Quantity
	QuantityReceived valueobjects.Quantity
	UnitCostUSD      valueobjects.Money
	LineTotalUSD     valueobjects.Money
	SalePricePen     valueobjects.Money
	CreatedAt        time.Time

	// QuantitySold is a read-only view computed by the repository from
	// the inventory movement ledger of the lots created out of this line.
	QuantitySold valueobjects.Quantity
}

// SoldOut reports whether every unit of the line has been sold.
func (li *PurchaseItem) SoldOut() bool {
	return li.QuantitySold.GreaterOrEqual(li.QuantityOrdered)
}

// PurchaseItemOptions is the input to NewPurchaseItem.
type PurchaseItemOptions struct {
	PurchaseID uuid.UUID
	ProductID       *uuid.UUID
	LineNumber      int
	Description     string
	UnitCode        string
	Quantity        valueobjects.Quantity
	UnitCostUSD     valueobjects.Money
	SalePricePen    valueobjects.Money
}

// NewPurchaseItem validates and constructs a line, computing the
// USD line total.
func NewPurchaseItem(opts PurchaseItemOptions) (*PurchaseItem, error) {
	if !opts.Quantity.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity must be greater than zero"))
	}
	if opts.UnitCostUSD.IsNegative() || opts.SalePricePen.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("line amounts cannot be negative"))
	}
	if opts.Description == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("line description is required"))
	}
	unitCode := opts.UnitCode
	if unitCode == "" {
		unitCode = DefaultUnitCode
	}
	return &PurchaseItem{
		ID:               uuid.New(),
		PurchaseID:  opts.PurchaseID,
		ProductID:        opts.ProductID,
		LineNumber:       opts.LineNumber,
		Description:      opts.Description,
		UnitCode:         unitCode,
		QuantityOrdered:  opts.Quantity,
		QuantityReceived: valueobjects.ZeroQuantity(),
		UnitCostUSD:      opts.UnitCostUSD,
		LineTotalUSD:     opts.UnitCostUSD.MulByDecimal(opts.Quantity.Decimal()).RoundToCurrencyPrecision(),
		SalePricePen:     opts.SalePricePen,
		CreatedAt:        time.Now().UTC(),
	}, nil
}
