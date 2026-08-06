package product

import (
	"context"

	"github.com/google/uuid"

)

// BrandRepository persists product brands. Brands are flat
// (non-hierarchical) reference data.
type BrandRepository interface {
	Create(ctx context.Context, b *ProductBrand) error
	Update(ctx context.Context, b *ProductBrand) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*ProductBrand, error)
	List(ctx context.Context, companyID uuid.UUID) ([]*ProductBrand, error)
}
