package customer

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// CustomerFilter is the input to CustomerRepository.List. Any
// non-zero field is included in the WHERE clause.
type CustomerFilter struct {
	Search         string
	Status         string
	IncludeDeleted bool
	repositories.PageRequest
}

// CustomerRepository persists customers and exposes the
// customer-aggregate business queries (outstanding balance, etc.).
type CustomerRepository interface {
	Create(ctx context.Context, c *Customer) error
	Update(ctx context.Context, c *Customer) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	GetByDocument(ctx context.Context, documentType, documentNumber string) (*Customer, error)
	GetByDescription(ctx context.Context, name string) (*Customer, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter CustomerFilter) (repositories.Page[*Customer], error)
	GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error)
}
