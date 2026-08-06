package valueobjects

import (
	"strings"

	"github.com/shopspring/decimal"
)

// ExchangeRatePrecision mirrors PostgreSQL NUMERIC(18,6).
const ExchangeRatePrecision int32 = 6

// ExchangeRate is the multiplier used to convert a transactional
// currency amount into the company's functional currency:
//
//	amount_in_functional = amount_in_transactional * rate
//
// Rates are stored as snapshots on every document (see the database
// schema). The domain layer is concerned only with arithmetic; the
// application layer decides where the rate comes from.
type ExchangeRate struct {
	d decimal.Decimal
}

// One returns an exchange rate of 1.0 (no conversion).
func One() ExchangeRate {
	d, _ := ExchangeRateFromDecimal(decimal.NewFromInt(1))
	return d
}

// ExchangeRateFromDecimal validates the value is positive and rounds.
func ExchangeRateFromDecimal(d decimal.Decimal) (ExchangeRate, error) {
	if d.IsZero() || d.IsNegative() {
		return ExchangeRate{}, wrapInvalid("exchange rate must be positive")
	}
	return ExchangeRate{d: d.Round(ExchangeRatePrecision)}, nil
}

// ExchangeRateFromString parses "3.75" or "1.000000".
func ExchangeRateFromString(s string) (ExchangeRate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ExchangeRate{}, wrapInvalid("exchange rate is empty")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return ExchangeRate{}, wrapInvalid("exchange rate is malformed: " + s)
	}
	return ExchangeRateFromDecimal(d)
}

func (r ExchangeRate) Decimal() decimal.Decimal { return r.d }

// Convert returns amount * rate as a Money value in the functional
// currency. The input Money is in the transactional currency.
func (r ExchangeRate) Convert(m Money) Money {
	return Money{d: m.d.Mul(r.d).Round(MoneyPrecision)}
}

func (r ExchangeRate) Equals(other ExchangeRate) bool { return r.d.Equal(other.d) }

func (r ExchangeRate) String() string { return r.d.StringFixed(ExchangeRatePrecision) }
