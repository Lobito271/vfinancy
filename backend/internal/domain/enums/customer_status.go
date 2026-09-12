package enums

// CustomerStatus is the lifecycle state of a customer.
type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
)

// AllCustomerStatuses returns every valid CustomerStatus. Useful for
// validation and for iterating to expose values in a UI select.
func AllCustomerStatuses() []CustomerStatus {
	return []CustomerStatus{
		CustomerStatusActive,
		CustomerStatusInactive,
	}
}

func (s CustomerStatus) Valid() bool {
	switch s {
	case CustomerStatusActive, CustomerStatusInactive:
		return true
	}
	return false
}

func (s CustomerStatus) String() string { return string(s) }
