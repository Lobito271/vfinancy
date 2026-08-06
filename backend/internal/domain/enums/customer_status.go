// Package enums defines strongly typed enumerations used across the
// domain. Each enum is a typed string with a closed set of valid values,
// constructor functions that validate on creation, and a parse function
// to recover the value from the database.
package enums

// CustomerStatus is the lifecycle state of a customer.
type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusBlocked  CustomerStatus = "blocked"
)

// AllCustomerStatuses returns every valid CustomerStatus. Useful for
// validation and for iterating to expose values in a UI select.
func AllCustomerStatuses() []CustomerStatus {
	return []CustomerStatus{
		CustomerStatusActive,
		CustomerStatusInactive,
		CustomerStatusBlocked,
	}
}

func (s CustomerStatus) Valid() bool {
	switch s {
	case CustomerStatusActive, CustomerStatusInactive, CustomerStatusBlocked:
		return true
	}
	return false
}

func (s CustomerStatus) String() string { return string(s) }
