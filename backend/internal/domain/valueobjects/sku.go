package valueobjects

import (
	"regexp"
	"strings"
)

// skuRe permits letters, digits, dashes, underscores and dots. SKUs
// must be at least 1 character and at most 50 (matching PostgreSQL's
// VARCHAR(50) for products.sku).
var skuRe = regexp.MustCompile(`^[A-Za-z0-9._\-]{1,50}$`)

// SKU is a stock-keeping unit code. It is normalized to upper case
// for comparison.
type SKU struct {
	value string
}

// NewSKU validates a SKU.
func NewSKU(s string) (SKU, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return SKU{}, wrapInvalid("SKU is empty")
	}
	s = strings.ToUpper(s)
	if !skuRe.MatchString(s) {
		return SKU{}, wrapInvalid("SKU may only contain letters, digits, dots, dashes, underscores (max 50)")
	}
	return SKU{value: s}, nil
}

func (s SKU) String() string  { return s.value }
func (s SKU) IsZero() bool    { return s.value == "" }
func (s SKU) Equals(other SKU) bool { return s.value == other.value }

// MustSKU is a test/seed convenience that panics on error.
func MustSKU(s string) SKU {
	v, err := NewSKU(s)
	if err != nil {
		panic(err)
	}
	return v
}
