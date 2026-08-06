package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/purchasing"
)

// SupplierPaymentFilter is the input to SupplierPaymentRepository.List.
type SupplierPaymentFilter struct {
	CompanyID  *uuid.UUID
	SupplierID *uuid.UUID
	Status     string
	PayRange   TimeRange
	PageRequest
}

// SupplierPaymentRepository persists supplier payments and their
// allocations to purchase orders.
type SupplierPaymentRepository interface {
	Create(ctx context.Context, p *purchasing.SupplierPayment) error
	Update(ctx context.Context, p *purchasing.SupplierPayment) error

	GetByID(ctx context.Context, id uuid.UUID) (*purchasing.SupplierPayment, error)
	List(ctx context.Context, filter SupplierPaymentFilter) (Page[*purchasing.SupplierPayment], error)

	// ListAllocationsForPurchase returns the payments that have been
	// allocated (in part or in full) to a given purchase. Used by
	// the supplier statement report.
	ListAllocationsForPurchase(ctx context.Context, purchaseID uuid.UUID) ([]*purchasing.SupplierPayment, error)
}

// AccountsPayableRepository exposes the open-balance queries used by
// the supplier statement and dashboard widgets. It is a read-only
// repository: balances are derived from purchases minus supplier
// payments, both of which are mutated through their own repositories.
type AccountsPayableRepository interface {
	// GetOpenBalanceForSupplier returns the supplier's outstanding debt.
	GetOpenBalanceForSupplier(ctx context.Context, supplierID uuid.UUID) (string, error)
	// ListAgingBucket returns the open balance bucketed by days overdue
	// (0..30, 31..60, 61..90, 90+). Used by the AP aging report.
	ListAgingBucket(ctx context.Context, supplierID uuid.UUID) (map[string]string, error)
}
