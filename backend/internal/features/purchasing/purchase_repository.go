package purchasing

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// PurchaseFilter is the input to PurchaseRepository.List.
type PurchaseFilter struct {
	CompanyID  *uuid.UUID
	SupplierID *uuid.UUID
	BranchID   *uuid.UUID
	Status     string
	IssueRange repositories.TimeRange
	repositories.PageRequest
}

// PurchaseRepository persists purchase orders and their line items.
type PurchaseRepository interface {
	Create(ctx context.Context, p *PurchaseOrder) error
	Update(ctx context.Context, p *PurchaseOrder) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*PurchaseOrder, error)
	GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*PurchaseOrder, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter PurchaseFilter) (repositories.Page[*PurchaseOrder], error)

	// GetNextNumber returns the next sequential number for the
	// company's purchase series. Implementations may use a database
	// sequence, a row in document_sequences, or a counter table.
	GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error)
}
