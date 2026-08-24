package valueobjects

import (
	"strings"
)

// CurrencyCode is an ISO 4217 alpha-3 currency code (PEN, USD, EUR, ...).
// It is always upper case.
type CurrencyCode struct {
	code string
}

// NewCurrencyCode validates a 3-letter ISO 4217 code.
func NewCurrencyCode(s string) (CurrencyCode, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	if len(s) != 3 {
		return CurrencyCode{}, wrapInvalid("currency code must be 3 letters (ISO 4217)")
	}
	for _, c := range s {
		if c < 'A' || c > 'Z' {
			return CurrencyCode{}, wrapInvalid("currency code must be ASCII letters")
		}
	}
	return CurrencyCode{code: s}, nil
}

// MustCurrencyCode is a test/seed convenience that panics on error.
func MustCurrencyCode(s string) CurrencyCode {
	c, err := NewCurrencyCode(s)
	if err != nil {
		panic(err)
	}
	return c
}

// PEN is the Peruvian Sol, the company's functional currency.
var PEN = MustCurrencyCode("PEN")

// USD is the US Dollar, the import order currency.
var USD = MustCurrencyCode("USD")

func (c CurrencyCode) String() string { return c.code }

func (c CurrencyCode) IsZero() bool { return c.code == "" }

func (c CurrencyCode) Equals(other CurrencyCode) bool { return c.code == other.code }
