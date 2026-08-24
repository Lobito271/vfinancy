package sales

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Sale is the root aggregate for a customer invoice. It owns its
// line items, tracks the running paid amount, and enforces state
// transitions.
type Sale struct {
	ID              uuid.UUID
	CompanyID       uuid.UUID
	BranchID        *uuid.UUID
	Number          string
	CustomerID      uuid.UUID
	CurrencyCode    valueobjects.CurrencyCode
	ExchangeRate    valueobjects.ExchangeRate
	Items           []*SaleItem
	Status          enums.SaleStatus
	Subtotal        valueobjects.Money
	DiscountAmount  valueobjects.Money
	TaxAmount       valueobjects.Money
	Total           valueobjects.Money
	CostTotal       valueobjects.Money
	Profit          valueobjects.Money
	Paid            valueobjects.Money
	DueDate         *valueobjects.Date
	Notes           string
	SaleDate        time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CancelledAt     *time.Time
	CancelledReason string
	CreatedBy       *uuid.UUID
	UpdatedBy       *uuid.UUID
}

// NewSaleOptions is the input to NewSale.
type NewSaleOptions struct {
	CompanyID    uuid.UUID
	BranchID     *uuid.UUID
	Number       string
	CustomerID   uuid.UUID
	CurrencyCode valueobjects.CurrencyCode
	ExchangeRate valueobjects.ExchangeRate
	SaleDate     time.Time
	DueDate      *valueobjects.Date
	Notes        string
}

// NewSale creates a new sale in pending status with no line items.
// The application layer is expected to add lines via AddItem and then
// either call Recalculate() to refresh totals or rely on it being
// called by the persistence layer on save.
func NewSale(now time.Time, opts NewSaleOptions) (*Sale, error) {
	if opts.CompanyID == uuid.Nil || opts.CustomerID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company and customer are required"))
	}
	if opts.Number == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("sale number is required"))
	}
	return &Sale{
		ID:             uuid.New(),
		CompanyID:      opts.CompanyID,
		BranchID:       opts.BranchID,
		Number:         opts.Number,
		CustomerID:     opts.CustomerID,
		CurrencyCode:   opts.CurrencyCode,
		ExchangeRate:   opts.ExchangeRate,
		Items:          []*SaleItem{},
		Status:         enums.SaleStatusPending,
		Subtotal:       valueobjects.Zero(),
		DiscountAmount: valueobjects.Zero(),
		TaxAmount:      valueobjects.Zero(),
		Total:          valueobjects.Zero(),
		CostTotal:      valueobjects.Zero(),
		Profit:         valueobjects.Zero(),
		Paid:           valueobjects.Zero(),
		DueDate:        opts.DueDate,
		Notes:          opts.Notes,
		SaleDate:       opts.SaleDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// AddItem appends a line. The aggregate is recalculated immediately.
func (s *Sale) AddItem(item *SaleItem) error {
	if s.Status == enums.SaleStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot add items to a cancelled sale"))
	}
	if s.Status == enums.SaleStatusPaid {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot add items to a paid sale"))
	}
	for _, existing := range s.Items {
		if existing.ProductID == item.ProductID {
			return derrors.Wrap(derrors.ErrDuplicateLine, errField("sale already has a line for this product"))
		}
	}
	item.SaleID = s.ID
	item.LineNumber = len(s.Items) + 1
	s.Items = append(s.Items, item)
	return s.Recalculate()
}

// RemoveItem drops a line by ID. The aggregate is recalculated.
func (s *Sale) RemoveItem(itemID uuid.UUID) error {
	if s.Status == enums.SaleStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot remove items from a cancelled sale"))
	}
	for i, li := range s.Items {
		if li.ID == itemID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			// renumber
			for j, ln := range s.Items {
				ln.LineNumber = j + 1
			}
			return s.Recalculate()
		}
	}
	return derrors.Wrap(derrors.ErrRequired, errField("line not found"))
}

// Recalculate refreshes subtotal, discount, tax, total, cost and
// profit. Called automatically by AddItem/RemoveItem and after any
// change to a line.
func (s *Sale) Recalculate() error {
	if len(s.Items) == 0 {
		s.Subtotal = valueobjects.Zero()
		s.DiscountAmount = valueobjects.Zero()
		s.TaxAmount = valueobjects.Zero()
		s.Total = valueobjects.Zero()
		s.CostTotal = valueobjects.Zero()
		s.Profit = valueobjects.Zero()
		return nil
	}
	sub := valueobjects.Zero()
	dis := valueobjects.Zero()
	tax := valueobjects.Zero()
	cost := valueobjects.Zero()
	for _, li := range s.Items {
		sub = sub.Add(li.LineSubtotal())
		dis = dis.Add(li.DiscountAmount)
		tax = tax.Add(li.TaxAmount)
		cost = cost.Add(li.LineCost())
	}
	s.Subtotal = sub
	s.DiscountAmount = dis
	s.TaxAmount = tax
	s.Total = sub.Sub(dis).Add(tax)
	s.CostTotal = cost
	s.Profit = s.Total.Sub(tax).Sub(cost)
	return nil
}

// CalculateSubtotal is a convenience accessor (read-only).
func (s *Sale) CalculateSubtotal() valueobjects.Money { return s.Subtotal }

// CalculateTaxes returns the total tax amount.
func (s *Sale) CalculateTaxes() valueobjects.Money { return s.TaxAmount }

// CalculateTotal returns the document total.
func (s *Sale) CalculateTotal() valueobjects.Money { return s.Total }

// CalculateProfit returns the realized gross margin.
func (s *Sale) CalculateProfit() valueobjects.Money { return s.Profit }

// Balance returns the outstanding amount (Total - Paid).
func (s *Sale) Balance() valueobjects.Money {
	b := s.Total.Sub(s.Paid)
	if b.IsNegative() {
		return valueobjects.Zero()
	}
	return b
}

// ApplyPayment records a payment against the sale. Returns the new
// balance. Transitions the status automatically.
func (s *Sale) ApplyPayment(amount valueobjects.Money) (valueobjects.Money, error) {
	if s.Status == enums.SaleStatusCancelled {
		return s.Balance(), derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot pay a cancelled sale"))
	}
	if s.Status == enums.SaleStatusPaid {
		return s.Balance(), derrors.Wrap(derrors.ErrSaleAlreadyPaid, errField("sale is already fully paid"))
	}
	if !amount.IsPositive() {
		return s.Balance(), derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	s.Paid = s.Paid.Add(amount)
	if s.Paid.GreaterThan(s.Total) {
		// We allow overpayment only via an explicit credit. For now
		// reject and let the caller decide what to do.
		s.Paid = s.Paid.Sub(amount)
		return s.Balance(), derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("payment exceeds sale total"))
	}
	s.transitionAfterPayment()
	return s.Balance(), nil
}

// transitionAfterPayment sets the status based on the new paid amount.
// Called from ApplyPayment.
func (s *Sale) transitionAfterPayment() {
	switch {
	case s.Paid.Equals(s.Total) && s.Total.IsPositive():
		s.Status = enums.SaleStatusPaid
	case s.Paid.IsPositive() && s.Paid.LessThan(s.Total):
		s.Status = enums.SaleStatusPartial
	default:
		s.Status = enums.SaleStatusPending
	}
}

// MarkAsPaid records the full balance as paid. Used when the entire
// sale is settled in cash or against an advance.
func (s *Sale) MarkAsPaid() error {
	if s.Status == enums.SaleStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot mark a cancelled sale as paid"))
	}
	if s.Total.IsZero() {
		return derrors.Wrap(derrors.ErrEmptyDocument, errField("cannot mark a zero-total sale as paid"))
	}
	s.Paid = s.Total
	s.Status = enums.SaleStatusPaid
	return nil
}

// MarkAsPartiallyPaid is a manual override for cases where payments
// were applied outside the system. The amount must be positive and
// strictly less than the total.
func (s *Sale) MarkAsPartiallyPaid(amount valueobjects.Money) error {
	if s.Status == enums.SaleStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot pay a cancelled sale"))
	}
	if !amount.IsPositive() || amount.GreaterOrEqual(s.Total) {
		return derrors.Wrap(derrors.ErrInvalidPayment, errField("amount must be positive and strictly less than total"))
	}
	s.Paid = amount
	s.Status = enums.SaleStatusPartial
	return nil
}

// Cancel marks the sale as cancelled. The application layer is
// responsible for emitting compensating inventory_movements and
// reversing the related journal_entry.
func (s *Sale) Cancel(reason string) error {
	if s.Status == enums.SaleStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("sale is already cancelled"))
	}
	now := time.Now().UTC()
	s.Status = enums.SaleStatusCancelled
	s.CancelledAt = &now
	s.CancelledReason = reason
	return nil
}

// IsPaid / IsCancelled / IsPending are convenience predicates.
func (s *Sale) IsPaid() bool       { return s.Status == enums.SaleStatusPaid }
func (s *Sale) IsCancelled() bool  { return s.Status == enums.SaleStatusCancelled }
func (s *Sale) IsPending() bool    { return s.Status == enums.SaleStatusPending }
func (s *Sale) IsPartiallyPaid() bool { return s.Status == enums.SaleStatusPartial }

// CanReceivePayments reports whether the sale is in a state that
// accepts new payments.
func (s *Sale) CanReceivePayments() bool {
	return s.Status == enums.SaleStatusPending || s.Status == enums.SaleStatusPartial
}
