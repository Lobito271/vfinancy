package inventory

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"
	"time"

	"github.com/google/uuid"

)

// InventoryBatchFilter is the input to InventoryBatchRepository.List.
type InventoryBatchFilter struct {
	CompanyID   *uuid.UUID
	ProductID   *uuid.UUID
	WarehouseID *uuid.UUID
	OnlyActive  bool          // exclude depleted / written-off
	OnlyClearance bool        // batches past their maximum sale date
	ArrivalRange repositories.TimeRange
	repositories.PageRequest
}

// InventoryBatchRepository persists inventory batches. A batch is the
// per-warehouse, per-arrival, per-lot group of units of a single
// product; current_quantity is denormalized for fast lookup.
type InventoryBatchRepository interface {
	Create(ctx context.Context, b *InventoryBatch) error
	Update(ctx context.Context, b *InventoryBatch) error

	GetByID(ctx context.Context, id uuid.UUID) (*InventoryBatch, error)
	List(ctx context.Context, filter InventoryBatchFilter) (repositories.Page[*InventoryBatch], error)

	// GetStockSummary returns the available quantity and weighted
	// average cost for a (product, warehouse) pair. The summary is
	// kept in sync by InventoryMovementRepository; this method
	// returns a snapshot.
	GetStockSummary(ctx context.Context, productID, warehouseID uuid.UUID) (quantity float64, averageCost string, err error)

	// ListClearance returns the batches past their maximum sale date
	// that still have stock. Used by the dashboard widget.
	ListClearance(ctx context.Context, companyID uuid.UUID, at time.Time) ([]*InventoryBatch, error)
}
