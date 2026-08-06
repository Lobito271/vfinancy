package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/masterdata"
)

// BrandRepository persists product brands. Brands are flat
// (non-hierarchical) reference data.
type BrandRepository interface {
	Create(ctx context.Context, b *masterdata.ProductBrand) error
	Update(ctx context.Context, b *masterdata.ProductBrand) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*masterdata.ProductBrand, error)
	List(ctx context.Context, companyID uuid.UUID) ([]*masterdata.ProductBrand, error)
}
