package product

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// ProductFilter is the input to ProductRepository.List.
type ProductFilter struct {
	CompanyID      *uuid.UUID
	Search         string
	CategoryID     *uuid.UUID
	BrandID        *uuid.UUID
	IsActive       *bool
	IncludeDeleted bool
	repositories.PageRequest
}

// ProductRepository persists products. Categories and brands have
// their own repositories (CategoryRepository, BrandRepository).
type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetBySKU(ctx context.Context, companyID uuid.UUID, sku string) (*Product, error)
	GetByBarcode(ctx context.Context, companyID uuid.UUID, barcode string) (*Product, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter ProductFilter) (repositories.Page[*Product], error)
}
