package purchasing

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// USD is the transactional currency of every purchase.
const USD = "USD"

// PurchasePaymentMethod is how a purchase is paid.
type PurchasePaymentMethod string

// Purchase payment methods.
const (
	PaymentCard          PurchasePaymentMethod = "card"
	PaymentCash          PurchasePaymentMethod = "cash"
	PaymentDigitalWallet PurchasePaymentMethod = "digital_wallet"
)

// DefaultPaymentMethod is applied when the form does not explicit one.
const DefaultPaymentMethod = PaymentCard

// Valid reports whether the method is supported.
func (p PurchasePaymentMethod) Valid() bool {
	switch p {
	case PaymentCard, PaymentCash, PaymentDigitalWallet:
		return true
	}
	return false
}

// Purchase is the root aggregate for a purchase. CostUSD is the base
// cost of the ordered lines, ExtraCostUSD is the sum of the extra
// costs normalised to USD, and TotalCostUSD is their sum. All three
// are in USD (the supplier currency) with the PEN landed-cost
// projections stored alongside.
type Purchase struct {
	ID                 uuid.UUID
	Number             string
	OrderDate          time.Time
	ExpectedDate       *time.Time
	ReceivedDate       *time.Time
	ArrivalDate        *time.Time
	Status             enums.PurchaseStatus
	CurrencyCode       string
	PaymentMethod      PurchasePaymentMethod
	ExchangeRate       valueobjects.ExchangeRate
	Notes              string
	CustomerID         *uuid.UUID
	SupplierID         *uuid.UUID
	CreditCardID       *uuid.UUID
	SaleID             *uuid.UUID
	CostUSD            valueobjects.Money
	ExtraCostUSD       valueobjects.Money
	TotalCostUSD       valueobjects.Money
	SalePricePen       valueobjects.Money
	RealCostPen        valueobjects.Money
	RefundAmount       valueobjects.Money
	Faulty             bool
	FaultyReason       string
	CancelledAt        *time.Time
	CancelledReason    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time

	// Items are the purchase lines. Loaded by the repository / service.
	Items []*PurchaseItem

	// SupplierName, CustomerName, SaleNumber, ProductsText and SoldOut
	// are read-only denormalized views filled by the repository / service
	// for the list and the detail. SoldOut is true when every line of the
	// purchase has been fully sold.
	SupplierName  string
	CustomerName  string
	SaleNumber    string
	ProductsText  string
	SoldOut       bool
}

// Validate checks the aggregate invariants that hold regardless of the
// operation being performed.
func (p *Purchase) Validate() error {
	if p.ID == uuid.Nil {
		return derrors.Wrap(derrors.ErrRequired, errField("purchase id is required"))
	}
	if p.Number == "" {
		return derrors.Wrap(derrors.ErrRequired, errField("purchase number is required"))
	}
	if p.OrderDate.IsZero() {
		return derrors.Wrap(derrors.ErrRequired, errField("order date is required"))
	}
	if !p.ExchangeRate.Decimal().IsPositive() {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("exchange rate must be positive"))
	}
	if !p.PaymentMethod.Valid() {
		return derrors.Wrap(derrors.ErrInvalidEnum, errField("payment method is invalid"))
	}
	// Recorded amounts must never be negative.
	if p.CostUSD.IsNegative() || p.ExtraCostUSD.IsNegative() || p.TotalCostUSD.IsNegative() ||
		p.SalePricePen.IsNegative() || p.RealCostPen.IsNegative() ||
		p.RefundAmount.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("financial amounts cannot be negative"))
	}
	return nil
}

// RecomputeTotals derives ExtraCostUSD and TotalCostUSD from the base
// cost and the normalised extra total. It is called after every extra
// cost mutation so the stored totals always satisfy
// TotalCostUSD = CostUSD + ExtraCostUSD.
func (p *Purchase) RecomputeTotals() {
	total, _ := valueobjects.MoneyFromDecimal(p.CostUSD.Decimal().Add(p.ExtraCostUSD.Decimal()))
	p.TotalCostUSD = total.RoundToCurrencyPrecision()
}

// IsPending / IsReceived / IsCancelled report the lifecycle state.
func (p *Purchase) IsPending() bool   { return p.Status == enums.PurchaseStatusPending }
func (p *Purchase) IsReceived() bool  { return p.Status == enums.PurchaseStatusReceived }
func (p *Purchase) IsCancelled() bool { return p.Status == enums.PurchaseStatusCancelled }

// MarkReceived transitions a pending order to received and stamps the
// receipt (and arrival) date used by the inventory aging rule.
func (p *Purchase) MarkReceived(at time.Time) error {
	if p.Status != enums.PurchaseStatusPending {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("only pending purchases can be marked as received"))
	}
	p.Status = enums.PurchaseStatusReceived
	p.ReceivedDate = &at
	p.ArrivalDate = &at
	return nil
}

// Cancel voids a pending or received order. The service restores
// inventory and releases the card charge as compensation.
func (p *Purchase) Cancel(reason string, at time.Time) error {
	if p.Status == enums.PurchaseStatusCancelled {
		return derrors.Wrap(derrors.ErrPurchaseCancelled, errField("purchase is already cancelled"))
	}
	p.Status = enums.PurchaseStatusCancelled
	p.CancelledAt = &at
	p.CancelledReason = reason
	return nil
}
