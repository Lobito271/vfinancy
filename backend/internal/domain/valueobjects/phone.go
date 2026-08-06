package valueobjects

import (
	"regexp"
	"strings"
)

// phoneRe matches a 9-digit Peruvian mobile/landline number (no country
// code, no separators). Validators in the customer/supplier tables
// strip the country prefix before persisting.
var phoneRe = regexp.MustCompile(`^9\d{8}$`)

// Phone is a validated phone number. The exact format depends on the
// country; this implementation targets the Peruvian format (9 digits).
// Other countries should introduce a country-aware variant.
type Phone struct {
	value string
}

// NewPhone validates a 9-digit phone number.
func NewPhone(s string) (Phone, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	if s == "" {
		return Phone{}, wrapInvalid("phone is empty")
	}
	if !phoneRe.MatchString(s) {
		return Phone{}, wrapInvalid("phone must be 9 digits starting with 9")
	}
	return Phone{value: s}, nil
}

// OptionalPhone returns an empty Phone when s is empty, otherwise validates.
// Useful for fields that are not required.
func OptionalPhone(s string) (Phone, error) {
	if strings.TrimSpace(s) == "" {
		return Phone{}, nil
	}
	return NewPhone(s)
}

func (p Phone) String() string { return p.value }

func (p Phone) IsZero() bool { return p.value == "" }

func (p Phone) Equals(other Phone) bool { return p.value == other.value }
