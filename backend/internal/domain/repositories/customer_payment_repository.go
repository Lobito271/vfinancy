package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/sales"
)

// CustomerPaymentFilter is the input to
// CustomerPaymentRepository.List.
type CustomerPaymentFilter struct {
	CompanyID  *uuid.UUID
	CustomerID *uuid.UUID
	Status     string
	PayRange   TimeRange
	PageRequest
}

// CustomerPaymentRepository persists customer payments and their
// allocations to sales.
type CustomerPaymentRepository interface {
	Create(ctx context.Context, p *sales.CustomerPayment) error
	Update(ctx context.Context, p *sales.CustomerPayment) error

	GetByID(ctx context.Context, id uuid.UUID) (*sales.CustomerPayment, error)
	List(ctx context.Context, filter CustomerPaymentFilter) (Page[*sales.CustomerPayment], error)

	ListAllocationsForSale(ctx context.Context, saleID uuid.UUID) ([]*sales.CustomerPayment, error)
}

// CustomerAdvanceRepository persists customer advances (payments
// received before an invoice). Customers apply advances to one or
// more future sales.
type CustomerAdvanceRepository interface {
	Create(ctx context.Context, a *sales.CustomerAdvance) error
	Update(ctx context.Context, a *sales.CustomerAdvance) error

	GetByID(ctx context.Context, id uuid.UUID) (*sales.CustomerAdvance, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*sales.CustomerAdvance, error)

	// ListApplicationsForSale returns the advances that have been
	// applied (in part or in full) to a given sale.
	ListApplicationsForSale(ctx context.Context, saleID uuid.UUID) ([]*sales.CustomerAdvance, error)
}

// AccountsReceivableRepository exposes the open-balance queries used
// by the customer statement and dashboard widgets.
type AccountsReceivableRepository interface {
	// GetOpenBalanceForCustomer returns the customer's outstanding debt.
	GetOpenBalanceForCustomer(ctx context.Context, customerID uuid.UUID) (string, error)
	// ListAgingBucket returns the open balance bucketed by days overdue.
	ListAgingBucket(ctx context.Context, customerID uuid.UUID) (map[string]string, error)
}
