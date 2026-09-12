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
// produce new compensating movements. The ledger is the source of
// truth for stock; InventoryBatch.quantity is a denormalized cache
// maintained by the service.
type InventoryMovement struct {
	ID            uuid.UUID
	BatchID       uuid.UUID
	ProductID     uuid.UUID
	MovementDate  time.Time
	Type          enums.InventoryMovementType
	Reference     *valueobjects.Reference
	QuantityDelta valueobjects.Quantity
	BalanceAfter  valueobjects.Quantity
	UnitCost      valueobjects.Money
	Notes         string
	CreatedAt     time.Time
}

// NewInventoryMovementOptions is the input to NewInventoryMovement.
type NewInventoryMovementOptions struct {
	BatchID      uuid.UUID
	ProductID    uuid.UUID
	Type         enums.InventoryMovementType
	Quantity     valueobjects.Quantity
	BalanceAfter valueobjects.Quantity
	UnitCost     valueobjects.Money
	Reference    *valueobjects.Reference
	Notes        string
}

// NewInventoryMovement validates and constructs a movement. The sign
// of the quantity must agree with the movement type.
func NewInventoryMovement(now time.Time, opts NewInventoryMovementOptions) (*InventoryMovement, error) {
	if opts.BatchID == uuid.Nil || opts.ProductID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("batch and product are required"))
	}
	if !opts.Type.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("movement type is invalid"))
	}
	if opts.Quantity.IsZero() {
		return nil, derrors.Wrap(derrors.ErrNegativeQuantity, errField("quantity cannot be zero"))
	}
	if opts.UnitCost.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("unit cost cannot be negative"))
	}
	if opts.Type.IsInbound() && opts.Quantity.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("inbound movements must have positive quantity"))
	}
	if opts.Type.IsOutbound() && opts.Quantity.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("outbound movements must have negative quantity"))
	}
	return &InventoryMovement{
		ID:            uuid.New(),
		BatchID:       opts.BatchID,
		ProductID:     opts.ProductID,
		MovementDate:  now,
		Type:          opts.Type,
		Reference:     opts.Reference,
		QuantityDelta: opts.Quantity,
		BalanceAfter:  opts.BalanceAfter,
		UnitCost:      opts.UnitCost,
		Notes:         opts.Notes,
		CreatedAt:     now,
	}, nil
}

// IsInbound / IsOutbound are convenience accessors.
func (m *InventoryMovement) IsInbound() bool  { return m.Type.IsInbound() }
func (m *InventoryMovement) IsOutbound() bool { return m.Type.IsOutbound() }

// SignedQuantity returns the quantity as a signed decimal string for
// display ("+12.5000" or "-3.0000").
func (m *InventoryMovement) SignedQuantity() string {
	if m.QuantityDelta.IsPositive() {
		return "+" + m.QuantityDelta.String()
	}
	return m.QuantityDelta.String()
}

// TotalCost returns the total cost of the movement (|quantity| * unit_cost).
// For inbound movements this is the value added to inventory; for
// outbound it is the COGS to record.
func (m *InventoryMovement) TotalCost() valueobjects.Money {
	return m.UnitCost.MulByDecimal(m.QuantityDelta.Decimal().Abs())
}
