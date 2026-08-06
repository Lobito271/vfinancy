package purchasing

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// SupplierPayment is a payment made to a supplier. It can be
// allocated to one or more purchase orders via SupplierPaymentAllocation.
type SupplierPayment struct {
	ID             uuid.UUID
	CompanyID      uuid.UUID
	SupplierID     uuid.UUID
	Number         string
	PaymentDate    valueobjects.Date
	Amount         valueobjects.Money
	CurrencyCode   valueobjects.CurrencyCode
	ExchangeRate   valueobjects.ExchangeRate
	Method         enums.PaymentMethod
	BankAccountID  *uuid.UUID
	CashRegisterID *uuid.UUID
	CreditCardID   *uuid.UUID
	Reference      string
	allocations    []SupplierPaymentAllocation
	Status         string
	Notes          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedBy      *uuid.UUID
	UpdatedBy      *uuid.UUID
}

// SupplierPaymentAllocation maps part of a payment to a purchase.
type SupplierPaymentAllocation struct {
	PurchaseOrderID uuid.UUID
	Amount          valueobjects.Money
}

// NewSupplierPaymentOptions is the input to NewSupplierPayment.
type NewSupplierPaymentOptions struct {
	CompanyID      uuid.UUID
	SupplierID     uuid.UUID
	Number         string
	PaymentDate    valueobjects.Date
	Amount         valueobjects.Money
	CurrencyCode   valueobjects.CurrencyCode
	ExchangeRate   valueobjects.ExchangeRate
	Method         enums.PaymentMethod
	BankAccountID  *uuid.UUID
	CashRegisterID *uuid.UUID
	CreditCardID   *uuid.UUID
	Reference      string
	Notes          string
}

// NewSupplierPayment validates and constructs a supplier payment.
func NewSupplierPayment(now time.Time, opts NewSupplierPaymentOptions) (*SupplierPayment, error) {
	if opts.CompanyID == uuid.Nil || opts.SupplierID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company and supplier are required"))
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
	return &SupplierPayment{
		ID:             uuid.New(),
		CompanyID:      opts.CompanyID,
		SupplierID:     opts.SupplierID,
		Number:         opts.Number,
		PaymentDate:    opts.PaymentDate,
		Amount:         opts.Amount,
		CurrencyCode:   opts.CurrencyCode,
		ExchangeRate:   opts.ExchangeRate,
		Method:         opts.Method,
		BankAccountID:  opts.BankAccountID,
		CashRegisterID: opts.CashRegisterID,
		CreditCardID:   opts.CreditCardID,
		Reference:      opts.Reference,
		allocations:    []SupplierPaymentAllocation{},
		Status:         "active",
		Notes:          opts.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// ApplyToPurchase allocates part of the payment to a specific purchase.
func (p *SupplierPayment) ApplyToPurchase(purchaseID uuid.UUID, amount valueobjects.Money) error {
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
	p.allocations = append(p.allocations, SupplierPaymentAllocation{
		PurchaseOrderID: purchaseID,
		Amount:          amount,
	})
	return nil
}

// AllocatedAmount returns the sum of all allocations.
func (p *SupplierPayment) AllocatedAmount() valueobjects.Money {
	sum := valueobjects.Zero()
	for _, a := range p.allocations {
		sum = sum.Add(a.Amount)
	}
	return sum
}

// UnallocatedAmount returns the portion of the payment not yet
// allocated.
func (p *SupplierPayment) UnallocatedAmount() valueobjects.Money {
	u := p.Amount.Sub(p.AllocatedAmount())
	if u.IsNegative() {
		return valueobjects.Zero()
	}
	return u
}
