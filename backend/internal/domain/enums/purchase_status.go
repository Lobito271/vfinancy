package enums

// PurchaseStatus is the lifecycle state of a purchase order:
// pending → received | cancelled.
type PurchaseStatus string

const (
	// PurchaseStatusPending — created, not yet received from the supplier.
	PurchaseStatusPending PurchaseStatus = "pending"
	// PurchaseStatusReceived — goods received into inventory.
	PurchaseStatusReceived PurchaseStatus = "received"
	// PurchaseStatusCancelled — voided; stock returned if it had been received.
	PurchaseStatusCancelled PurchaseStatus = "cancelled"
)

// AllPurchaseStatuses returns every valid purchase status.
func AllPurchaseStatuses() []PurchaseStatus {
	return []PurchaseStatus{
		PurchaseStatusPending,
		PurchaseStatusReceived,
		PurchaseStatusCancelled,
	}
}

// Valid reports whether the status is a known member.
func (s PurchaseStatus) Valid() bool {
	switch s {
	case PurchaseStatusPending, PurchaseStatusReceived, PurchaseStatusCancelled:
		return true
	}
	return false
}

// IsTerminal reports whether the status ends the lifecycle.
func (s PurchaseStatus) IsTerminal() bool {
	return s == PurchaseStatusCancelled
}

// String returns the canonical string form.
func (s PurchaseStatus) String() string { return string(s) }
