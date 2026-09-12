package enums

// SaleType distinguishes a sale that ships from local stock from one
// backed by a client order (special purchase).
type SaleType string

const (
	// SaleTypeStock — the goods ship from existing inventory batches.
	SaleTypeStock SaleType = "stock"
	// SaleTypeClientOrder — the goods come from a client order placed
	// with the supplier; no stock is reserved.
	SaleTypeClientOrder SaleType = "client_order"
)

// AllSaleTypes returns every valid sale type.
func AllSaleTypes() []SaleType {
	return []SaleType{SaleTypeStock, SaleTypeClientOrder}
}

// Valid reports whether the value is a recognized sale type.
func (s SaleType) Valid() bool {
	switch s {
	case SaleTypeStock, SaleTypeClientOrder:
		return true
	}
	return false
}

// String returns the raw enum value.
func (s SaleType) String() string { return string(s) }
