// Package shipment implements the business logic for the Envios
// (shipments) module: physical dispatch records with an immutable
// four-digit numeric code that serves as the shareable reference.
package shipment

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
)

// Shipment tracks the dispatch of goods to a customer, optionally
// linked to the sale that originated it.
type Shipment struct {
	ID          uuid.UUID
	Code        string
	SaleID      *uuid.UUID
	CustomerID  *uuid.UUID
	Description string
	Notes       string
	Status      enums.ShipmentStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Validate checks the aggregate invariants. The code is assigned at
// creation and is never mutated by later updates (newCode holds after
// the first insert).
func (s *Shipment) Validate(newCode bool) error {
	if s.ID == uuid.Nil {
		return derrors.Wrap(derrors.ErrRequired, errField("shipment id is required"))
	}
	if newCode && !validCode(s.Code) {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("code must be exactly 4 digits"))
	}
	if !s.Status.Valid() {
		return derrors.Wrap(derrors.ErrInvalidEnum, errField("status is invalid"))
	}
	return nil
}

// validCode reports whether s is exactly four numeric digits.
func validCode(s string) bool {
	if len(s) != 4 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
