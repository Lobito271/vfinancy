package valueobjects

import (
	"regexp"
	"strings"

	"vfinancy/backend/internal/domain/enums"
)

var (
	dniRe = regexp.MustCompile(`^[0-9]{8}$`)
	rucRe = regexp.MustCompile(`^(?:10|20)[0-9]{9}$`)
)

// DocumentNumber pairs an identification document type with its raw
// number. The zero value is a valid, optional document.
type DocumentNumber struct {
	docType enums.DocumentType
	value   string
}

// NewDocumentNumber validates the (type, number) pair. An empty number
// yields the zero value (no document). When a number is present the
// type must be DNI or RUC and the number must match its pattern.
func NewDocumentNumber(t enums.DocumentType, number string) (DocumentNumber, error) {
	number = strings.TrimSpace(number)
	if number == "" {
		return DocumentNumber{}, nil
	}
	if t != enums.DocumentTypeDNI && t != enums.DocumentTypeRUC {
		return DocumentNumber{}, wrapInvalid("document type must be DNI or RUC when a number is present")
	}
	if err := ValidateDocument(t, number); err != nil {
		return DocumentNumber{}, err
	}
	return DocumentNumber{docType: t, value: number}, nil
}

// ValidateDocument checks a DNI (8 digits) / RUC (11 digits, prefix
// 10|20) pair. An empty type+number pair is valid (optional document).
func ValidateDocument(t enums.DocumentType, number string) error {
	if t == enums.TypeNone && number == "" {
		return nil
	}
	if t == enums.TypeNone {
		return wrapInvalid("document type is required when a number is present")
	}
	if number == "" {
		return wrapInvalid("document number is required when a type is present")
	}
	switch t {
	case enums.DocumentTypeDNI:
		if !dniRe.MatchString(number) {
			return wrapInvalid("DNI must be 8 digits")
		}
	case enums.DocumentTypeRUC:
		if !rucRe.MatchString(number) {
			return wrapInvalid("RUC must be 11 digits starting with 10 or 20")
		}
	default:
		return wrapInvalid("document type is invalid: " + string(t))
	}
	return nil
}

func (d DocumentNumber) Type() enums.DocumentType { return d.docType }
func (d DocumentNumber) Number() string           { return d.value }

func (d DocumentNumber) String() string {
	if d.value == "" {
		return ""
	}
	return string(d.docType) + ":" + d.value
}

func (d DocumentNumber) IsZero() bool { return d.value == "" }
