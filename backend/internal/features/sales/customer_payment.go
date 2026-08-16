package sales

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CustomerPayment is a payment received from a customer. It can be
// applied to one or more sales via CustomerPaymentAllocation.
type CustomerPayment struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	CustomerID    uuid.UUID
	Number        string
	PaymentDate   valueobjects.Date
	Amount        valueobjects.Money
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	Method        enums.PaymentMethod
	BankAccountID *uuid.UUID
	CashRegisterID *uuid.UUID
	Reference     string
	allocations   []PaymentAllocation
	Status        string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     *uuid.UUID
	UpdatedBy     *uuid.UUID
}

// PaymentAllocation is an internal record of how much of a payment
// has been applied to a specific sale. The aggregate owns the
// allocations and enforces that their sum never exceeds the payment.
type PaymentAllocation struct {
	SaleID uuid.UUID
	Amount valueobjects.Money
}

// NewCustomerPaymentOptions is the input to NewCustomerPayment.
type NewCustomerPaymentOptions struct {
	CompanyID     uuid.UUID
	CustomerID    uuid.UUID
	Number        string
	PaymentDate   valueobjects.Date
	Amount        valueobjects.Money
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	Method        enums.PaymentMethod
	BankAccountID *uuid.UUID
	CashRegisterID *uuid.UUID
	Reference     string
	Notes         string
}

// NewCustomerPayment validates and constructs a payment.
func NewCustomerPayment(now time.Time, opts NewCustomerPaymentOptions) (*CustomerPayment, error) {
	if opts.CompanyID == uuid.Nil || opts.CustomerID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company and customer are required"))
	}
	if opts.Number == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("payment number is required"))
	}
	if !opts.Amount.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	if !opts.Method.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("payment method is invalid"))
	}
	return &CustomerPayment{
		ID:             uuid.New(),
		CompanyID:      opts.CompanyID,
		CustomerID:     opts.CustomerID,
		Number:         opts.Number,
		PaymentDate:    opts.PaymentDate,
		Amount:         opts.Amount,
		CurrencyCode:   opts.CurrencyCode,
		ExchangeRate:   opts.ExchangeRate,
		Method:         opts.Method,
		BankAccountID:  opts.BankAccountID,
		CashRegisterID: opts.CashRegisterID,
		Reference:      opts.Reference,
		allocations:    []PaymentAllocation{},
		Status:         "active",
		Notes:          opts.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// ApplyToSale records an allocation. The payment is allowed to be
// applied to multiple sales as long as the total allocated does not
// exceed the payment amount.
func (p *CustomerPayment) ApplyToSale(saleID uuid.UUID, amount valueobjects.Money) error {
	if !amount.IsPositive() {
		return derrors.Wrap(derrors.ErrInvalidPayment, errField("allocation amount must be positive"))
	}
	allocated := valueobjects.Zero()
	for _, a := range p.allocations {
		allocated = allocated.Add(a.Amount)
	}
	if allocated.Add(amount).GreaterThan(p.Amount) {
		return derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("allocation would exceed payment amount"))
	}
	p.allocations = append(p.allocations, PaymentAllocation{SaleID: saleID, Amount: amount})
	return nil
}

// AllocatedAmount returns the sum of all allocations.
func (p *CustomerPayment) AllocatedAmount() valueobjects.Money {
	sum := valueobjects.Zero()
	for _, a := range p.allocations {
		sum = sum.Add(a.Amount)
	}
	return sum
}

// Allocations returns a copy of the payment's allocations.
func (p *CustomerPayment) Allocations() []PaymentAllocation {
	out := make([]PaymentAllocation, len(p.allocations))
	copy(out, p.allocations)
	return out
}

// UnallocatedAmount returns the portion of the payment not yet
// allocated to any sale.
func (p *CustomerPayment) UnallocatedAmount() valueobjects.Money {
	u := p.Amount.Sub(p.AllocatedAmount())
	if u.IsNegative() {
		return valueobjects.Zero()
	}
	return u
}
