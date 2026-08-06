// Package inventory contains the domain entities that model physical
// stock movement, batches, transfers and adjustments.
//
// CRITICAL DESIGN RULE: stock is NEVER stored on the product. It is
// computed as the sum of InventoryMovement rows for a (product,
// warehouse) pair. The InventoryBatch entity carries a denormalized
// current_quantity for fast lookup, but that field is maintained by
// the application layer through the movement log and is always
// recomputable from the source of truth.
package inventory

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// ClearanceDays is the maximum number of days a batch may sit in stock
// before it is flagged as clearance. The 25-day rule is a hard business
// rule driven by the perishable-goods nature of the business.
//
// Past this date, batches are sold at reduced (clearance) prices and
// appear on the operator dashboard under "Productos en remate".
const ClearanceDays = 25

// InventoryBatch groups a set of units of a single product that share
// the same arrival date, lot and (optionally) supplier. Each batch
// tracks its own quantity and its clearance deadline.
type InventoryBatch struct {
	ID               uuid.UUID
	CompanyID        uuid.UUID
	ProductID        uuid.UUID
	WarehouseID      uuid.UUID
	SupplierID       *uuid.UUID
	PurchaseLineID   *uuid.UUID
	LotNumber        valueobjects.LotNumber
	SerialNumber     string
	ArrivalDate      valueobjects.Date
	ExpiryDate       *valueobjects.Date
	InitialQuantity  valueobjects.Quantity
	CurrentQuantity  valueobjects.Quantity
	UnitCost         valueobjects.Money
	CurrencyCode     valueobjects.CurrencyCode
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedBy        *uuid.UUID
	UpdatedBy        *uuid.UUID
}

// NewInventoryBatchOptions is the input to NewInventoryBatch.
type NewInventoryBatchOptions struct {
	CompanyID       uuid.UUID
	ProductID       uuid.UUID
	WarehouseID     uuid.UUID
	SupplierID      *uuid.UUID
	PurchaseLineID  *uuid.UUID
	LotNumber       valueobjects.LotNumber
	SerialNumber    string
	ArrivalDate     valueobjects.Date
	ExpiryDate      *valueobjects.Date
	InitialQuantity valueobjects.Quantity
	UnitCost        valueobjects.Money
	CurrencyCode    valueobjects.CurrencyCode
}

// NewInventoryBatch validates and constructs a new batch. The batch
// starts with status "active" and current_quantity == initial_quantity.
func NewInventoryBatch(now time.Time, opts NewInventoryBatchOptions) (*InventoryBatch, error) {
	if opts.CompanyID == uuid.Nil || opts.ProductID == uuid.Nil || opts.WarehouseID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company, product and warehouse are required"))
	}
	if !opts.InitialQuantity.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrNegativeQuantity, errField("initial quantity must be positive"))
	}
	if opts.UnitCost.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("unit cost cannot be negative"))
	}
	return &InventoryBatch{
		ID:              uuid.New(),
		CompanyID:       opts.CompanyID,
		ProductID:       opts.ProductID,
		WarehouseID:     opts.WarehouseID,
		SupplierID:      opts.SupplierID,
		PurchaseLineID:  opts.PurchaseLineID,
		LotNumber:       opts.LotNumber,
		SerialNumber:    opts.SerialNumber,
		ArrivalDate:     opts.ArrivalDate,
		ExpiryDate:      opts.ExpiryDate,
		InitialQuantity: opts.InitialQuantity,
		CurrentQuantity: opts.InitialQuantity,
		UnitCost:        opts.UnitCost,
		CurrencyCode:    opts.CurrencyCode,
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// MaximumSaleDate is arrival_date + 25 days. This is the boundary
// beyond which a batch is sold as clearance.
func (b *InventoryBatch) MaximumSaleDate() valueobjects.Date {
	return valueobjects.AddDays(b.ArrivalDate, ClearanceDays)
}

// DaysInStock returns the number of days between the arrival date and
// `today` (negative if the batch has not yet arrived).
func (b *InventoryBatch) DaysInStock(today valueobjects.Date) int {
	return int(today.Sub(b.ArrivalDate).Hours() / 24)
}

// DaysUntilClearance returns the number of days remaining before the
// batch hits its clearance date. Negative if already clearance.
func (b *InventoryBatch) DaysUntilClearance(today valueobjects.Date) int {
	return int(b.MaximumSaleDate().Sub(today).Hours() / 24)
}

// IsClearance reports whether the batch is past its clearance date.
// A batch with zero quantity is NOT clearance (it is depleted).
func (b *InventoryBatch) IsClearance(today valueobjects.Date) bool {
	if b.CurrentQuantity.IsZero() {
		return false
	}
	return today.After(b.MaximumSaleDate()) || today.Equal(b.MaximumSaleDate())
}

// NeedsClearanceSoon reports whether the batch is within 3 days of
// its clearance date. Used by the dashboard to surface early warnings.
func (b *InventoryBatch) NeedsClearanceSoon(today valueobjects.Date) bool {
	if b.CurrentQuantity.IsZero() {
		return false
	}
	return b.DaysUntilClearance(today) <= 3 && !b.IsClearance(today)
}

// Consume reduces current_quantity by a positive amount. Used by sales
// and outbound adjustments. Cannot reduce below zero.
func (b *InventoryBatch) Consume(amount valueobjects.Quantity) error {
	if !amount.IsPositive() {
		return derrors.Wrap(derrors.ErrNegativeQuantity, errField("consume amount must be positive"))
	}
	if amount.GreaterThan(b.CurrentQuantity) {
		return derrors.Wrap(derrors.ErrInsufficientStock, errField("consume exceeds available quantity"))
	}
	b.CurrentQuantity = b.CurrentQuantity.Sub(amount)
	if b.CurrentQuantity.IsZero() {
		b.Status = "depleted"
	}
	return nil
}

// Receive adds quantity to the batch (e.g. on a stock correction
// adjustment or a partial re-receipt from a supplier). Returns the
// new current quantity.
func (b *InventoryBatch) Receive(amount valueobjects.Quantity) (valueobjects.Quantity, error) {
	if !amount.IsPositive() {
		return b.CurrentQuantity, derrors.Wrap(derrors.ErrNegativeQuantity, errField("receive amount must be positive"))
	}
	b.CurrentQuantity = b.CurrentQuantity.Add(amount)
	if b.Status == "depleted" {
		b.Status = "active"
	}
	return b.CurrentQuantity, nil
}

// WriteOff marks the batch as written off (damage, expiry) and
// zeroes the quantity. The write-off event must be accompanied by an
// inventory_movement row in the application layer.
func (b *InventoryBatch) WriteOff() {
	b.CurrentQuantity = valueobjects.ZeroQuantity()
	b.Status = "written_off"
}
