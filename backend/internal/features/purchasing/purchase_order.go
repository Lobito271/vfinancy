package purchasing

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// PurchaseOrder is the root aggregate for a supplier purchase. It
// owns its line items, tracks the running paid amount, and enforces
// the purchase state machine.
type PurchaseOrder struct {
	ID             uuid.UUID
	CompanyID      uuid.UUID
	BranchID       *uuid.UUID
	Number         string
	SupplierID     uuid.UUID
	CurrencyCode   valueobjects.CurrencyCode
	ExchangeRate   valueobjects.ExchangeRate
	Items          []*PurchaseOrderItem
	Status         enums.PurchaseStatus
	Subtotal       valueobjects.Money
	DiscountAmount valueobjects.Money
	TaxAmount      valueobjects.Money
	Total          valueobjects.Money
	Paid           valueobjects.Money
	OrderDate      valueobjects.Date
	ExpectedDate   *valueobjects.Date
	ReceivedDate   *valueobjects.Date
	Notes          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CancelledAt    *time.Time
	CancelledReason string
	CreatedBy      *uuid.UUID
	UpdatedBy      *uuid.UUID
}

// NewPurchaseOrderOptions is the input to NewPurchaseOrder.
type NewPurchaseOrderOptions struct {
	CompanyID    uuid.UUID
	BranchID     *uuid.UUID
	Number       string
	SupplierID   uuid.UUID
	CurrencyCode valueobjects.CurrencyCode
	ExchangeRate valueobjects.ExchangeRate
	OrderDate    valueobjects.Date
	ExpectedDate *valueobjects.Date
	Notes        string
}

// NewPurchaseOrder creates a new purchase in pending status.
func NewPurchaseOrder(now time.Time, opts NewPurchaseOrderOptions) (*PurchaseOrder, error) {
	if opts.CompanyID == uuid.Nil || opts.SupplierID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company and supplier are required"))
	}
	if opts.Number == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("purchase number is required"))
	}
	return &PurchaseOrder{
		ID:             uuid.New(),
		CompanyID:      opts.CompanyID,
		BranchID:       opts.BranchID,
		Number:         opts.Number,
		SupplierID:     opts.SupplierID,
		CurrencyCode:   opts.CurrencyCode,
		ExchangeRate:   opts.ExchangeRate,
		Items:          []*PurchaseOrderItem{},
		Status:         enums.PurchaseStatusPending,
		Subtotal:       valueobjects.Zero(),
		DiscountAmount: valueobjects.Zero(),
		TaxAmount:      valueobjects.Zero(),
		Total:          valueobjects.Zero(),
		Paid:           valueobjects.Zero(),
		OrderDate:      opts.OrderDate,
		ExpectedDate:   opts.ExpectedDate,
		Notes:          opts.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// AddItem appends a line. The aggregate is recalculated.
func (p *PurchaseOrder) AddItem(item *PurchaseOrderItem) error {
	if p.Status == enums.PurchaseStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot add items to a cancelled purchase"))
	}
	if p.Status == enums.PurchaseStatusReconciled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot add items to a reconciled purchase"))
	}
	for _, existing := range p.Items {
		if existing.ProductID == item.ProductID {
			return derrors.Wrap(derrors.ErrDuplicateLine, errField("purchase already has a line for this product"))
		}
	}
	item.PurchaseOrderID = p.ID
	item.LineNumber = len(p.Items) + 1
	p.Items = append(p.Items, item)
	return p.Recalculate()
}

// RemoveItem drops a line and recalculates.
func (p *PurchaseOrder) RemoveItem(itemID uuid.UUID) error {
	if p.Status == enums.PurchaseStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot remove items from a cancelled purchase"))
	}
	for i, li := range p.Items {
		if li.ID == itemID {
			p.Items = append(p.Items[:i], p.Items[i+1:]...)
			for j, ln := range p.Items {
				ln.LineNumber = j + 1
			}
			return p.Recalculate()
		}
	}
	return derrors.Wrap(derrors.ErrRequired, errField("line not found"))
}

// Recalculate refreshes subtotal, discount, tax, total.
func (p *PurchaseOrder) Recalculate() error {
	if len(p.Items) == 0 {
		p.Subtotal = valueobjects.Zero()
		p.DiscountAmount = valueobjects.Zero()
		p.TaxAmount = valueobjects.Zero()
		p.Total = valueobjects.Zero()
		return nil
	}
	sub := valueobjects.Zero()
	dis := valueobjects.Zero()
	tax := valueobjects.Zero()
	for _, li := range p.Items {
		sub = sub.Add(li.LineSubtotal())
		dis = dis.Add(li.DiscountAmount)
		tax = tax.Add(li.TaxAmount)
	}
	p.Subtotal = sub
	p.DiscountAmount = dis
	p.TaxAmount = tax
	p.Total = sub.Sub(dis).Add(tax)
	return nil
}

// CalculateTotal returns the document total.
func (p *PurchaseOrder) CalculateTotal() valueobjects.Money { return p.Total }

// Balance returns the outstanding amount (Total - Paid).
func (p *PurchaseOrder) Balance() valueobjects.Money {
	b := p.Total.Sub(p.Paid)
	if b.IsNegative() {
		return valueobjects.Zero()
	}
	return b
}

// Approve transitions the purchase from pending to received (which
// indicates the goods are expected to arrive). The application layer
// records the actual receipt and triggers the inventory_movements.
func (p *PurchaseOrder) Approve() error {
	if p.Status == enums.PurchaseStatusCancelled {
		return derrors.Wrap(derrors.ErrPurchaseCancelled, errField("cannot approve a cancelled purchase"))
	}
	if p.Status == enums.PurchaseStatusReconciled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("purchase is already reconciled"))
	}
	if len(p.Items) == 0 {
		return derrors.Wrap(derrors.ErrEmptyDocument, errField("cannot approve an empty purchase"))
	}
	p.Status = enums.PurchaseStatusReceived
	return nil
}

// MarkAsReceived sets the receipt date and moves status to received.
// Called when the warehouse confirms the goods arrived.
func (p *PurchaseOrder) MarkAsReceived(at valueobjects.Date) error {
	if p.Status == enums.PurchaseStatusCancelled {
		return derrors.Wrap(derrors.ErrPurchaseCancelled, errField("cannot mark a cancelled purchase as received"))
	}
	if p.ReceivedDate != nil {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("purchase is already received"))
	}
	p.ReceivedDate = &at
	p.Status = enums.PurchaseStatusReceived
	return nil
}

// MarkAsPaid sets the entire balance as paid.
func (p *PurchaseOrder) MarkAsPaid() error {
	if p.Status == enums.PurchaseStatusCancelled {
		return derrors.Wrap(derrors.ErrPurchaseCancelled, errField("cannot pay a cancelled purchase"))
	}
	if p.Total.IsZero() {
		return derrors.Wrap(derrors.ErrEmptyDocument, errField("cannot pay a zero-total purchase"))
	}
	p.Paid = p.Total
	p.Status = enums.PurchaseStatusPaid
	return nil
}

// Reconcile marks the purchase as reconciled (paid + matched against
// the bank statement). Terminal state.
func (p *PurchaseOrder) Reconcile() error {
	if p.Status != enums.PurchaseStatusPaid {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("only paid purchases can be reconciled"))
	}
	p.Status = enums.PurchaseStatusReconciled
	return nil
}

// Cancel marks the purchase as cancelled. The application layer
// records compensating inventory_movements and reverses the journal
// entry.
func (p *PurchaseOrder) Cancel(reason string) error {
	if p.Status == enums.PurchaseStatusCancelled {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("purchase is already cancelled"))
	}
	now := time.Now().UTC()
	p.Status = enums.PurchaseStatusCancelled
	p.CancelledAt = &now
	p.CancelledReason = reason
	return nil
}

// ApplyPayment records a payment against the purchase. Returns the
// new balance.
func (p *PurchaseOrder) ApplyPayment(amount valueobjects.Money) (valueobjects.Money, error) {
	if p.Status == enums.PurchaseStatusCancelled {
		return p.Balance(), derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot pay a cancelled purchase"))
	}
	if !amount.IsPositive() {
		return p.Balance(), derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	if amount.GreaterThan(p.Balance()) {
		return p.Balance(), derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("payment exceeds outstanding balance"))
	}
	p.Paid = p.Paid.Add(amount)
	if p.Paid.Equals(p.Total) {
		p.Status = enums.PurchaseStatusPaid
	}
	return p.Balance(), nil
}

// IsPaid / IsCancelled / IsPending.
func (p *PurchaseOrder) IsPaid() bool      { return p.Status == enums.PurchaseStatusPaid }
func (p *PurchaseOrder) IsCancelled() bool { return p.Status == enums.PurchaseStatusCancelled }
func (p *PurchaseOrder) IsPending() bool   { return p.Status == enums.PurchaseStatusPending }
func (p *PurchaseOrder) IsReceived() bool  { return p.Status == enums.PurchaseStatusReceived }
func (p *PurchaseOrder) IsReconciled() bool { return p.Status == enums.PurchaseStatusReconciled }
