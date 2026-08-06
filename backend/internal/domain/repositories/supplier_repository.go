package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/masterdata"
)

// SupplierFilter is the input to SupplierRepository.List.
type SupplierFilter struct {
	CompanyID      *uuid.UUID
	Search         string
	Status         string
	IncludeDeleted bool
	PageRequest
}

// SupplierRepository persists suppliers and the supplier
// business queries.
type SupplierRepository interface {
	Create(ctx context.Context, s *masterdata.Supplier) error
	Update(ctx context.Context, s *masterdata.Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Supplier, error)
	GetByDocument(ctx context.Context, companyID uuid.UUID, documentNumber string) (*masterdata.Supplier, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter SupplierFilter) (Page[*masterdata.Supplier], error)

	GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error)
}
