package sales

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// SaleFilter is the input to SalesRepository.List.
type SaleFilter struct {
	CompanyID  *uuid.UUID
	CustomerID *uuid.UUID
	BranchID   *uuid.UUID
	SellerID   *uuid.UUID
	Status     string
	IssueRange repositories.TimeRange
	repositories.PageRequest
}

// SalesRepository persists sales and their line items. Updates to a
// posted sale are blocked at the entity level; the repository does
// not enforce immutability on its own.
type SalesRepository interface {
	Create(ctx context.Context, s *Sale) error
	Update(ctx context.Context, s *Sale) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*Sale, error)
	GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*Sale, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter SaleFilter) (repositories.Page[*Sale], error)

	GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error)
}
