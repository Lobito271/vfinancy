// Package shipment implements the business logic for the Envios
// (shipments) module: physical dispatch records linked to a sale, with
// an immutable alphanumeric code and a four-digit security code the
// recipient presents on delivery.
package shipment

import (
	"regexp"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
)

var (
	// codePattern matches the year-scoped shipment code, e.g. ENV-2026-00001.
	codePattern = regexp.MustCompile(`^ENV-\d{4}-\d{5}$`)
	// securityCodePattern matches the four digits the recipient reads out.
	securityCodePattern = regexp.MustCompile(`^\d{4}$`)
)

// Shipment tracks the dispatch of goods to a customer. Code,
// SecurityCode and SaleID are assigned at creation and are never
// mutated afterwards.
type Shipment struct {
	ID           uuid.UUID
	Code         string
	SecurityCode string
	SaleID       uuid.UUID
	CustomerID   *uuid.UUID
	Location     string
	ShipmentDate *time.Time
	DeliveredAt  *time.Time
	Description  string
	Notes        string
	Status       enums.ShipmentStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time

	// SaleNumber is a read-only view of the originating sale, filled by
	// the repository for the list and the detail.
	SaleNumber string
}

// Validate checks the aggregate invariants. Code and SecurityCode are
// only re-checked while the shipment is new; once inserted they are
// immutable and no longer part of the edit surface.
func (s *Shipment) Validate(isNew bool) error {
	if s.ID == uuid.Nil {
		return derrors.Wrap(derrors.ErrRequired, errField("shipment id is required"))
	}
	if isNew {
		if !codePattern.MatchString(s.Code) {
			return derrors.Wrap(derrors.ErrOutOfRange, errField("code must look like ENV-2026-00001"))
		}
		if !securityCodePattern.MatchString(s.SecurityCode) {
			return derrors.Wrap(derrors.ErrOutOfRange, errField("security code must be exactly 4 digits"))
		}
	}
	if s.SaleID == uuid.Nil {
		return derrors.Wrap(derrors.ErrRequired, errField("shipment must be linked to a sale"))
	}
	if !s.Status.Valid() {
		return derrors.Wrap(derrors.ErrInvalidEnum, errField("status is invalid"))
	}
	if s.ShipmentDate != nil && s.DeliveredAt != nil && s.DeliveredAt.Before(*s.ShipmentDate) {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("delivery date cannot precede the shipment date"))
	}
	return nil
}
