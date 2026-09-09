package product

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// ProductFilter is the input to ProductRepository.List. Any non-zero
// field is included in the WHERE clause.
type ProductFilter struct {
	Search         string
	IsActive       *bool
	IncludeDeleted bool
	repositories.PageRequest
}

// ProductRepository persists products.
type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetByDescription(ctx context.Context, description string) (*Product, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter ProductFilter) (repositories.Page[*Product], error)
}
