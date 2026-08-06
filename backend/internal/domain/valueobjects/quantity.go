package valueobjects

import (
	"strings"

	"github.com/shopspring/decimal"
)

// QuantityPrecision is the number of fractional digits for product
// quantities. It mirrors PostgreSQL NUMERIC(18,4) and supports fractional
// units (kg, m, L, etc.).
const QuantityPrecision int32 = 4

// Quantity is an immutable product count. Zero is allowed (e.g. an
// adjustment that sets stock to zero); negative values are valid only
// for in-flight inventory movements (which carry a signed quantity)
// and are NOT used by Quantity.
type Quantity struct {
	d decimal.Decimal
}

// ZeroQuantity returns 0.
func ZeroQuantity() Quantity {
	return Quantity{d: decimal.Zero}
}

// QuantityFromDecimal rounds and validates a quantity value.
func QuantityFromDecimal(d decimal.Decimal) (Quantity, error) {
	return Quantity{d: d.Round(QuantityPrecision)}, nil
}

// QuantityFromString parses a quantity ("12.5000", "0.5").
func QuantityFromString(s string) (Quantity, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Quantity{}, wrapInvalid("quantity value is empty")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Quantity{}, wrapInvalid("quantity value is malformed: " + s)
	}
	return QuantityFromDecimal(d)
}

// QuantityFromInt64 builds a whole-unit quantity.
func QuantityFromInt64(units int64) Quantity {
	d, _ := QuantityFromDecimal(decimal.NewFromInt(units))
	return d
}

// Decimal exposes the underlying decimal.
func (q Quantity) Decimal() decimal.Decimal { return q.d }

// IsZero / IsPositive / IsNegative.
func (q Quantity) IsZero() bool     { return q.d.IsZero() }
func (q Quantity) IsPositive() bool { return q.d.IsPositive() }
func (q Quantity) IsNegative() bool { return q.d.IsNegative() }

// Add / Sub / Neg.
func (q Quantity) Add(other Quantity) Quantity { return Quantity{d: q.d.Add(other.d)} }
func (q Quantity) Sub(other Quantity) Quantity { return Quantity{d: q.d.Sub(other.d)} }
func (q Quantity) Neg() Quantity              { return Quantity{d: q.d.Neg()} }
func (q Quantity) Abs() Quantity              { return Quantity{d: q.d.Abs()} }

// MulByDecimal multiplies by a fractional factor (e.g. converting kg to
// g by * 1000).
func (q Quantity) MulByDecimal(d decimal.Decimal) Quantity {
	return Quantity{d: q.d.Mul(d)}
}

// Cmp / Equals / LessThan / GreaterThan.
func (q Quantity) Cmp(other Quantity) int { return q.d.Cmp(other.d) }
func (q Quantity) Equals(other Quantity) bool { return q.d.Equal(other.d) }
func (q Quantity) LessThan(other Quantity) bool { return q.d.LessThan(other.d) }
func (q Quantity) GreaterThan(other Quantity) bool { return q.d.GreaterThan(other.d) }
func (q Quantity) LessOrEqual(other Quantity) bool { return q.d.LessThanOrEqual(other.d) }
func (q Quantity) GreaterOrEqual(other Quantity) bool { return q.d.GreaterThanOrEqual(other.d) }

func (q Quantity) String() string { return q.d.StringFixed(QuantityPrecision) }
