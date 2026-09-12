package sales

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Sale is the root aggregate for a customer invoice. Amounts are in
// PEN; the status tracks the paid amount against the total.
type Sale struct {
	ID              uuid.UUID
	CustomerID      uuid.UUID
	Number          string
	SaleDate        time.Time
	DueDate         *time.Time
	Status          enums.SaleStatus
	SaleType        enums.SaleType
	Total           valueobjects.Money
	PaidAmount      valueobjects.Money
	CostTotal       valueobjects.Money
	Profit          valueobjects.Money
	Notes           string
	CancelledAt     *time.Time
	CancelledReason string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	CreatedBy       *uuid.UUID
	UpdatedBy       *uuid.UUID
	Items           []*SaleItem
}

// NewSaleOptions is the input to NewSale.
type NewSaleOptions struct {
	Number     string
	CustomerID uuid.UUID
	SaleType   enums.SaleType
	SaleDate   time.Time
	DueDate    *time.Time
	Notes      string
}

// NewSale creates a sale in pending status with zero totals.
func NewSale(now time.Time, opts NewSaleOptions) (*Sale, error) {
	if opts.CustomerID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("customer is required"))
	}
	if opts.Number == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("sale number is required"))
	}
	if !opts.SaleType.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("sale type is invalid"))
	}
	return &Sale{
		ID:         uuid.New(),
		CustomerID: opts.CustomerID,
		Number:     opts.Number,
		SaleDate:   opts.SaleDate,
		DueDate:    opts.DueDate,
		Status:     enums.SaleStatusPending,
		SaleType:   opts.SaleType,
		Total:      valueobjects.Zero(),
		PaidAmount: valueobjects.Zero(),
		CostTotal:  valueobjects.Zero(),
		Profit:     valueobjects.Zero(),
		Notes:      opts.Notes,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Validate checks the date and amount invariants.
func (s *Sale) Validate() error {
	if s.SaleDate.IsZero() {
		return derrors.Wrap(derrors.ErrRequired, errField("sale date is required"))
	}
	if s.DueDate != nil && s.DueDate.Before(s.SaleDate) {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("due date cannot be before the sale date"))
	}
	if s.Total.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("total cannot be negative"))
	}
	return nil
}

// Outstanding returns the unpaid balance (total - paid), floored at
// zero.
func (s *Sale) Outstanding() valueobjects.Money {
	b := s.Total.Sub(s.PaidAmount)
	if b.IsNegative() {
		return valueobjects.Zero()
	}
	return b
}

// RegisterPayment adds a positive payment to the paid amount and
// transitions the status (pending / partial / paid). It returns the
// remaining outstanding balance.
func (s *Sale) RegisterPayment(amount valueobjects.Money) (valueobjects.Money, error) {
	if s.Status == enums.SaleStatusCancelled {
		return s.Outstanding(), derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot pay a cancelled sale"))
	}
	if !amount.IsPositive() {
		return s.Outstanding(), derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	newPaid := s.PaidAmount.Add(amount)
	if newPaid.GreaterThan(s.Total) {
		return s.Outstanding(), derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("payment exceeds sale total"))
	}
	s.PaidAmount = newPaid
	switch {
	case newPaid.Equals(s.Total) && s.Total.IsPositive():
		s.Status = enums.SaleStatusPaid
	case newPaid.IsPositive():
		s.Status = enums.SaleStatusPartial
	default:
		s.Status = enums.SaleStatusPending
	}
	s.UpdatedAt = time.Now().UTC()
	return s.Outstanding(), nil
}

// Cancel marks the sale as cancelled. Compensating actions (stock
// return, debt adjustment) are the service's responsibility.
func (s *Sale) Cancel(reason string) error {
	if s.Status == enums.SaleStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("sale is already cancelled"))
	}
	now := time.Now().UTC()
	s.Status = enums.SaleStatusCancelled
	s.CancelledAt = &now
	s.CancelledReason = reason
	s.UpdatedAt = now
	return nil
}
