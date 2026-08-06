package product

import (
	"context"

	"github.com/google/uuid"

)

// WarehouseRepository persists warehouses and the global currency
// catalog. Both are simple lookups so they share a repository type
// for the moment; future separation is possible.
type WarehouseRepository interface {
	Create(ctx context.Context, w *Warehouse) error
	Update(ctx context.Context, w *Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*Warehouse, error)
	List(ctx context.Context, companyID uuid.UUID) ([]*Warehouse, error)
}
