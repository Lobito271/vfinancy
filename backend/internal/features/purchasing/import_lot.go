package purchasing

import (
	"time"

	"github.com/google/uuid"
)

// ImportLot is a group of purchase orders bundled together for customs
// tracking. The custom authority applies a simplified import limit;
// exceeding it is allowed but requires explicit user confirmation.
type ImportLot struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Code        string
	Description string
	Status      string
	Members     []*ImportLotMember
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// ImportLotMember links a purchase order to an import lot.
type ImportLotMember struct {
	ImportLotID uuid.UUID
	PurchaseID  uuid.UUID
	AddedAt     time.Time
}
