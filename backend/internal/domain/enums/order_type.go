package enums

// OrderType is the kind of purchase order. "general" orders restock
// the store; "customer" orders are placed for a specific customer and
// carry an anticipo (down payment) with the balance tracked as
// "por cobrar".
type OrderType string

const (
	// OrderTypeGeneral — store stock purchase from a supplier.
	OrderTypeGeneral OrderType = "general"
	// OrderTypeCustomer — purchase placed for a specific customer order.
	OrderTypeCustomer OrderType = "customer"
)

// AllOrderTypes returns every valid order type.
func AllOrderTypes() []OrderType {
	return []OrderType{OrderTypeGeneral, OrderTypeCustomer}
}

// Valid reports whether the order type is known.
func (t OrderType) Valid() bool {
	switch t {
	case OrderTypeGeneral, OrderTypeCustomer:
		return true
	}
	return false
}

// String returns the canonical string form.
func (t OrderType) String() string { return string(t) }
