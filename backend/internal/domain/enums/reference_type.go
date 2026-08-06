package enums

// ReferenceType is the polymorphic owner of a soft reference (audit
// logs, inventory movements, attachments). A reference is always
// (ReferenceType, ID). The actual FK is enforced by application code
// or by a CHECK + the matching record in the target table.
type ReferenceType string

const (
	ReferenceTypeSale        ReferenceType = "sale"
	ReferenceTypePurchase    ReferenceType = "purchase"
	ReferenceTypeTransfer    ReferenceType = "transfer"
	ReferenceTypeAdjustment  ReferenceType = "adjustment"
	ReferenceTypeReturn      ReferenceType = "return"
	ReferenceTypePayment     ReferenceType = "payment"
	ReferenceTypeJournalEntry ReferenceType = "journal_entry"
	ReferenceTypeUser        ReferenceType = "user"
	ReferenceTypeManual      ReferenceType = "manual"
)

func (r ReferenceType) Valid() bool {
	switch r {
	case ReferenceTypeSale, ReferenceTypePurchase, ReferenceTypeTransfer,
		ReferenceTypeAdjustment, ReferenceTypeReturn, ReferenceTypePayment,
		ReferenceTypeJournalEntry, ReferenceTypeUser, ReferenceTypeManual:
		return true
	}
	return false
}

func (r ReferenceType) String() string { return string(r) }
