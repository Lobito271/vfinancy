package enums

// DocumentType identifies the kind of personal identification document
// (DNI or RUC). TypeNone means the customer has no document.
type DocumentType string

const (
	TypeNone        DocumentType = ""
	DocumentTypeDNI DocumentType = "DNI"
	DocumentTypeRUC DocumentType = "RUC"
)

// Valid reports whether d is a known document type, including the
// empty (no document) type.
func (d DocumentType) Valid() bool {
	switch d {
	case TypeNone, DocumentTypeDNI, DocumentTypeRUC:
		return true
	}
	return false
}

func (d DocumentType) String() string { return string(d) }
