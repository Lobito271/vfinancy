package purchasing

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// PurchaseOrderItem is a single line in a purchase order. Mirrors
// SaleItem in structure.
type PurchaseOrderItem struct {
	ID              uuid.UUID
	PurchaseOrderID uuid.UUID
	ProductID       uuid.UUID
	LineNumber      int
	Quantity        valueobjects.Quantity
	UnitPrice       valueobjects.Money
	DiscountPercent valueobjects.Percentage
	DiscountAmount  valueobjects.Money
	TaxRate         valueobjects.Percentage
	TaxAmount       valueobjects.Money
	Description     string
	CreatedAt       time.Time
}

// NewPurchaseOrderItemOptions is the input to NewPurchaseOrderItem.
type NewPurchaseOrderItemOptions struct {
	ProductID       uuid.UUID
	LineNumber      int
	Quantity        valueobjects.Quantity
	UnitPrice       valueobjects.Money
	DiscountPercent valueobjects.Percentage
	DiscountAmount  valueobjects.Money
	TaxRate         valueobjects.Percentage
	TaxAmount       valueobjects.Money
	Description     string
}

// NewPurchaseOrderItem validates and constructs a line.
func NewPurchaseOrderItem(opts NewPurchaseOrderItemOptions) (*PurchaseOrderItem, error) {
	if opts.ProductID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("product id is required"))
	}
	if !opts.Quantity.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity must be positive"))
	}
	if opts.UnitPrice.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("unit price cannot be negative"))
	}
	if opts.DiscountAmount.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("discount cannot be negative"))
	}
	if opts.TaxAmount.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("tax cannot be negative"))
	}
	return &PurchaseOrderItem{
		ID:              uuid.New(),
		ProductID:       opts.ProductID,
		LineNumber:      opts.LineNumber,
		Quantity:        opts.Quantity,
		UnitPrice:       opts.UnitPrice,
		DiscountPercent: opts.DiscountPercent,
		DiscountAmount:  opts.DiscountAmount,
		TaxRate:         opts.TaxRate,
		TaxAmount:       opts.TaxAmount,
		Description:     opts.Description,
	}, nil
}

// LineSubtotal is unit_price * quantity.
func (li *PurchaseOrderItem) LineSubtotal() valueobjects.Money {
	qDec := li.Quantity.Decimal()
	m, _ := valueobjects.MoneyFromDecimal(li.UnitPrice.Decimal().Mul(qDec))
	return m
}

// LineTotal is line_subtotal - discount + tax.
func (li *PurchaseOrderItem) LineTotal() valueobjects.Money {
	sub := li.LineSubtotal()
	return sub.Sub(li.DiscountAmount).Add(li.TaxAmount)
}

// ChangeQuantity updates the quantity.
func (li *PurchaseOrderItem) ChangeQuantity(q valueobjects.Quantity) error {
	if !q.IsPositive() {
		return derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity must be positive"))
	}
	li.Quantity = q
	return nil
}

// silence unused import warning for decimal in this file.
var _ = decimal.NewFromInt
