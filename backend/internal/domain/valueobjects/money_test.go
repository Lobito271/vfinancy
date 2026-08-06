package valueobjects

import (
	"testing"
)

func TestMoneyFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple", "100", "100.00", false},
		{"with decimals", "1234.56", "1234.56", false},
		{"negative", "-12.34", "-12.34", false},
		{"zero", "0", "0.00", false},
		{"empty", "", "", true},
		{"bad format", "abc", "", true},
		{"rounds to 2dp", "1.005", "1.01", false}, // banker's rounding
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := MoneyFromString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if err == nil && m.String() != tc.want {
				t.Errorf("got %s, want %s", m.String(), tc.want)
			}
		})
	}
}

func TestMoneyAddSubNegAbs(t *testing.T) {
	a := mustMoney(t, "100.00")
	b := mustMoney(t, "30.50")
	if a.Add(b).String() != "130.50" {
		t.Errorf("add: %s", a.Add(b))
	}
	if a.Sub(b).String() != "69.50" {
		t.Errorf("sub: %s", a.Sub(b))
	}
	if b.Neg().String() != "-30.50" {
		t.Errorf("neg: %s", b.Neg())
	}
	if b.Sub(a).Abs().String() != "69.50" {
		t.Errorf("abs: %s", b.Sub(a).Abs())
	}
}

func TestMoneyMulIntMulPercent(t *testing.T) {
	price := mustMoney(t, "12.50")
	qty := int64(4)
	if price.MulInt(qty).String() != "50.00" {
		t.Errorf("mulInt: %s", price.MulInt(qty))
	}
	tax := mustPercent(t, "18")
	// 12.50 * 18% = 2.25
	if price.MulPercent(tax).String() != "2.25" {
		t.Errorf("mulPercent: %s", price.MulPercent(tax))
	}
}

func TestMoneyComparisons(t *testing.T) {
	a := mustMoney(t, "10.00")
	b := mustMoney(t, "20.00")
	if !a.LessThan(b) {
		t.Error("a < b")
	}
	if !b.GreaterThan(a) {
		t.Error("b > a")
	}
	if !a.LessOrEqual(a) {
		t.Error("a <= a")
	}
	if a.Equals(b) {
		t.Error("a != b")
	}
}

func TestMoneyIsZero(t *testing.T) {
	z := Zero()
	if !z.IsZero() {
		t.Error("Zero() should be zero")
	}
	nz := mustMoney(t, "0.01")
	if nz.IsZero() {
		t.Error("0.01 should not be zero")
	}
}

func TestMoneyRoundToCurrencyPrecision(t *testing.T) {
	m, _ := MoneyFromString("1.234")
	if m.RoundToCurrencyPrecision().String() != "1.23" {
		t.Errorf("round: %s", m.RoundToCurrencyPrecision())
	}
}

func mustMoney(t *testing.T, s string) Money {
	t.Helper()
	m, err := MoneyFromString(s)
	if err != nil {
		t.Fatalf("mustMoney(%q): %v", s, err)
	}
	return m
}
