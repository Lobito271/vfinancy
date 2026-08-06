package valueobjects

import (
	"strings"
	"time"
)

// Address is a postal address. We intentionally do not try to model
// structured international address components (street / city / region /
// postal code / country) at the value-object level; that level of
// granularity is country-specific and belongs in master data tables.
type Address struct {
	lines []string
}

// NewAddress builds an Address from 1..N free-form lines. Empty input
// is rejected.
func NewAddress(lines ...string) (Address, error) {
	cleaned := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			cleaned = append(cleaned, l)
		}
	}
	if len(cleaned) == 0 {
		return Address{}, wrapInvalid("address is empty")
	}
	return Address{lines: cleaned}, nil
}

func (a Address) Lines() []string { return a.lines }
func (a Address) IsEmpty() bool   { return len(a.lines) == 0 }

func (a Address) String() string {
	return strings.Join(a.lines, ", ")
}

func (a Address) Equals(other Address) bool {
	if len(a.lines) != len(other.lines) {
		return false
	}
	for i := range a.lines {
		if a.lines[i] != other.lines[i] {
			return false
		}
	}
	return true
}

// FullName is a non-empty, trimmed person name. Stored once on User
// and on customer/supplier contacts.
type FullName struct {
	value string
}

func NewFullName(s string) (FullName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return FullName{}, wrapInvalid("full name is empty")
	}
	if len(s) > 200 {
		return FullName{}, wrapInvalid("full name exceeds 200 characters")
	}
	return FullName{value: s}, nil
}

// MustFullName is a test/seed convenience that panics on error.
func MustFullName(s string) FullName {
	n, err := NewFullName(s)
	if err != nil {
		panic(err)
	}
	return n
}

func (n FullName) String() string         { return n.value }
func (n FullName) Equals(other FullName) bool { return n.value == other.value }

// LotNumber is the optional supplier/manufacturer lot code.
type LotNumber string

func NewLotNumber(s string) (LotNumber, error) {
	s = strings.TrimSpace(s)
	if len(s) > 50 {
		return LotNumber(""), wrapInvalid("lot number exceeds 50 characters")
	}
	return LotNumber(s), nil
}

func (l LotNumber) String() string  { return string(l) }
func (l LotNumber) IsEmpty() bool    { return string(l) == "" }

// ShortCode is a generic short identifier (1..20 chars, uppercase,
// letters/digits/dots/dashes). Used for company codes, branch codes,
// role codes, etc.
type ShortCode string

func NewShortCode(s string) (ShortCode, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	if s == "" {
		return ShortCode(""), wrapInvalid("short code is empty")
	}
	if len(s) > 20 {
		return ShortCode(""), wrapInvalid("short code exceeds 20 characters")
	}
	for _, c := range s {
		ok := (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.'
		if !ok {
			return ShortCode(""), wrapInvalid("short code may only contain A-Z, 0-9, '-', '_', '.'")
		}
	}
	return ShortCode(s), nil
}

func (s ShortCode) String() string { return string(s) }
func (s ShortCode) IsEmpty() bool { return string(s) == "" }

// Date is an inclusive day (no time, no timezone). It is the same Go
// type as time.Time but the constructors reject time-of-day input.
//
// The reason for the type alias is to make signatures self-documenting
// and to prevent passing a datetime where a date is expected.
type Date = time.Time

// NewDate builds a Date from a year/month/day triple.
func NewDate(year int, month time.Month, day int) (Date, error) {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || t.Month() != month || t.Day() != day {
		return Date{}, wrapInvalid("date is out of range")
	}
	return t, nil
}

// NewDateFromTime truncates a time.Time to its date part. Use this when
// accepting a timestamp from the outside world (e.g. an API request).
func NewDateFromTime(t time.Time) Date {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// DateFromString parses a date in YYYY-MM-DD format.
func DateFromString(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return Date{}, wrapInvalid("date is malformed: " + s)
	}
	return t, nil
}

// IsAfter / IsBefore / IsSameDay / AddDays.
func IsSameDay(a, b Date) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func AddDays(d Date, days int) Date {
	return d.AddDate(0, 0, days)
}
