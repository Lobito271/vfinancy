package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/masterdata"
)

// ProductFilter is the input to ProductRepository.List.
type ProductFilter struct {
	CompanyID      *uuid.UUID
	Search         string
	CategoryID     *uuid.UUID
	BrandID        *uuid.UUID
	IsActive       *bool
	IncludeDeleted bool
	PageRequest
}

// ProductRepository persists products. Categories and brands have
// their own repositories (CategoryRepository, BrandRepository).
type ProductRepository interface {
	Create(ctx context.Context, p *masterdata.Product) error
	Update(ctx context.Context, p *masterdata.Product) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Product, error)
	GetBySKU(ctx context.Context, companyID uuid.UUID, sku string) (*masterdata.Product, error)
	GetByBarcode(ctx context.Context, companyID uuid.UUID, barcode string) (*masterdata.Product, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter ProductFilter) (Page[*masterdata.Product], error)
}
