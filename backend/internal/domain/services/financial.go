// Package services contains stateless domain services. These are
// functions that operate on entities from multiple bounded contexts
// or that perform a calculation that does not naturally belong to any
// single entity.
//
// Services do not hold state and do not perform I/O. They are called by
// the application layer to centralize business math.
package services

import (
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/valueobjects"
)

// ProfitCalculator is a stateless helper for margin and profit math.
type ProfitCalculator struct{}

// NewProfitCalculator returns a calculator. Stateless, so a single
// instance can be reused everywhere.
func NewProfitCalculator() *ProfitCalculator { return &ProfitCalculator{} }

// GrossMargin returns salePrice - cost.
func (ProfitCalculator) GrossMargin(salePrice, cost valueobjects.Money) valueobjects.Money {
	return salePrice.Sub(cost)
}

// GrossMarginPercent returns the margin as a percentage of salePrice.
// Zero sale price is reported as 0% to avoid division by zero.
func (ProfitCalculator) GrossMarginPercent(salePrice, cost valueobjects.Money) float64 {
	if salePrice.IsZero() {
		return 0
	}
	margin := salePrice.Decimal().Sub(cost.Decimal()).
		Div(salePrice.Decimal()).
		Mul(decimal.NewFromInt(100))
	f, _ := margin.Float64()
	return f
}

// TaxCalculator applies tax rates to a base amount, both exclusive
// (tax added on top) and inclusive (tax already inside the total).
type TaxCalculator struct{}

// NewTaxCalculator returns a calculator.
func NewTaxCalculator() *TaxCalculator { return &TaxCalculator{} }

// TaxExclusive calculates the tax for a base that does NOT include tax.
//   tax        = base * rate / 100
//   total      = base + tax
func (TaxCalculator) TaxExclusive(base valueobjects.Money, rate valueobjects.Percentage) (tax, total valueobjects.Money) {
	tax = base.MulPercent(rate)
	total = base.Add(tax)
	return
}

// TaxInclusive extracts the tax from a total that already includes tax.
//   tax   = total * rate / (100 + rate)
//   base  = total - tax
func (TaxCalculator) TaxInclusive(total valueobjects.Money, rate valueobjects.Percentage) (tax, base valueobjects.Money) {
	hundred := decimal.NewFromInt(100)
	r := rate.Decimal()
	taxDec := total.Decimal().Mul(r).Div(hundred.Add(r))
	t, _ := valueobjects.MoneyFromDecimal(taxDec)
	tax = t.RoundToCurrencyPrecision()
	base = total.Sub(tax)
	return
}

// ApplyDiscountExclusive applies a percentage discount BEFORE tax.
//   discounted = base * (1 - rate/100)
//   tax        = discounted * taxRate / 100
//   total      = discounted + tax
func (TaxCalculator) ApplyDiscountExclusive(base valueobjects.Money, discount, taxRate valueobjects.Percentage) (discounted, tax, total valueobjects.Money) {
	hundred := decimal.NewFromInt(100)
	discountFactor := hundred.Sub(discount.Decimal()).Div(hundred)
	disc, _ := valueobjects.MoneyFromDecimal(base.Decimal().Mul(discountFactor))
	discounted = disc.RoundToCurrencyPrecision()
	tax = discounted.MulPercent(taxRate)
	total = discounted.Add(tax)
	return
}
