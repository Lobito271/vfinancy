package enums

// TaxCategory is the SUNAT-style classification of how a product
// or document line is treated for tax purposes.
type TaxCategory string

const (
	// TaxCategoryTaxed — grava IGV.
	TaxCategoryTaxed TaxCategory = "taxed"
	// TaxCategoryExempt — exonerado.
	TaxCategoryExempt TaxCategory = "exempt"
	// TaxCategoryUnaffected — inafecto.
	TaxCategoryUnaffected TaxCategory = "unaffected"
	// TaxCategoryExport — exported, IGV not applicable.
	TaxCategoryExport TaxCategory = "export"
)

func (t TaxCategory) Valid() bool {
	switch t {
	case TaxCategoryTaxed, TaxCategoryExempt, TaxCategoryUnaffected, TaxCategoryExport:
		return true
	}
	return false
}

func (t TaxCategory) String() string { return string(t) }
