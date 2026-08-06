package sales

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// SaleItem is a single line in a sale. It carries the product, the
// quantity, the unit price, any line-level discount, and the tax
// (snapshot). The cost snapshot is used for profit calculation.
type SaleItem struct {
	ID            uuid.UUID
	SaleID        uuid.UUID
	ProductID     uuid.UUID
	LineNumber    int
	Quantity      valueobjects.Quantity
	UnitPrice     valueobjects.Money
	DiscountPercent valueobjects.Percentage
	DiscountAmount valueobjects.Money
	TaxRate       valueobjects.Percentage
	TaxAmount     valueobjects.Money
	CostSnapshot  valueobjects.Money
	Description   string
	CreatedAt     time.Time
}

// NewSaleItemOptions is the input to NewSaleItem.
type NewSaleItemOptions struct {
	ProductID       uuid.UUID
	LineNumber      int
	Quantity        valueobjects.Quantity
	UnitPrice       valueobjects.Money
	DiscountPercent valueobjects.Percentage
	DiscountAmount  valueobjects.Money
	TaxRate         valueobjects.Percentage
	TaxAmount       valueobjects.Money
	CostSnapshot    valueobjects.Money
	Description     string
}

// NewSaleItem validates and constructs a SaleItem.
func NewSaleItem(opts NewSaleItemOptions) (*SaleItem, error) {
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
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("discount amount cannot be negative"))
	}
	if opts.TaxAmount.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("tax amount cannot be negative"))
	}
	if opts.CostSnapshot.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("cost snapshot cannot be negative"))
	}
	return &SaleItem{
		ID:              uuid.New(),
		ProductID:       opts.ProductID,
		LineNumber:      opts.LineNumber,
		Quantity:        opts.Quantity,
		UnitPrice:       opts.UnitPrice,
		DiscountPercent: opts.DiscountPercent,
		DiscountAmount:  opts.DiscountAmount,
		TaxRate:         opts.TaxRate,
		TaxAmount:       opts.TaxAmount,
		CostSnapshot:    opts.CostSnapshot,
		Description:     opts.Description,
	}, nil
}

// LineSubtotal is the gross line amount: unit_price * quantity.
func (li *SaleItem) LineSubtotal() valueobjects.Money {
	qDec := li.Quantity.Decimal()
	return MoneyFromDecimal(li.UnitPrice.Decimal().Mul(qDec))
}

// LineTotal is line_subtotal - discount + tax.
func (li *SaleItem) LineTotal() valueobjects.Money {
	sub := li.LineSubtotal()
	afterDiscount := sub.Sub(li.DiscountAmount)
	return afterDiscount.Add(li.TaxAmount)
}

// LineCost is cost_snapshot * quantity.
func (li *SaleItem) LineCost() valueobjects.Money {
	qDec := li.Quantity.Decimal()
	return MoneyFromDecimal(li.CostSnapshot.Decimal().Mul(qDec))
}

// LineProfit is line_total - tax - cost (the realized gross margin on
// the line).
func (li *SaleItem) LineProfit() valueobjects.Money {
	return li.LineTotal().Sub(li.TaxAmount).Sub(li.LineCost())
}

// ChangeQuantity updates the quantity. The new line totals are not
// recomputed here — the application layer (or the sale aggregate) is
// responsible for refreshing the line_total and tax_amount when the
// quantity changes. The sale's recalculate() method does that.
func (li *SaleItem) ChangeQuantity(q valueobjects.Quantity) error {
	if !q.IsPositive() {
		return derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity must be positive"))
	}
	li.Quantity = q
	return nil
}

// ChangeUnitPrice updates the unit price. Negative rejected.
func (li *SaleItem) ChangeUnitPrice(p valueobjects.Money) error {
	if p.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("unit price cannot be negative"))
	}
	li.UnitPrice = p
	return nil
}

// ChangeDiscountAmount updates the line-level discount.
func (li *SaleItem) ChangeDiscountAmount(amount valueobjects.Money) error {
	if amount.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("discount cannot be negative"))
	}
	li.DiscountAmount = amount
	return nil
}

// MoneyFromDecimal is a tiny helper that wraps decimal.NewFromBigInt-or-
// like construction. Lives in this file to avoid an import cycle with
// the valueobjects package.
func MoneyFromDecimal(d decimal.Decimal) valueobjects.Money {
	m, _ := valueobjects.MoneyFromDecimal(d)
	return m
}
