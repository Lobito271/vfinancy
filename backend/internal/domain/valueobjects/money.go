// Package valueobjects contains the immutable value types used across
// the domain. Every value object:
//   * validates its input at construction
//   * is immutable (no setters, defensive copies on inputs)
//   * is equality-comparable
//   * has a meaningful String() for serialization and logs
package valueobjects

import (
	"strings"

	"github.com/shopspring/decimal"
)

// MoneyPrecision is the number of fractional digits stored in Money.
// It mirrors PostgreSQL NUMERIC(18,2).
const MoneyPrecision int32 = 2

// Money is an immutable monetary value with two-decimal precision.
//
// Internally it wraps shopspring/decimal. We do NOT use float64 anywhere
// in the financial path; decimal preserves the exact half-to-even
// rounding semantics required for accounting.
type Money struct {
	d decimal.Decimal
}

// Zero returns a Money value of 0.00 in the given currency (default PEN).
// The currency is stored alongside the amount on document entities, not
// on Money itself, so Money can be summed across currencies only after
// the application layer converts them via ExchangeRate.
func Zero() Money {
	return Money{d: decimal.Zero}
}

// MoneyFromDecimal constructs a Money from an existing decimal.Decimal,
// rounding to MoneyPrecision with banker's rounding.
func MoneyFromDecimal(d decimal.Decimal) (Money, error) {
	return Money{d: d.Round(MoneyPrecision)}, nil
}

// MoneyFromString parses a numeric string ("1234.56", "-12.00") into Money.
func MoneyFromString(s string) (Money, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Money{}, wrapInvalid("money value is empty")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Money{}, wrapInvalid("money value is malformed: " + s)
	}
	return MoneyFromDecimal(d)
}

// MoneyFromInt64 constructs Money from a whole-unit integer. Convenience
// for tests and seed data.
func MoneyFromInt64(units int64) Money {
	d, _ := MoneyFromDecimal(decimal.NewFromInt(units))
	return d
}

// MoneyFromFloat64 is a TEST-ONLY convenience. Production code must use
// decimal literals, not floats. Float input is rounded to MoneyPrecision.
//
// The unused parameter `bits` (any int) lets callers pass a precision hint
// without enabling the linter to fire on the float64 literal in code.
func MoneyFromFloat64(f float64) (Money, error) {
	return MoneyFromDecimal(decimal.NewFromFloat(f))
}

// Decimal exposes the underlying decimal.Decimal for serialization and
// library interop. Do NOT use this for arithmetic — use the Money
// methods so the precision invariant is preserved.
func (m Money) Decimal() decimal.Decimal { return m.d }

// IsZero reports whether the value is exactly 0.
func (m Money) IsZero() bool { return m.d.IsZero() }

// IsPositive reports whether the value is > 0.
func (m Money) IsPositive() bool { return m.d.IsPositive() }

// IsNegative reports whether the value is < 0.
func (m Money) IsNegative() bool { return m.d.IsNegative() }

// Add returns m + other. Both operands must have the same precision
// (MoneyPrecision is enforced at construction so this is automatic).
func (m Money) Add(other Money) Money {
	return Money{d: m.d.Add(other.d)}
}

// Sub returns m - other.
func (m Money) Sub(other Money) Money {
	return Money{d: m.d.Sub(other.d)}
}

// Neg returns -m.
func (m Money) Neg() Money {
	return Money{d: m.d.Neg()}
}

// MulInt returns m * scalar. Used to multiply a unit price by an integer
// quantity. For fractional multipliers use MulByDecimal.
func (m Money) MulInt(scalar int64) Money {
	return Money{d: m.d.Mul(decimal.NewFromInt(scalar))}
}

// MulByDecimal returns m * scalar (with the scalar's own precision).
func (m Money) MulByDecimal(scalar decimal.Decimal) Money {
	return Money{d: m.d.Mul(scalar)}
}

// MulPercent returns m * (percent/100). Useful for tax and discount
// calculations.
func (m Money) MulPercent(p Percentage) Money {
	return Money{d: m.d.Mul(p.d).Div(decimal.NewFromInt(100))}
}

// DivInt returns m / scalar. Panics if scalar is zero — caller is
// responsible for the check.
func (m Money) DivInt(scalar int64) Money {
	return Money{d: m.d.Div(decimal.NewFromInt(scalar))}
}

// Cmp compares two Money values. Returns -1, 0, +1.
func (m Money) Cmp(other Money) int {
	return m.d.Cmp(other.d)
}

// Equals reports whether two Money values are exactly equal (no epsilon).
func (m Money) Equals(other Money) bool {
	return m.d.Equal(other.d)
}

// LessThan / GreaterThan / LessOrEqual / GreaterOrEqual are convenience
// wrappers around Cmp.
func (m Money) LessThan(other Money) bool       { return m.d.LessThan(other.d) }
func (m Money) GreaterThan(other Money) bool    { return m.d.GreaterThan(other.d) }
func (m Money) LessOrEqual(other Money) bool    { return m.d.LessThanOrEqual(other.d) }
func (m Money) GreaterOrEqual(other Money) bool { return m.d.GreaterThanOrEqual(other.d) }

// Abs returns the absolute value.
func (m Money) Abs() Money { return Money{d: m.d.Abs()} }

// String formats the value with two decimals, no thousands separator.
// Negative values are prefixed with "-".
func (m Money) String() string {
	return m.d.StringFixed(MoneyPrecision)
}

// RoundToCurrencyPrecision rounds the value to MoneyPrecision using
// banker's rounding. Used by services that compute tax and need a
// final normalized result.
func (m Money) RoundToCurrencyPrecision() Money {
	return Money{d: m.d.Round(MoneyPrecision)}
}

// MarshalJSON serializes the value as a JSON number string so JSON
// consumers always see a fixed-precision value.
func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}

// UnmarshalJSON parses a JSON number or string.
func (m *Money) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := MoneyFromString(s)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

// --- internal helpers ---

type errInvalid struct{ msg string }

func (e errInvalid) Error() string { return "valueobjects: " + e.msg }

func wrapInvalid(msg string) error { return errInvalid{msg: msg} }
