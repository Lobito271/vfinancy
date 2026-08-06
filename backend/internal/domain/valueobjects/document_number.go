package valueobjects

import (
	"regexp"
	"strings"

	"vfinancy/backend/internal/domain/enums"
)

// dniRe is 8 digits; rucRe is 11 digits starting with 10 or 20.
var (
	dniRe  = regexp.MustCompile(`^\d{8}$`)
	rucRe  = regexp.MustCompile(`^[12]0\d{9}$`)
	ceRe   = regexp.MustCompile(`^\d{9,12}$`)
)

// DocumentNumber is a country-aware identification number. It pairs a
// document type (DNI / RUC / CE / PASSPORT) with the raw number and
// validates that the number matches the expected pattern for its type.
type DocumentNumber struct {
	docType enums.DocumentType
	value   string
}

// NewDocumentNumber validates the (type, number) pair.
func NewDocumentNumber(docType enums.DocumentType, number string) (DocumentNumber, error) {
	if !docType.Valid() {
		return DocumentNumber{}, wrapInvalid("document type is invalid: " + string(docType))
	}
	number = strings.TrimSpace(number)
	switch docType {
	case enums.DocumentTypeDNI:
		if !dniRe.MatchString(number) {
			return DocumentNumber{}, wrapInvalid("DNI must be 8 digits")
		}
	case enums.DocumentTypeRUC:
		if !rucRe.MatchString(number) {
			return DocumentNumber{}, wrapInvalid("RUC must be 11 digits starting with 10 or 20")
		}
	case enums.DocumentTypeCE:
		if !ceRe.MatchString(number) {
			return DocumentNumber{}, wrapInvalid("CE must be 9-12 digits")
		}
	case enums.DocumentTypePassport:
		if len(number) < 5 || len(number) > 20 {
			return DocumentNumber{}, wrapInvalid("passport length out of range")
		}
	}
	return DocumentNumber{docType: docType, value: number}, nil
}

func (d DocumentNumber) Type() enums.DocumentType { return d.docType }
func (d DocumentNumber) Number() string         { return d.value }
func (d DocumentNumber) String() string         { return string(d.docType) + ":" + d.value }

func (d DocumentNumber) IsZero() bool { return d.value == "" }

func (d DocumentNumber) Equals(other DocumentNumber) bool {
	return d.docType == other.docType && d.value == other.value
}
