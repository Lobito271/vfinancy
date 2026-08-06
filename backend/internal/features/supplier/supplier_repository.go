package supplier

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// SupplierFilter is the input to SupplierRepository.List.
type SupplierFilter struct {
	CompanyID      *uuid.UUID
	Search         string
	Status         string
	IncludeDeleted bool
	repositories.PageRequest
}

// SupplierRepository persists suppliers and the supplier
// business queries.
type SupplierRepository interface {
	Create(ctx context.Context, s *Supplier) error
	Update(ctx context.Context, s *Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*Supplier, error)
	GetByDocument(ctx context.Context, companyID uuid.UUID, documentNumber string) (*Supplier, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter SupplierFilter) (repositories.Page[*Supplier], error)

	GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error)
}
