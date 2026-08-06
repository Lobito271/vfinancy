package sales

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CustomerAdvance is a payment received from a customer without a
// backing invoice. The customer may apply the advance to one or more
// future sales via CustomerAdvanceApplication.
type CustomerAdvance struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	CustomerID    uuid.UUID
	Number        string
	AdvanceDate   valueobjects.Date
	Amount        valueobjects.Money
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	Method        string
	BankAccountID *uuid.UUID
	applications  []AdvanceApplication
	Status        string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     *uuid.UUID
	UpdatedBy     *uuid.UUID
}

// AdvanceApplication is an internal record of an advance applied to
// a specific sale.
type AdvanceApplication struct {
	SaleID uuid.UUID
	Amount valueobjects.Money
}

// NewCustomerAdvanceOptions is the input to NewCustomerAdvance.
type NewCustomerAdvanceOptions struct {
	CompanyID     uuid.UUID
	CustomerID    uuid.UUID
	Number        string
	AdvanceDate   valueobjects.Date
	Amount        valueobjects.Money
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	Method        string
	BankAccountID *uuid.UUID
	Notes         string
}

// NewCustomerAdvance validates and constructs a customer advance.
func NewCustomerAdvance(now time.Time, opts NewCustomerAdvanceOptions) (*CustomerAdvance, error) {
	if opts.CompanyID == uuid.Nil || opts.CustomerID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company and customer are required"))
	}
	if !opts.Amount.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("advance amount must be positive"))
	}
	return &CustomerAdvance{
		ID:             uuid.New(),
		CompanyID:      opts.CompanyID,
		CustomerID:     opts.CustomerID,
		Number:         opts.Number,
		AdvanceDate:    opts.AdvanceDate,
		Amount:         opts.Amount,
		CurrencyCode:   opts.CurrencyCode,
		ExchangeRate:   opts.ExchangeRate,
		Method:         opts.Method,
		BankAccountID:  opts.BankAccountID,
		applications:   []AdvanceApplication{},
		Status:         "active",
		Notes:          opts.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// ApplyToSale consumes part of the advance against a sale. Returns the
// new remaining balance.
func (a *CustomerAdvance) ApplyToSale(saleID uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	if !amount.IsPositive() {
		return a.Remaining(), derrors.Wrap(derrors.ErrInvalidPayment, errField("application amount must be positive"))
	}
	if amount.GreaterThan(a.Remaining()) {
		return a.Remaining(), derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("application exceeds remaining advance"))
	}
	a.applications = append(a.applications, AdvanceApplication{SaleID: saleID, Amount: amount})
	return a.Remaining(), nil
}

// Applied returns the total amount already applied to sales.
func (a *CustomerAdvance) Applied() valueobjects.Money {
	sum := valueobjects.Zero()
	for _, x := range a.applications {
		sum = sum.Add(x.Amount)
	}
	return sum
}

// Remaining returns the amount that is still available for future
// applications.
func (a *CustomerAdvance) Remaining() valueobjects.Money {
	r := a.Amount.Sub(a.Applied())
	if r.IsNegative() {
		return valueobjects.Zero()
	}
	return r
}

// CanCoverSale reports whether the remaining advance is enough to
// cover the given sale total.
func (a *CustomerAdvance) CanCoverSale(total valueobjects.Money) bool {
	return total.LessOrEqual(a.Remaining()) && total.IsPositive()
}
