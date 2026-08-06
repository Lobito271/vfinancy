package valueobjects

import (
	"regexp"
	"strings"
)

// emailRe is a pragmatic RFC 5322 subset. We intentionally do not try
// to be exhaustive — RFC 5322 is famously irregular. What we enforce is
// enough to catch the most common typos.
var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is a validated email address. The string is normalized to
// lowercase on construction.
type Email struct {
	value string
}

// NewEmail parses and validates an email. Empty input is rejected.
func NewEmail(s string) (Email, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Email{}, wrapInvalid("email is empty")
	}
	s = strings.ToLower(s)
	if !emailRe.MatchString(s) {
		return Email{}, wrapInvalid("email is malformed: " + s)
	}
	return Email{value: s}, nil
}

// MustEmail is a test/seed convenience that panics on error.
func MustEmail(s string) Email {
	e, err := NewEmail(s)
	if err != nil {
		panic(err)
	}
	return e
}

func (e Email) String() string { return e.value }

func (e Email) Equals(other Email) bool { return e.value == other.value }

func (e Email) IsZero() bool { return e.value == "" }
