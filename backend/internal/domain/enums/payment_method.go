package enums

// PaymentMethod is how a customer payment was settled.
type PaymentMethod string

const (
	// PaymentMethodCash — cash received at the counter.
	PaymentMethodCash PaymentMethod = "cash"
	// PaymentMethodTransfer — bank transfer.
	PaymentMethodTransfer PaymentMethod = "transfer"
	// PaymentMethodOther — any other settlement instrument.
	PaymentMethodOther PaymentMethod = "other"
)

// AllPaymentMethods returns every valid payment method.
func AllPaymentMethods() []PaymentMethod {
	return []PaymentMethod{PaymentMethodCash, PaymentMethodTransfer, PaymentMethodOther}
}

// Valid reports whether the value is a recognized payment method.
func (p PaymentMethod) Valid() bool {
	switch p {
	case PaymentMethodCash, PaymentMethodTransfer, PaymentMethodOther:
		return true
	}
	return false
}

// String returns the raw enum value.
func (p PaymentMethod) String() string { return string(p) }
