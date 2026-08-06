package enums

// DocumentType is the per-country identification document type
// (DNI, RUC, CE, etc.). It is *separate* from the fiscal document type
// used in sales/purchases (FACTURA, BOLETA, NC, ND, etc.).
type DocumentType string

const (
	DocumentTypeDNI       DocumentType = "DNI"
	DocumentTypeRUC       DocumentType = "RUC"
	DocumentTypeCE        DocumentType = "CE"
	DocumentTypePassport  DocumentType = "PASSPORT"
)

func (d DocumentType) Valid() bool {
	switch d {
	case DocumentTypeDNI, DocumentTypeRUC, DocumentTypeCE, DocumentTypePassport:
		return true
	}
	return false
}

func (d DocumentType) String() string { return string(d) }
