package services

import (
	"testing"

	"vfinancy/backend/internal/domain/valueobjects"
)

func mustMoney(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, _ := valueobjects.MoneyFromString(s)
	return m
}

func mustPercent(t *testing.T, s string) valueobjects.Percentage {
	t.Helper()
	p, _ := valueobjects.PercentageFromString(s)
	return p
}

func TestTaxExclusive(t *testing.T) {
	calc := NewTaxCalculator()
	base := mustMoney(t, "100.00")
	rate := mustPercent(t, "18")
	tax, total := calc.TaxExclusive(base, rate)
	if tax.String() != "18.00" {
		t.Errorf("tax: %s", tax)
	}
	if total.String() != "118.00" {
		t.Errorf("total: %s", total)
	}
}

func TestTaxInclusive(t *testing.T) {
	calc := NewTaxCalculator()
	total := mustMoney(t, "118.00")
	rate := mustPercent(t, "18")
	tax, base := calc.TaxInclusive(total, rate)
	// 118 * 18 / 118 = 18
	if tax.String() != "18.00" {
		t.Errorf("tax: %s", tax)
	}
	if base.String() != "100.00" {
		t.Errorf("base: %s", base)
	}
}

func TestApplyDiscountExclusive(t *testing.T) {
	calc := NewTaxCalculator()
	base := mustMoney(t, "100.00")
	discount := mustPercent(t, "10")
	tax := mustPercent(t, "18")
	discounted, tx, total := calc.ApplyDiscountExclusive(base, discount, tax)
	// discounted = 100 * 0.9 = 90
	if discounted.String() != "90.00" {
		t.Errorf("discounted: %s", discounted)
	}
	// tax = 90 * 0.18 = 16.20
	if tx.String() != "16.20" {
		t.Errorf("tax: %s", tx)
	}
	// total = 90 + 16.20 = 106.20
	if total.String() != "106.20" {
		t.Errorf("total: %s", total)
	}
}

func TestProfitCalculatorMarginPercent(t *testing.T) {
	c := NewProfitCalculator()
	price := mustMoney(t, "10")
	cost := mustMoney(t, "4")
	if got := c.GrossMarginPercent(price, cost); got != 60 {
		t.Errorf("margin%%: %v", got)
	}
	if got := c.GrossMarginPercent(mustMoney(t, "0"), cost); got != 0 {
		t.Errorf("zero price: %v", got)
	}
}

func TestProfitCalculatorGrossMargin(t *testing.T) {
	c := NewProfitCalculator()
	price := mustMoney(t, "100")
	cost := mustMoney(t, "40")
	if got := c.GrossMargin(price, cost).String(); got != "60.00" {
		t.Errorf("margin: %s", got)
	}
}
