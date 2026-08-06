package treasury

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CreditCard is a company-issued credit card. The current_balance
// represents the outstanding debt (positive = money owed to the issuer).
type CreditCard struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	Issuer        string
	LastFour      string
	CardHolder    string
	ExpirationMonth int
	ExpirationYear  int
	CreditLimit   valueobjects.Money
	CurrentBalance valueobjects.Money
	CutOffDay     int
	PaymentDueDay int
	CurrencyCode  valueobjects.CurrencyCode
	GLAccountID   uuid.UUID
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	CreatedBy     *uuid.UUID
	UpdatedBy     *uuid.UUID
}

// NewCreditCardOptions is the input to NewCreditCard.
type NewCreditCardOptions struct {
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	Issuer        string
	LastFour      string
	CardHolder    string
	ExpirationMonth int
	ExpirationYear  int
	CreditLimit   valueobjects.Money
	CutOffDay     int
	PaymentDueDay int
	CurrencyCode  valueobjects.CurrencyCode
	GLAccountID   uuid.UUID
}

// NewCreditCard validates and constructs a credit card.
func NewCreditCard(now time.Time, opts NewCreditCardOptions) (*CreditCard, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company id is required"))
	}
	if opts.Issuer == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("issuer is required"))
	}
	if len(opts.LastFour) != 4 {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("last four must be 4 digits"))
	}
	if opts.ExpirationMonth < 1 || opts.ExpirationMonth > 12 {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("expiration month must be 1..12"))
	}
	if opts.ExpirationYear < 2000 || opts.ExpirationYear > 2100 {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("expiration year out of range"))
	}
	if opts.CreditLimit.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("credit limit cannot be negative"))
	}
	if opts.CutOffDay < 1 || opts.CutOffDay > 31 {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("cut-off day must be 1..31"))
	}
	if opts.PaymentDueDay < 1 || opts.PaymentDueDay > 31 {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("payment-due day must be 1..31"))
	}
	if opts.GLAccountID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("gl account is required"))
	}
	return &CreditCard{
		ID:              uuid.New(),
		CompanyID:       opts.CompanyID,
		BranchID:        opts.BranchID,
		Issuer:          opts.Issuer,
		LastFour:        opts.LastFour,
		CardHolder:      opts.CardHolder,
		ExpirationMonth: opts.ExpirationMonth,
		ExpirationYear:  opts.ExpirationYear,
		CreditLimit:     opts.CreditLimit,
		CurrentBalance:  valueobjects.Zero(),
		CutOffDay:       opts.CutOffDay,
		PaymentDueDay:   opts.PaymentDueDay,
		CurrencyCode:    opts.CurrencyCode,
		GLAccountID:     opts.GLAccountID,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Charge adds a purchase amount to the card balance. Rejects charges
// that would exceed the credit limit.
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

// Pay reduces the card balance by a payment amount.
func (c *CreditCard) Pay(amount valueobjects.Money) error {
	if !amount.IsPositive() {
		return derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	if amount.GreaterThan(c.CurrentBalance) {
		return derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("payment exceeds outstanding balance"))
	}
	c.CurrentBalance = c.CurrentBalance.Sub(amount)
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
