package valueobjects

import (
	"regexp"
	"strings"
)

// barcodeRe validates an EAN-13 barcode (the global standard). EAN-8
// and UPC-A barcodes are not yet supported but can be added by
// extending this regex.
var ean13Re = regexp.MustCompile(`^\d{13}$`)

// Barcode is a globally unique product barcode.
type Barcode struct {
	value string
}

// NewBarcode validates an EAN-13 barcode.
func NewBarcode(s string) (Barcode, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Barcode{}, wrapInvalid("barcode is empty")
	}
	if !ean13Re.MatchString(s) {
		return Barcode{}, wrapInvalid("barcode must be 13 digits (EAN-13)")
	}
	if !ean13Valid(s) {
		return Barcode{}, wrapInvalid("barcode check digit is invalid")
	}
	return Barcode{value: s}, nil
}

// OptionalBarcode accepts an empty string and returns the zero value.
func OptionalBarcode(s string) (Barcode, error) {
	if strings.TrimSpace(s) == "" {
		return Barcode{}, nil
	}
	return NewBarcode(s)
}

func (b Barcode) String() string       { return b.value }
func (b Barcode) IsZero() bool         { return b.value == "" }
func (b Barcode) Equals(other Barcode) bool { return b.value == other.value }

// ean13Valid verifies the EAN-13 check digit. The last digit of a
// valid EAN-13 is a checksum: sum of (odd-position * 1 + even-position * 3),
// modulo 10, must be 0 (or 10 → 0).
func ean13Valid(s string) bool {
	digits := make([]int, 13)
	for i, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		digits[i] = int(c - '0')
	}
	sum := 0
	for i := 0; i < 12; i++ {
		if i%2 == 0 {
			sum += digits[i]
		} else {
			sum += digits[i] * 3
		}
	}
	check := (10 - (sum % 10)) % 10
	return check == digits[12]
}
