package inventory

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"
	"time"

	"github.com/google/uuid"

)

// InventoryMovementFilter is the input to
// InventoryMovementRepository.List. Used both for reporting and for
// reconciling a document's effects on stock.
type InventoryMovementFilter struct {
	CompanyID   *uuid.UUID
	ProductID   *uuid.UUID
	WarehouseID *uuid.UUID
	BatchID     *uuid.UUID
	ReferenceType string
	ReferenceID *uuid.UUID
	OccurredRange repositories.TimeRange
	repositories.PageRequest
}

// InventoryMovementRepository persists stock events. Movements are
// append-only: there is no Update or Delete method.
type InventoryMovementRepository interface {
	// Create persists a single movement. The implementation must
	// maintain the inventory_stock summary in the same transaction
	// (when called inside one) and update the affected batch's
	// current_quantity.
	Create(ctx context.Context, m *InventoryMovement) error

	GetByID(ctx context.Context, id uuid.UUID) (*InventoryMovement, error)
	List(ctx context.Context, filter InventoryMovementFilter) (repositories.Page[*InventoryMovement], error)

	// StockAt returns the net quantity (sum of inbound minus outbound)
	// for a (product, warehouse) pair as of `at`. This is the
	// recomputable source of truth; in production most reads use
	// the inventory_stock summary, but this method exists for
	// reconciliation.
	StockAt(ctx context.Context, productID, warehouseID uuid.UUID, at time.Time) (float64, error)
}
