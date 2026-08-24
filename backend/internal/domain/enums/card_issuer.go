package enums

// CardIssuer is the brand of a company-issued credit card. Each issuer
// carries its own billing cycle (cut-off and payment-due day), used for
// the card billing projections on import orders.
type CardIssuer string

const (
	// CardIssuerVisa — Visa corporate card. Cut-off on the 25th, payment
	// due on the 20th of the following month.
	CardIssuerVisa CardIssuer = "visa"
	// CardIssuerDiners — Diners corporate card. Cut-off on the 3rd,
	// payment due on the 20th of the following month.
	CardIssuerDiners CardIssuer = "diners"
)

// AllCardIssuers returns every supported issuer.
func AllCardIssuers() []CardIssuer {
	return []CardIssuer{CardIssuerVisa, CardIssuerDiners}
}

// Valid reports whether the issuer is supported by the company.
func (c CardIssuer) Valid() bool {
	switch c {
	case CardIssuerVisa, CardIssuerDiners:
		return true
	}
	return false
}

// String returns the canonical (database) form of the issuer.
func (c CardIssuer) String() string { return string(c) }

// Label returns the display name of the issuer.
func (c CardIssuer) Label() string {
	switch c {
	case CardIssuerVisa:
		return "Visa"
	case CardIssuerDiners:
		return "Diners"
	}
	return string(c)
}

// CutOffDay returns the monthly billing cut-off day of the issuer.
func (c CardIssuer) CutOffDay() int {
	switch c {
	case CardIssuerVisa:
		return 25
	case CardIssuerDiners:
		return 3
	}
	return 0
}

// PaymentDueDay returns the monthly payment-due day of the issuer.
func (c CardIssuer) PaymentDueDay() int {
	switch c {
	case CardIssuerVisa, CardIssuerDiners:
		return 20
	}
	return 0
}
