package enums

// JournalType classifies the source of a journal entry. Used both for
// reporting and for routing reversal logic.
type JournalType string

const (
	// JournalTypeSale — auto-generated from a sale posting.
	JournalTypeSale JournalType = "sale"
	// JournalTypePurchase — auto-generated from a purchase posting.
	JournalTypePurchase JournalType = "purchase"
	// JournalTypePayment — auto-generated from a customer/supplier
	// payment.
	JournalTypePayment JournalType = "payment"
	// JournalTypeManual — entered by an accountant.
	JournalTypeManual JournalType = "manual"
	// JournalTypeAdjustment — adjustment / revaluation.
	JournalTypeAdjustment JournalType = "adjustment"
	// JournalTypeClosing — period-closing entry.
	JournalTypeClosing JournalType = "closing"
	// JournalTypeOpening — period-opening entry.
	JournalTypeOpening JournalType = "opening"
)

func (j JournalType) Valid() bool {
	switch j {
	case JournalTypeSale, JournalTypePurchase, JournalTypePayment,
		JournalTypeManual, JournalTypeAdjustment, JournalTypeClosing, JournalTypeOpening:
		return true
	}
	return false
}

func (j JournalType) String() string { return string(j) }
