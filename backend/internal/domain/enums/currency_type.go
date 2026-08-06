package enums

// CurrencyType classifies a currency's role in the system. Most
// records only deal with Transactional or Functional currencies.
type CurrencyType string

const (
	// CurrencyTypeFunctional — the company's reporting currency. Set
	// once per company (see companies.functional_currency_code).
	CurrencyTypeFunctional CurrencyType = "functional"
	// CurrencyTypeTransactional — used in customer/supplier documents.
	CurrencyTypeTransactional CurrencyType = "transactional"
	// CurrencyTypeReference — only displayed (e.g. BCRP published rates
	// for currencies the company does not transact in).
	CurrencyTypeReference CurrencyType = "reference"
)

func (c CurrencyType) Valid() bool {
	switch c {
	case CurrencyTypeFunctional, CurrencyTypeTransactional, CurrencyTypeReference:
		return true
	}
	return false
}

func (c CurrencyType) String() string { return string(c) }
