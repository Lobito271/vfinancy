package sales

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// CustomerPaymentFilter is the input to
// CustomerPaymentRepository.List.
type CustomerPaymentFilter struct {
	CompanyID  *uuid.UUID
	CustomerID *uuid.UUID
	Status     string
	PayRange   repositories.TimeRange
	repositories.PageRequest
}

// CustomerPaymentRepository persists customer payments and their
// allocations to 
type CustomerPaymentRepository interface {
	Create(ctx context.Context, p *CustomerPayment) error
	Update(ctx context.Context, p *CustomerPayment) error

	GetByID(ctx context.Context, id uuid.UUID) (*CustomerPayment, error)
	List(ctx context.Context, filter CustomerPaymentFilter) (repositories.Page[*CustomerPayment], error)

	ListAllocationsForSale(ctx context.Context, saleID uuid.UUID) ([]*CustomerPayment, error)

	// GetNextNumber returns the next sequential number for the
	// company's customer payment series.
	GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error)
}

// CustomerAdvanceRepository persists customer advances (payments
// received before an invoice). Customers apply advances to one or
// more future 
type CustomerAdvanceRepository interface {
	Create(ctx context.Context, a *CustomerAdvance) error
	Update(ctx context.Context, a *CustomerAdvance) error

	GetByID(ctx context.Context, id uuid.UUID) (*CustomerAdvance, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*CustomerAdvance, error)

	// ListApplicationsForSale returns the advances that have been
	// applied (in part or in full) to a given sale.
	ListApplicationsForSale(ctx context.Context, saleID uuid.UUID) ([]*CustomerAdvance, error)
}

// AccountsReceivableRepository exposes the open-balance queries used
// by the customer statement and dashboard widgets.
type AccountsReceivableRepository interface {
	// GetOpenBalanceForCustomer returns the customer's outstanding debt.
	GetOpenBalanceForCustomer(ctx context.Context, customerID uuid.UUID) (string, error)
	// ListAgingBucket returns the open balance bucketed by days overdue.
	ListAgingBucket(ctx context.Context, customerID uuid.UUID) (map[string]string, error)
}
