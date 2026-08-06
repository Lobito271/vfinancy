package customer

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// CustomerFilter is the input to CustomerRepository.List. Any
// non-zero field is included in the WHERE clause.
type CustomerFilter struct {
	CompanyID      *uuid.UUID
	Search         string // matches business_name, document_number, contact_name
	Status         string
	BranchID       *uuid.UUID
	IncludeDeleted bool
	repositories.PageRequest
}

// CustomerRepository persists customers and exposes the
// customer-aggregate business queries (outstanding balance, etc.).
type CustomerRepository interface {
	Create(ctx context.Context, c *Customer) error
	Update(ctx context.Context, c *Customer) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	GetByDocument(ctx context.Context, companyID uuid.UUID, documentType, documentNumber string) (*Customer, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter CustomerFilter) (repositories.Page[*Customer], error)

	// GetOutstandingBalance returns the customer's current debt as
	// stored in the row. The application layer is responsible for
	// surfacing "over-limit" warnings based on credit_limit.
	GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error)
}
