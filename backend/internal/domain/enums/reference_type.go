package enums

// ReferenceType is the polymorphic owner of a soft reference (audit
// logs, inventory movements, attachments). A reference is always
// (ReferenceType, ID). The actual FK is enforced by application code
// or by a CHECK + the matching record in the target table.
type ReferenceType string

const (
	ReferenceTypeSale     ReferenceType = "sale"
	ReferenceTypePurchase ReferenceType = "purchase"
)

func (r ReferenceType) Valid() bool {
	switch r {
	case ReferenceTypeSale, ReferenceTypePurchase:
		return true
	}
	return false
}

func (r ReferenceType) String() string { return string(r) }
