package sales

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Payment lifecycle statuses stored on customer_payments.status.
const (
	PaymentStatusActive   = "active"
	PaymentStatusRefunded = "refunded"
)

// CustomerPayment is a payment received from a customer, applied to
// one or more sales via PaymentAllocation.
type CustomerPayment struct {
	ID            uuid.UUID
	CustomerID    uuid.UUID
	Number        string
	PaymentDate   time.Time
	Amount        valueobjects.Money
	PaymentMethod enums.PaymentMethod
	Reference     string
	Notes         string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	CreatedBy     *uuid.UUID
	UpdatedBy     *uuid.UUID
}

// PaymentAllocation records how much of a payment was applied to a
// specific sale.
type PaymentAllocation struct {
	ID                uuid.UUID
	CustomerPaymentID uuid.UUID
	SaleID            uuid.UUID
	AllocatedAmount   valueobjects.Money
	CreatedAt         time.Time
}

// SaleCollection is the amount of an active payment allocated to a sale
// on a given date. Used by the dashboard to attribute collected cash
// (and the corresponding margin) to the period it was received.
type SaleCollection struct {
	SaleID      uuid.UUID
	PaymentDate time.Time
	Amount      valueobjects.Money
}

// NewCustomerPaymentOptions is the input to NewCustomerPayment.
type NewCustomerPaymentOptions struct {
	CustomerID    uuid.UUID
	Number        string
	PaymentDate   time.Time
	Amount        valueobjects.Money
	PaymentMethod enums.PaymentMethod
	Reference     string
	Notes         string
}

// NewCustomerPayment validates and constructs a payment in active
// status.
func NewCustomerPayment(now time.Time, opts NewCustomerPaymentOptions) (*CustomerPayment, error) {
	if opts.CustomerID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("customer is required"))
	}
	if opts.Number == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("payment number is required"))
	}
	if opts.PaymentDate.IsZero() {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("payment date is required"))
	}
	if !opts.Amount.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	if !opts.PaymentMethod.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("payment method is invalid"))
	}
	return &CustomerPayment{
		ID:            uuid.New(),
		CustomerID:    opts.CustomerID,
		Number:        opts.Number,
		PaymentDate:   opts.PaymentDate,
		Amount:        opts.Amount,
		PaymentMethod: opts.PaymentMethod,
		Reference:     opts.Reference,
		Status:        PaymentStatusActive,
		Notes:         opts.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}
