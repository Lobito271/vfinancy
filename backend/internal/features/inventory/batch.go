package inventory

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// ClearanceDays is the default number of days a batch may sit in stock
// before it is flagged as clearance. The 25-day rule is a hard business
// rule driven by the perishable-goods nature of the business; the
// user-configured setting (if any) overrides it. Past this date,
// batches are sold at reduced (clearance) prices and appear on the
// operator dashboard under "Productos en remate".
const ClearanceDays = 25

// InventoryBatch groups the units of a single product that arrived on
// the same date (typically from one purchase line). Each batch tracks
// its own quantity and its clearance deadline.
type InventoryBatch struct {
	ID                  uuid.UUID
	ProductID           uuid.UUID
	PurchaseOrderItemID *uuid.UUID
	ArrivalDate         valueobjects.Date
	Quantity            valueobjects.Quantity
	OriginalQuantity    valueobjects.Quantity
	UnitCost            valueobjects.Money
	ExchangeRate        valueobjects.ExchangeRate
	Status              enums.BatchStatus
	IsClearance         bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CreatedBy           string
	UpdatedBy           string
}

// NewInventoryBatchOptions is the input to NewInventoryBatch.
type NewInventoryBatchOptions struct {
	ProductID           uuid.UUID
	PurchaseOrderItemID *uuid.UUID
	ArrivalDate         valueobjects.Date
	InitialQuantity     valueobjects.Quantity
	UnitCost            valueobjects.Money
	ExchangeRate        valueobjects.ExchangeRate
}

// NewInventoryBatch validates and constructs a new batch. The batch
// starts with status "active" and quantity == original_quantity; the
// arrival date cannot be in the future.
func NewInventoryBatch(now time.Time, opts NewInventoryBatchOptions) (*InventoryBatch, error) {
	if opts.ProductID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("product is required"))
	}
	if opts.ArrivalDate.IsZero() {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("arrival date is required"))
	}
	today := valueobjects.NewDateFromTime(now)
	if opts.ArrivalDate.After(today) {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("arrival date cannot be in the future"))
	}
	if !opts.InitialQuantity.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrNegativeQuantity, errField("initial quantity must be positive"))
	}
	if opts.UnitCost.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("unit cost cannot be negative"))
	}
	if !opts.ExchangeRate.Decimal().IsPositive() {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("exchange rate must be positive"))
	}
	return &InventoryBatch{
		ID:                  uuid.New(),
		ProductID:           opts.ProductID,
		PurchaseOrderItemID: opts.PurchaseOrderItemID,
		ArrivalDate:         opts.ArrivalDate,
		Quantity:            opts.InitialQuantity,
		OriginalQuantity:    opts.InitialQuantity,
		UnitCost:            opts.UnitCost,
		ExchangeRate:        opts.ExchangeRate,
		Status:              enums.BatchStatusActive,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// Consume reduces quantity by a positive amount. Used by sales and
// outbound movements. It cannot reduce below zero and rejects
// consumption from a voided batch.
func (b *InventoryBatch) Consume(qty valueobjects.Quantity) error {
	if b.Status == enums.BatchStatusVoided {
		return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot consume from a voided batch"))
	}
	if !qty.IsPositive() {
		return derrors.Wrap(derrors.ErrNegativeQuantity, errField("consume quantity must be positive"))
	}
	if qty.GreaterThan(b.Quantity) {
		return derrors.Wrap(derrors.ErrInsufficientStock, errField("consume exceeds available quantity"))
	}
	b.Quantity = b.Quantity.Sub(qty)
	if b.Quantity.IsZero() {
		b.Status = enums.BatchStatusDepleted
	}
	return nil
}

// Receive adds quantity to the batch (stock corrections, sale voids).
// It returns the new quantity and rejects receipt into a voided batch.
func (b *InventoryBatch) Receive(qty valueobjects.Quantity) (valueobjects.Quantity, error) {
	if b.Status == enums.BatchStatusVoided {
		return b.Quantity, derrors.Wrap(derrors.ErrInvalidStateTransition, errField("cannot receive into a voided batch"))
	}
	if !qty.IsPositive() {
		return b.Quantity, derrors.Wrap(derrors.ErrNegativeQuantity, errField("receive quantity must be positive"))
	}
	b.Quantity = b.Quantity.Add(qty)
	if b.Status == enums.BatchStatusDepleted {
		b.Status = enums.BatchStatusActive
	}
	return b.Quantity, nil
}

// Touch refreshes the updated_at timestamp.
func (b *InventoryBatch) Touch() { b.UpdatedAt = time.Now().UTC() }

// Validate checks the entity invariants.
func (b *InventoryBatch) Validate() error {
	if b.ProductID == uuid.Nil {
		return derrors.Wrap(derrors.ErrRequired, errField("product is required"))
	}
	if b.Quantity.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity cannot be negative"))
	}
	if b.UnitCost.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("unit cost cannot be negative"))
	}
	if !b.ExchangeRate.Decimal().IsPositive() {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("exchange rate must be positive"))
	}
	if !b.Status.Valid() {
		return derrors.Wrap(derrors.ErrInvalidEnum, errField("status is invalid"))
	}
	return nil
}

// MaxSaleDate returns the last date the batch may be sold at full
// price: arrival date plus the clearance period.
func (b *InventoryBatch) MaxSaleDate(clearanceDays int) valueobjects.Date {
	if clearanceDays < 0 {
		clearanceDays = ClearanceDays
	}
	return valueobjects.AddDays(b.ArrivalDate, clearanceDays)
}

// IsClearanceOn reports whether a stocked batch has reached its
// clearance date as of `at`.
func (b *InventoryBatch) IsClearanceOn(clearanceDays int, at time.Time) bool {
	if b.Quantity.IsZero() {
		return false
	}
	today := valueobjects.NewDateFromTime(at)
	date := b.MaxSaleDate(clearanceDays)
	return today.After(date) || today.Equal(date)
}
