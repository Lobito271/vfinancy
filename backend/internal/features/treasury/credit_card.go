package treasury

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CreditCard is a company-issued credit card. Cards are USD by
// definition; the current_balance represents the outstanding debt
// (positive = money owed to the issuer).
type CreditCard struct {
	ID             uuid.UUID
	Issuer         string
	LastFour       string
	CreditLimit    valueobjects.Money
	CurrentBalance valueobjects.Money
	CutOffDay      int
	PaymentDueDay  int
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	CreatedBy      *uuid.UUID
	UpdatedBy      *uuid.UUID
}

// NewCreditCardOptions is the input to NewCreditCard.
type NewCreditCardOptions struct {
	Issuer        string
	LastFour      string
	CreditLimit   valueobjects.Money
	CutOffDay     int
	PaymentDueDay int
}

// Validate enforces the card invariants: non-blank issuer, exactly 4
// digits as last_four, cut-off and payment-due days 1..31 and a
// non-negative credit limit.
func (c *CreditCard) Validate() error {
	if len(c.Issuer) > 100 || isBlank(c.Issuer) {
		return derrors.Wrap(derrors.ErrRequired, errField("issuer is required"))
	}
	if len(c.LastFour) != 4 || !allDigits(c.LastFour) {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("last four must be 4 digits"))
	}
	if c.CutOffDay < 1 || c.CutOffDay > 31 {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("cut-off day must be 1..31"))
	}
	if c.PaymentDueDay < 1 || c.PaymentDueDay > 31 {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("payment-due day must be 1..31"))
	}
	if c.CreditLimit.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("credit limit cannot be negative"))
	}
	return nil
}

// allDigits reports whether s consists of exactly four arabic digits.
// Empty strings fail because len is checked by the caller.
func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// NewCreditCard validates and constructs a credit card.
func NewCreditCard(now time.Time, opts NewCreditCardOptions) (*CreditCard, error) {
	if opts.Issuer == "" || isBlank(opts.Issuer) {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("issuer is required"))
	}
	if len(opts.LastFour) != 4 || !allDigits(opts.LastFour) {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("last four must be 4 digits"))
	}
	card := &CreditCard{
		ID:             uuid.New(),
		Issuer:         opts.Issuer,
		LastFour:       opts.LastFour,
		CreditLimit:    opts.CreditLimit,
		CurrentBalance: valueobjects.Zero(),
		CutOffDay:      opts.CutOffDay,
		PaymentDueDay:  opts.PaymentDueDay,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := card.Validate(); err != nil {
		return nil, err
	}
	return card, nil
}

// isBlank reports whether s is empty or whitespace only.
func isBlank(s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
			return false
		}
	}
	return true
}

// Charge adds a purchase amount to the card balance. Rejects charges
// that are not positive or that would exceed the credit limit.
func (c *CreditCard) Charge(amount valueobjects.Money) error {
	if !amount.IsPositive() {
		return derrors.Wrap(derrors.ErrInvalidPayment, errField("charge amount must be positive"))
	}
	if c.CurrentBalance.Add(amount).GreaterThan(c.CreditLimit) {
		return derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("charge would exceed credit limit"))
	}
	c.CurrentBalance = c.CurrentBalance.Add(amount)
	return nil
}

// Release reverses a charge, floors at zero.
func (c *CreditCard) Release(amount valueobjects.Money) error {
	if !amount.IsPositive() {
		return derrors.Wrap(derrors.ErrInvalidPayment, errField("release amount must be positive"))
	}
	c.CurrentBalance = subFloorZero(c.CurrentBalance, amount)
	return nil
}

// Pay reduces the card balance by a payment amount. Overpayments
// liquidate the balance to zero instead of failing.
func (c *CreditCard) Pay(amount valueobjects.Money) error {
	if !amount.IsPositive() {
		return derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	c.CurrentBalance = subFloorZero(c.CurrentBalance, amount)
	return nil
}

// AvailableCredit returns credit_limit - current_balance. Never negative.
func (c *CreditCard) AvailableCredit() valueobjects.Money {
	avail := c.CreditLimit.Sub(c.CurrentBalance)
	if avail.IsNegative() {
		return valueobjects.Zero()
	}
	return avail
}

// subFloorZero returns a - b, clamped to 0 when b exceeds a.
func subFloorZero(a, b valueobjects.Money) valueobjects.Money {
	if b.GreaterThan(a) {
		return valueobjects.Zero()
	}
	return a.Sub(b)
}
