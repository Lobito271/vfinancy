package purchasing

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// ImportLotFilter filters the import lot listing.
type ImportLotFilter struct {
	Search string
	Status string
	repositories.PageRequest
}

// ImportLotRepository persists import lots and their membership.
type ImportLotRepository interface {
	// Create inserts a new lot.
	Create(ctx context.Context, lot *ImportLot) error
	// Update persists the mutable lot fields.
	Update(ctx context.Context, lot *ImportLot) error
	// SoftDelete marks the lot as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// GetByID loads a single lot without its members.
	GetByID(ctx context.Context, id uuid.UUID) (*ImportLot, error)
	// List returns the lots matching the filter.
	List(ctx context.Context, filter ImportLotFilter) (repositories.Page[*ImportLot], error)
	// NextCode returns the next "LOTE-" zero-padded sequence code.
	NextCode(ctx context.Context) (string, error)
	// AddMembers links purchase orders to the lot.
	AddMembers(ctx context.Context, lotID uuid.UUID, purchaseIDs []uuid.UUID) error
	// RemoveMember unlinks a purchase order from the lot.
	RemoveMember(ctx context.Context, lotID, purchaseID uuid.UUID) error
	// ListMembers returns the membership rows of a lot.
	ListMembers(ctx context.Context, lotID uuid.UUID) ([]*ImportLotMember, error)
	// ListPurchasesForLot returns the purchase orders of a lot.
	ListPurchasesForLot(ctx context.Context, lotID uuid.UUID) ([]*PurchaseOrder, error)
	// LotForPurchase returns the lot that contains the order, if any.
	LotForPurchase(ctx context.Context, purchaseID uuid.UUID) (*ImportLot, error)
}
