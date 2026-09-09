package purchasing

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// DefaultUnitCode is used when a line carries no unit.
const DefaultUnitCode = "Unidad"

// PurchaseOrderItem is a single line of a purchase order. A line with
// a nil ProductID auto-creates the product from its description on
// order creation.
type PurchaseOrderItem struct {
	ID               uuid.UUID
	PurchaseOrderID  uuid.UUID
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
}

// PurchaseOrderItemOptions is the input to NewPurchaseOrderItem.
type PurchaseOrderItemOptions struct {
	PurchaseOrderID uuid.UUID
	ProductID       *uuid.UUID
	LineNumber      int
	Description     string
	UnitCode        string
	Quantity        valueobjects.Quantity
	UnitCostUSD     valueobjects.Money
	SalePricePen    valueobjects.Money
}

// NewPurchaseOrderItem validates and constructs a line, computing the
// USD line total.
func NewPurchaseOrderItem(opts PurchaseOrderItemOptions) (*PurchaseOrderItem, error) {
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
	return &PurchaseOrderItem{
		ID:               uuid.New(),
		PurchaseOrderID:  opts.PurchaseOrderID,
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

// LineRealCostPen is the landed cost in PEN of one unit of the line:
// unit_cost_usd * exchange_rate.
func (li *PurchaseOrderItem) LineRealCostPen(rate valueobjects.ExchangeRate) valueobjects.Money {
	return rate.Convert(li.UnitCostUSD)
}
