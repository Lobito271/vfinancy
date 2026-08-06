package enums

// SupplierStatus is the lifecycle state of a supplier.
type SupplierStatus string

const (
	SupplierStatusActive   SupplierStatus = "active"
	SupplierStatusInactive SupplierStatus = "inactive"
)

func AllSupplierStatuses() []SupplierStatus {
	return []SupplierStatus{SupplierStatusActive, SupplierStatusInactive}
}

func (s SupplierStatus) Valid() bool {
	return s == SupplierStatusActive || s == SupplierStatusInactive
}

func (s SupplierStatus) String() string { return string(s) }
