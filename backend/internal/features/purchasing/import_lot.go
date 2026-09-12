package purchasing

import (
	"time"

	"github.com/google/uuid"
)

// ImportLotStatus groups the lot lifecycle states.
const (
	ImportLotStatusActive = "active"
	ImportLotStatusClosed = "closed"
)

// ImportLot is a group of purchase orders bundled together for customs
// tracking. The customs authority applies a simplified import limit;
// exceeding it is allowed but the UI must confirm.
type ImportLot struct {
	ID          uuid.UUID
	Code        string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// ImportLotMember links a purchase order to an import lot.
type ImportLotMember struct {
	LotID           uuid.UUID
	PurchaseOrderID uuid.UUID
	AddedAt         time.Time
}
