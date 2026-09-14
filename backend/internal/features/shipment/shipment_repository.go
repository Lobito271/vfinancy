package shipment

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// ShipmentFilter is the input to ShipmentRepository.List.
type ShipmentFilter struct {
	Search string
	Status string
	repositories.PageRequest
}

// ShipmentRepository persists shipments.
type ShipmentRepository interface {
	// Create inserts the shipment.
	Create(ctx context.Context, s *Shipment) error
	// Update persists the mutable shipment fields.
	Update(ctx context.Context, s *Shipment) error
	// SoftDelete marks the shipment as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// GetByID loads a single shipment.
	GetByID(ctx context.Context, id uuid.UUID) (*Shipment, error)
	// NextCode returns the next four-digit sequence code for new
	// shipments.
	NextCode(ctx context.Context) (string, error)
	// List returns the shipments matching the filter.
	List(ctx context.Context, filter ShipmentFilter) (repositories.Page[*Shipment], error)
}
