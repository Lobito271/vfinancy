package valueobjects

import (
	"strings"

	"github.com/shopspring/decimal"
)

// PercentagePrecision is the number of fractional digits stored.
// 4 digits gives 0.0001% precision which is more than enough for tax
// and discount rates.
const PercentagePrecision int32 = 4

// Percentage is an immutable value in the range [0, 100], with up to
// 4 decimal places. Examples: 18% (IGV), 0.5% (financial transactions
// tax), 12.3456% (custom discounts).
type Percentage struct {
	d decimal.Decimal
}

// ZeroPercent returns 0.
func ZeroPercent() Percentage { return Percentage{d: decimal.Zero} }

// PercentageFromDecimal validates the value is in [0, 100] and rounds.
func PercentageFromDecimal(d decimal.Decimal) (Percentage, error) {
	if d.IsNegative() {
		return Percentage{}, wrapInvalid("percentage cannot be negative")
	}
	if d.GreaterThan(decimal.NewFromInt(100)) {
		return Percentage{}, wrapInvalid("percentage cannot exceed 100")
	}
	return Percentage{d: d.Round(PercentagePrecision)}, nil
}

// PercentageFromString parses "18" or "18.5" or "0.5".
func PercentageFromString(s string) (Percentage, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Percentage{}, wrapInvalid("percentage value is empty")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Percentage{}, wrapInvalid("percentage value is malformed: " + s)
	}
	return PercentageFromDecimal(d)
}

// Decimal returns the underlying value. Use it for display formatting.
func (p Percentage) Decimal() decimal.Decimal { return p.d }

// AsDecimal returns the ratio in [0, 1] — e.g. 0.18 for 18%. Use this
// for tax and discount math: amount * percentage.AsDecimal().
func (p Percentage) AsDecimal() decimal.Decimal {
	return p.d.Div(decimal.NewFromInt(100))
}

// IsZero reports p == 0.
func (p Percentage) IsZero() bool { return p.d.IsZero() }

func (p Percentage) Equals(other Percentage) bool { return p.d.Equal(other.d) }

func (p Percentage) String() string {
	return p.d.StringFixed(PercentagePrecision)
}
