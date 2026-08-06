package enums

// PurchaseStatus is the lifecycle state of a purchase order.
type PurchaseStatus string

const (
	// PurchaseStatusPending — created, not yet received from the supplier.
	PurchaseStatusPending PurchaseStatus = "pending"
	// PurchaseStatusReceived — goods received, awaiting payment.
	PurchaseStatusReceived PurchaseStatus = "received"
	// PurchaseStatusPaid — fully paid.
	PurchaseStatusPaid PurchaseStatus = "paid"
	// PurchaseStatusReconciled — paid and matched against the supplier
	// invoice / bank statement.
	PurchaseStatusReconciled PurchaseStatus = "reconciled"
	// PurchaseStatusCancelled — voided; stock returned if applicable.
	PurchaseStatusCancelled PurchaseStatus = "cancelled"
)

func AllPurchaseStatuses() []PurchaseStatus {
	return []PurchaseStatus{
		PurchaseStatusPending,
		PurchaseStatusReceived,
		PurchaseStatusPaid,
		PurchaseStatusReconciled,
		PurchaseStatusCancelled,
	}
}

func (s PurchaseStatus) Valid() bool {
	switch s {
	case PurchaseStatusPending, PurchaseStatusReceived, PurchaseStatusPaid,
		PurchaseStatusReconciled, PurchaseStatusCancelled:
		return true
	}
	return false
}

func (s PurchaseStatus) IsTerminal() bool {
	return s == PurchaseStatusReconciled || s == PurchaseStatusCancelled
}

func (s PurchaseStatus) String() string { return string(s) }
