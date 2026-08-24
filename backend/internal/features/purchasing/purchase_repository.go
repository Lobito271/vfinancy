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
	CustomerID *uuid.UUID
	BranchID   *uuid.UUID
	Status     string
	OrderType  string
	Search     string
	IssueRange repositories.TimeRange
	repositories.PageRequest
}

// PurchaseRepository persists purchase orders, their line items and the
// customer-order payment ledger.
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

	// SaveCustomerPayment persists a new customer down payment.
	SaveCustomerPayment(ctx context.Context, p *CustomerOrderPayment) error
	// UpdateCustomerPayment persists changes to a customer down payment.
	UpdateCustomerPayment(ctx context.Context, p *CustomerOrderPayment) error
	// ListCustomerPayments returns the payments recorded against an order.
	ListCustomerPayments(ctx context.Context, purchaseOrderID uuid.UUID) ([]*CustomerOrderPayment, error)
	// GetNextCustomerPaymentNumber returns the next sequential number for
	// the company's customer-payment series.
	GetNextCustomerPaymentNumber(ctx context.Context, companyID uuid.UUID) (string, error)
}
