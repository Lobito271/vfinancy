package customer

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Customer is a buyer of the company's products. Customers carry
// credit limits, debt, and a status that gates operations.
type Customer struct {
	ID              uuid.UUID
	CompanyID       uuid.UUID
	Document        valueobjects.DocumentNumber
	BusinessName    valueobjects.FullName
	TradeName       valueobjects.FullName
	TaxCategory     enums.TaxCategory
	CreditLimit     valueobjects.Money
	CurrentDebt     valueobjects.Money
	PaymentTermDays int
	Status          enums.CustomerStatus
	BlockedReason   string
	Email           valueobjects.Email
	Phone           valueobjects.Phone
	Address         valueobjects.Address
	BranchID        *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	CreatedBy       *uuid.UUID
	UpdatedBy       *uuid.UUID
}

// NewCustomerOptions is the input to NewCustomer.
type NewCustomerOptions struct {
	CompanyID       uuid.UUID
	Document        valueobjects.DocumentNumber
	BusinessName    valueobjects.FullName
	TradeName       valueobjects.FullName  // optional
	TaxCategory     enums.TaxCategory
	CreditLimit     valueobjects.Money
	PaymentTermDays int
	Email           valueobjects.Email
	Phone           valueobjects.Phone
	Address         valueobjects.Address
	BranchID        *uuid.UUID
}

// NewCustomer validates inputs and constructs a Customer. The customer
// starts active with zero debt.
func NewCustomer(now time.Time, opts NewCustomerOptions) (*Customer, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if !opts.TaxCategory.Valid() {
		return nil, errors.Wrap(errors.ErrInvalidEnum, errField("tax category is invalid"))
	}
	if opts.CreditLimit.IsNegative() {
		return nil, errors.Wrap(errors.ErrNegativeMoney, errField("credit limit cannot be negative"))
	}
	if opts.PaymentTermDays < 0 {
		return nil, errors.Wrap(errors.ErrOutOfRange, errField("payment term days cannot be negative"))
	}
	return &Customer{
		ID:              uuid.New(),
		CompanyID:       opts.CompanyID,
		Document:        opts.Document,
		BusinessName:    opts.BusinessName,
		TradeName:       opts.TradeName,
		TaxCategory:     opts.TaxCategory,
		CreditLimit:     opts.CreditLimit,
		CurrentDebt:     valueobjects.Zero(),
		PaymentTermDays: opts.PaymentTermDays,
		Status:          enums.CustomerStatusActive,
		Email:           opts.Email,
		Phone:           opts.Phone,
		Address:         opts.Address,
		BranchID:        opts.BranchID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Activate sets the customer status to active.
func (c *Customer) Activate() { c.Status = enums.CustomerStatusActive }

// Deactivate sets the customer status to inactive. The customer
// cannot be the target of new sales, but historical data is preserved.
func (c *Customer) Deactivate() { c.Status = enums.CustomerStatusInactive }

// Block sets the customer status to blocked with a reason. Blocked
// customers cannot place new sales.
func (c *Customer) Block(reason string) {
	c.Status = enums.CustomerStatusBlocked
	c.BlockedReason = reason
}

// Unblock clears the blocked status, returning the customer to active.
func (c *Customer) Unblock() {
	c.Status = enums.CustomerStatusActive
	c.BlockedReason = ""
}

// UpdateCreditLimit changes the credit limit. The new limit must be
// non-negative and not below the current debt (otherwise the customer
// would be over-limit immediately, which is allowed but flagged).
func (c *Customer) UpdateCreditLimit(limit valueobjects.Money) error {
	if limit.IsNegative() {
		return errors.Wrap(errors.ErrNegativeMoney, errField("credit limit cannot be negative"))
	}
	c.CreditLimit = limit
	return nil
}

// RecordSale adds a new sale amount to the customer's current debt.
// Returns the resulting debt balance.
func (c *Customer) RecordSale(amount valueobjects.Money) valueobjects.Money {
	c.CurrentDebt = c.CurrentDebt.Add(amount)
	return c.CurrentDebt
}

// RecordPayment reduces the customer's debt by a payment amount.
// Returns the resulting debt balance.
func (c *Customer) RecordPayment(amount valueobjects.Money) (valueobjects.Money, error) {
	if amount.IsNegative() || amount.IsZero() {
		return c.CurrentDebt, errors.Wrap(errors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	c.CurrentDebt = c.CurrentDebt.Sub(amount)
	if c.CurrentDebt.IsNegative() {
		c.CurrentDebt = valueobjects.Zero()
	}
	return c.CurrentDebt, nil
}

// AvailableCredit returns credit limit minus current debt. Never
// negative. Returns zero when no limit is set (CreditLimit == 0).
func (c *Customer) AvailableCredit() valueobjects.Money {
	if c.CreditLimit.IsZero() {
		return valueobjects.Zero()
	}
	avail := c.CreditLimit.Sub(c.CurrentDebt)
	if avail.IsNegative() {
		return valueobjects.Zero()
	}
	return avail
}

// IsOverLimit reports whether current debt exceeds the credit limit.
// Always false when no limit is set (CreditLimit == 0).
func (c *Customer) IsOverLimit() bool {
	if c.CreditLimit.IsZero() {
		return false
	}
	return c.CurrentDebt.GreaterThan(c.CreditLimit)
}

// CanPlaceSale reports whether a new sale of the given amount can be
// placed against this customer. A sale is allowed if the customer is
// active, not blocked, and the resulting debt does not exceed the
// credit limit.
func (c *Customer) CanPlaceSale(amount valueobjects.Money) error {
	if c.Status == enums.CustomerStatusBlocked {
		return errors.Wrap(errors.ErrCustomerInactive, errField("customer is blocked: "+c.BlockedReason))
	}
	if c.Status == enums.CustomerStatusInactive {
		return errors.Wrap(errors.ErrCustomerInactive, errField("customer is inactive"))
	}
	if amount.IsNegative() || amount.IsZero() {
		return errors.Wrap(errors.ErrInvalidPayment, errField("sale amount must be positive"))
	}
	if c.CreditLimit.IsZero() {
		return nil
	}
	resulting := c.CurrentDebt.Add(amount)
	if resulting.GreaterThan(c.CreditLimit) {
		return errors.Wrap(
			errors.ErrPaymentExceedsBalance,
			errField("sale would exceed credit limit"),
		)
	}
	return nil
}

// UpdateContact updates the email / phone / address.
func (c *Customer) UpdateContact(email valueobjects.Email, phone valueobjects.Phone, address valueobjects.Address) {
	c.Email = email
	c.Phone = phone
	c.Address = address
}

// ChangePaymentTerms updates the net-N days term. Negative is rejected.
func (c *Customer) ChangePaymentTerms(days int) error {
	if days < 0 {
		return errors.Wrap(errors.ErrOutOfRange, errField("payment term days cannot be negative"))
	}
	c.PaymentTermDays = days
	return nil
}
