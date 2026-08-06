package inventory

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// InventoryMovement is an immutable, append-only stock event. Once
// persisted, a movement is never updated or deleted — corrections
// produce new compensating movements. This is the source of truth
// for stock; InventoryBatch.current_quantity is a denormalized cache.
type InventoryMovement struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	BatchID     *uuid.UUID
	Type        enums.InventoryMovementType
	// Quantity is signed. Positive for inbound, negative for outbound.
	// The entity validates sign matches Type on construction.
	Quantity    valueobjects.Quantity
	UnitCost    valueobjects.Money
	CurrencyCode valueobjects.CurrencyCode
	Reference   *valueobjects.Reference
	OccurredAt  time.Time
	Notes       string
	CreatedAt   time.Time
	CreatedBy   *uuid.UUID
}

// NewInventoryMovementOptions is the input to NewInventoryMovement.
type NewInventoryMovementOptions struct {
	CompanyID    uuid.UUID
	ProductID    uuid.UUID
	WarehouseID  uuid.UUID
	BatchID      *uuid.UUID
	Type         enums.InventoryMovementType
	Quantity     valueobjects.Quantity
	UnitCost     valueobjects.Money
	CurrencyCode valueobjects.CurrencyCode
	Reference    *valueobjects.Reference
	OccurredAt   time.Time
	Notes        string
}

// NewInventoryMovement validates and constructs a movement. The sign
// of quantity must agree with the movement type.
func NewInventoryMovement(opts NewInventoryMovementOptions) (*InventoryMovement, error) {
	if !opts.Type.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("movement type is invalid"))
	}
	if opts.Quantity.IsZero() {
		return nil, derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity cannot be zero"))
	}
	if opts.UnitCost.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("unit cost cannot be negative"))
	}
	if opts.CompanyID == uuid.Nil || opts.ProductID == uuid.Nil || opts.WarehouseID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company, product and warehouse are required"))
	}
	if opts.OccurredAt.IsZero() {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("occurred at is required"))
	}

	inbound := opts.Type.IsInbound()
	outbound := opts.Type.IsOutbound()
	if inbound && opts.Quantity.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("inbound movements must have positive quantity"))
	}
	if outbound && opts.Quantity.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("outbound movements must have negative quantity"))
	}

	return &InventoryMovement{
		ID:           uuid.New(),
		CompanyID:    opts.CompanyID,
		ProductID:    opts.ProductID,
		WarehouseID:  opts.WarehouseID,
		BatchID:      opts.BatchID,
		Type:         opts.Type,
		Quantity:     opts.Quantity,
		UnitCost:     opts.UnitCost,
		CurrencyCode: opts.CurrencyCode,
		Reference:    opts.Reference,
		OccurredAt:   opts.OccurredAt,
		Notes:        opts.Notes,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

// IsInbound / IsOutbound are convenience accessors.
func (m *InventoryMovement) IsInbound() bool  { return m.Type.IsInbound() }
func (m *InventoryMovement) IsOutbound() bool { return m.Type.IsOutbound() }

// SignedQuantity returns the quantity as a signed decimal string for
// display ("+12.5000" or "-3.0000").
func (m *InventoryMovement) SignedQuantity() string {
	if m.Quantity.IsPositive() {
		return "+" + m.Quantity.String()
	}
	return m.Quantity.String()
}

// TotalCost returns the total cost of the movement (|quantity| * unit_cost).
// For inbound movements this is the value added to inventory; for
// outbound it is the COGS to record.
func (m *InventoryMovement) TotalCost() valueobjects.Money {
	return m.UnitCost.MulByDecimal(m.Quantity.Decimal().Abs())
}
