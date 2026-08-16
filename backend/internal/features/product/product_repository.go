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

	// ListUnits returns the units of measure available for a company.
	ListUnits(ctx context.Context, companyID uuid.UUID) ([]*UnitOfMeasure, error)

	// ListCategories returns the product categories available for a company.
	ListCategories(ctx context.Context, companyID uuid.UUID) ([]*ProductCategory, error)

	// CreateCategory persists a new category.
	CreateCategory(ctx context.Context, c *ProductCategory) error

	// UpdateCategory persists the mutable category fields.
	UpdateCategory(ctx context.Context, c *ProductCategory) error

	// DeleteCategory soft-deletes a category (sets deleted_at).
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// GetCategoryByID returns a single non-deleted category.
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*ProductCategory, error)

	// ListBrands returns the product brands available for a company.
	ListBrands(ctx context.Context, companyID uuid.UUID) ([]*ProductBrand, error)

	// CreateBrand persists a new brand.
	CreateBrand(ctx context.Context, b *ProductBrand) error

	// UpdateBrand persists the mutable brand fields.
	UpdateBrand(ctx context.Context, b *ProductBrand) error

	// DeleteBrand soft-deletes a brand (sets deleted_at).
	DeleteBrand(ctx context.Context, id uuid.UUID) error

	// GetBrandByID returns a single non-deleted brand.
	GetBrandByID(ctx context.Context, id uuid.UUID) (*ProductBrand, error)
}
