package sales

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// CustomerPaymentFilter is the input to CustomerPaymentRepository.List.
type CustomerPaymentFilter struct {
	CustomerID *uuid.UUID
	SaleID     *uuid.UUID
	From       *time.Time
	To         *time.Time
	repositories.PageRequest
}

// CustomerPaymentRepository persists customer payments and their
// allocations to sales.
type CustomerPaymentRepository interface {
	Create(ctx context.Context, p *CustomerPayment, allocations []PaymentAllocation) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	NextNumber(ctx context.Context) (string, error)
	List(ctx context.Context, filter CustomerPaymentFilter) (repositories.Page[*CustomerPayment], error)
	ListForSale(ctx context.Context, saleID uuid.UUID) ([]*CustomerPayment, error)
	// ListCollections returns the sale allocations of active payments
	// whose payment date falls in [from, to).
	ListCollections(ctx context.Context, from, to time.Time) ([]SaleCollection, error)
}
