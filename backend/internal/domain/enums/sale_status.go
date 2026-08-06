package enums

// SaleStatus is the lifecycle state of a sale.
type SaleStatus string

const (
	// SaleStatusPending — created, no payment applied yet.
	SaleStatusPending SaleStatus = "pending"
	// SaleStatusPartial — at least one payment applied, balance > 0.
	SaleStatusPartial SaleStatus = "partial"
	// SaleStatusPaid — fully paid.
	SaleStatusPaid SaleStatus = "paid"
	// SaleStatusCancelled — voided; stock returned; accounting reversed.
	SaleStatusCancelled SaleStatus = "cancelled"
)

func AllSaleStatuses() []SaleStatus {
	return []SaleStatus{SaleStatusPending, SaleStatusPartial, SaleStatusPaid, SaleStatusCancelled}
}

func (s SaleStatus) Valid() bool {
	switch s {
	case SaleStatusPending, SaleStatusPartial, SaleStatusPaid, SaleStatusCancelled:
		return true
	}
	return false
}

// IsTerminal returns true if no further payments or modifications
// are allowed in this state.
func (s SaleStatus) IsTerminal() bool {
	return s == SaleStatusPaid || s == SaleStatusCancelled
}

func (s SaleStatus) String() string { return string(s) }
