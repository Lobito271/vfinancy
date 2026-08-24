package purchasing

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CustomerOrderPayment is a partial down payment (anticipo) received
// from a customer against a customer-type purchase order. Multiple
// payments can be recorded per order; the order's Anticipo field is the
// running total. A payment is refunded automatically when the order
// arrives faulty and is voided.
type CustomerOrderPayment struct {
	ID              uuid.UUID
	CompanyID       uuid.UUID
	PurchaseOrderID uuid.UUID
	CustomerID      uuid.UUID
	Number          string
	PaymentDate     valueobjects.Date
	Amount          valueobjects.Money
	Method          enums.PaymentMethod
	CurrencyCode    valueobjects.CurrencyCode
	ExchangeRate    valueobjects.ExchangeRate
	Reference       string
	Notes           string
	Status          string // "active" | "refunded"
	RefundedAmount  valueobjects.Money
	RefundedAt      *time.Time
	RefundReason    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       *uuid.UUID
	UpdatedBy       *uuid.UUID
}

// NewCustomerOrderPaymentOptions is the input to NewCustomerOrderPayment.
type NewCustomerOrderPaymentOptions struct {
	CompanyID       uuid.UUID
	PurchaseOrderID uuid.UUID
	CustomerID      uuid.UUID
	Number          string
	PaymentDate     valueobjects.Date
	Amount          valueobjects.Money
	Method          enums.PaymentMethod
	CurrencyCode    valueobjects.CurrencyCode
	ExchangeRate    valueobjects.ExchangeRate
	Reference       string
	Notes           string
}

// NewCustomerOrderPayment validates and constructs a down payment.
func NewCustomerOrderPayment(now time.Time, opts NewCustomerOrderPaymentOptions) (*CustomerOrderPayment, error) {
	if opts.CompanyID == uuid.Nil || opts.PurchaseOrderID == uuid.Nil || opts.CustomerID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company, order and customer are required"))
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
	return &CustomerOrderPayment{
		ID:              uuid.New(),
		CompanyID:       opts.CompanyID,
		PurchaseOrderID: opts.PurchaseOrderID,
		CustomerID:      opts.CustomerID,
		Number:          opts.Number,
		PaymentDate:     opts.PaymentDate,
		Amount:          opts.Amount,
		Method:          opts.Method,
		CurrencyCode:    opts.CurrencyCode,
		ExchangeRate:    opts.ExchangeRate,
		Reference:       opts.Reference,
		Notes:           opts.Notes,
		Status:          "active",
		RefundedAmount:  valueobjects.Zero(),
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// MarkRefunded sets the payment as fully refunded.
func (p *CustomerOrderPayment) MarkRefunded(at time.Time, reason string) error {
	if p.Status != "active" {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("payment is not active"))
	}
	if reason == "" {
		return derrors.Wrap(derrors.ErrRequired, errField("refund reason is required"))
	}
	p.Status = "refunded"
	p.RefundedAmount = p.Amount
	p.RefundedAt = &at
	p.RefundReason = reason
	p.UpdatedAt = at
	return nil
}

// IsRefunded reports whether the payment has been refunded.
func (p *CustomerOrderPayment) IsRefunded() bool { return p.Status == "refunded" }
