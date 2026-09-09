package purchasing

import (
	"context"
	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// ImportLotFilter filters the import lot listing.
type ImportLotFilter struct {
	CompanyID *uuid.UUID
	Status    string
	Search    string
	repositories.PageRequest
}

// ImportLotRepository persists import lots and their members.
type ImportLotRepository interface {
	Create(ctx context.Context, lot *ImportLot) error
	Update(ctx context.Context, lot *ImportLot) error
	GetByID(ctx context.Context, id uuid.UUID, loadMembers bool) (*ImportLot, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter ImportLotFilter) (repositories.Page[*ImportLot], error)
	GetNextCode(ctx context.Context, companyID uuid.UUID) (string, error)
	AddMember(ctx context.Context, lotID, purchaseID uuid.UUID) error
	RemoveMember(ctx context.Context, lotID, purchaseID uuid.UUID) error
	MemberCount(ctx context.Context, lotID uuid.UUID) (int, error)
	TotalUSD(ctx context.Context, lotID uuid.UUID) (valueobjects.Money, error)
}
