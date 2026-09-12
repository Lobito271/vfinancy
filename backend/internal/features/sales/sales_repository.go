package sales

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// SaleFilter is the input to SaleRepository.List. Any non-zero field
// is included in the WHERE clause.
type SaleFilter struct {
	Search         string
	Status         string
	SaleType       string
	CustomerID     *uuid.UUID
	From           *time.Time
	To             *time.Time
	OnlyUnpaid     bool
	IncludeDeleted bool
	repositories.PageRequest
}

// SaleRepository persists sales and their line items.
type SaleRepository interface {
	Create(ctx context.Context, sale *Sale, items []*SaleItem) error
	Update(ctx context.Context, sale *Sale) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Sale, error)
	ListItems(ctx context.Context, saleID uuid.UUID) ([]*SaleItem, error)
	NextNumber(ctx context.Context) (string, error)
	List(ctx context.Context, filter SaleFilter) (repositories.Page[*Sale], error)
}
