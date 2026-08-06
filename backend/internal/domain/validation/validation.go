// Package validation provides the low-level building blocks used by value
// objects and entities to validate inputs. Each function returns a
// domain error (or nil) so callers can wrap with context.
package validation

import (
	"strings"
	"unicode/utf8"

	derrors "vfinancy/backend/internal/domain/errors"
)

// RequiredString returns ErrRequired if s is empty or whitespace-only.
func RequiredString(field, s string) error {
	if strings.TrimSpace(s) == "" {
		return derrors.Wrap(derrors.ErrRequired, fmtErr(field+" is required"))
	}
	return nil
}

// MaxLength returns an error if s exceeds max.
func MaxLength(field, s string, max int) error {
	if utf8.RuneCountInString(s) > max {
		return derrors.Wrap(derrors.ErrOutOfRange, fmtErr(field+" exceeds "+itoa(max)+" characters"))
	}
	return nil
}

// MinLength returns an error if s is shorter than min.
func MinLength(field, s string, min int) error {
	if utf8.RuneCountInString(s) < min {
		return derrors.Wrap(derrors.ErrOutOfRange, fmtErr(field+" must be at least "+itoa(min)+" characters"))
	}
	return nil
}

// InRange returns an error if n is outside [min, max] (inclusive).
func InRange[T ~int | ~int32 | ~int64 | ~float64](field string, n, min, max T) error {
	if n < min || n > max {
		return derrors.Wrap(derrors.ErrOutOfRange, fmtErr(field+" is out of range"))
	}
	return nil
}

// Positive returns an error if n is not strictly positive.
func Positive[T ~int | ~int32 | ~int64 | ~float64](field string, n T) error {
	if n <= 0 {
		return derrors.Wrap(derrors.ErrOutOfRange, fmtErr(field+" must be positive"))
	}
	return nil
}

// NonNegative returns an error if n is negative.
func NonNegative[T ~int | ~int32 | ~int64 | ~float64](field string, n T) error {
	if n < 0 {
		return derrors.Wrap(derrors.ErrOutOfRange, fmtErr(field+" cannot be negative"))
	}
	return nil
}

// OneOf returns an error if s is not in allowed.
func OneOf(field, s string, allowed ...string) error {
	for _, a := range allowed {
		if s == a {
			return nil
		}
	}
	return derrors.Wrap(derrors.ErrInvalidEnum, fmtErr(field+" is not one of the allowed values"))
}

// fmtErr is a tiny helper to wrap a plain error with context. We use it
// instead of fmt.Errorf to keep the import surface small.
type fieldErr struct{ msg string }

func (e fieldErr) Error() string { return e.msg }
func fmtErr(s string) error      { return fieldErr{msg: s} }

// itoa is a strconv-free integer formatter. Used to avoid pulling in
// strconv just for two call sites.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
