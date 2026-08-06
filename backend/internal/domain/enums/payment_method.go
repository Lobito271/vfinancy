package enums

// PaymentMethod is how a payment was (or will be) settled.
//
// Used by both supplier payments and customer payments. Settlement
// instruments (bank account, credit card, cash register) are stored
// separately as FKs on the payment entity, not as enum values.
type PaymentMethod string

const (
	PaymentMethodCash           PaymentMethod = "cash"
	PaymentMethodBankTransfer   PaymentMethod = "bank_transfer"
	PaymentMethodCheck          PaymentMethod = "check"
	PaymentMethodCard           PaymentMethod = "card"
	PaymentMethodCredit         PaymentMethod = "credit"
	PaymentMethodOther          PaymentMethod = "other"
)

func AllPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PaymentMethodCash, PaymentMethodBankTransfer, PaymentMethodCheck,
		PaymentMethodCard, PaymentMethodCredit, PaymentMethodOther,
	}
}

func (p PaymentMethod) Valid() bool {
	switch p {
	case PaymentMethodCash, PaymentMethodBankTransfer, PaymentMethodCheck,
		PaymentMethodCard, PaymentMethodCredit, PaymentMethodOther:
		return true
	}
	return false
}

func (p PaymentMethod) String() string { return string(p) }
