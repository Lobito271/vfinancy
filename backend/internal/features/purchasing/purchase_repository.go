package purchasing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// PurchaseFilter is the input to PurchaseRepository.List.
type PurchaseFilter struct {
	Search         string
	Status         string
	OrderType      string
	CreditCardID   *uuid.UUID
	ImportLotID    *uuid.UUID
	From           *time.Time
	To             *time.Time
	IncludeDeleted bool
	repositories.PageRequest
}

// PurchaseRepository persists purchase orders and their line items.
type PurchaseRepository interface {
	// Create inserts the order and all of its items.
	Create(ctx context.Context, po *PurchaseOrder, items []*PurchaseOrderItem) error
	// Update persists the mutable order fields.
	Update(ctx context.Context, po *PurchaseOrder) error
	// SoftDelete marks the order as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// GetByID loads a single order without its items.
	GetByID(ctx context.Context, id uuid.UUID) (*PurchaseOrder, error)
	// ListItems returns the lines of an order ordered by line number.
	ListItems(ctx context.Context, purchaseOrderID uuid.UUID) ([]*PurchaseOrderItem, error)
	// UpdateItemReceipt records the received quantity of a line.
	UpdateItemReceipt(ctx context.Context, itemID uuid.UUID, received valueobjects.Quantity) error
	// NextNumber returns the next "PO-" zero-padded sequence number.
	NextNumber(ctx context.Context) (string, error)
	// List returns the orders matching the filter; when ImportLotID is
	// set only the orders in that lot are returned.
	List(ctx context.Context, filter PurchaseFilter) (repositories.Page[*PurchaseOrder], error)
}
