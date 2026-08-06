package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/masterdata"
)

// WarehouseRepository persists warehouses and the global currency
// catalog. Both are simple lookups so they share a repository type
// for the moment; future separation is possible.
type WarehouseRepository interface {
	Create(ctx context.Context, w *masterdata.Warehouse) error
	Update(ctx context.Context, w *masterdata.Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Warehouse, error)
	List(ctx context.Context, companyID uuid.UUID) ([]*masterdata.Warehouse, error)
}

// CurrencyRepository persists the global currency catalog.
type CurrencyRepository interface {
	GetByCode(ctx context.Context, code string) (*masterdata.Currency, error)
	List(ctx context.Context) ([]*masterdata.Currency, error)
	Upsert(ctx context.Context, c *masterdata.Currency) error
}
