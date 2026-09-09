package inventory

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// InventoryMovementFilter is the input to
// InventoryMovementRepository.List. Used both for reporting and for
// reconciling a document's effects on stock.
type InventoryMovementFilter struct {
	ProductID     *uuid.UUID
	BatchID       *uuid.UUID
	ReferenceType string
	ReferenceID   *uuid.UUID
	repositories.PageRequest
}

// InventoryMovementRepository persists stock events. Movements are
// append-only: there is no Update or Delete method.
type InventoryMovementRepository interface {
	Create(ctx context.Context, m *InventoryMovement) error
	List(ctx context.Context, filter InventoryMovementFilter) (repositories.Page[*InventoryMovement], error)
}
