package valueobjects

import (
	"github.com/google/uuid"
	"vfinancy/backend/internal/domain/enums"
)

// Reference is a polymorphic pointer from one entity to another. The
// target entity is identified by (Type, ID) instead of a typed FK.
//
// The domain layer does not enforce referential integrity for
// references — that is the database's job (via a CHECK on Type).
// In-memory, the application layer is responsible for resolving the
// reference to a concrete aggregate.
type Reference struct {
	Type enums.ReferenceType
	ID   uuid.UUID
}

// NewReference builds and validates a polymorphic reference.
func NewReference(t enums.ReferenceType, id uuid.UUID) (Reference, error) {
	if !t.Valid() {
		return Reference{}, wrapInvalid("reference type is invalid: " + string(t))
	}
	if id == uuid.Nil {
		return Reference{}, wrapInvalid("reference id is the zero UUID")
	}
	return Reference{Type: t, ID: id}, nil
}

func (r Reference) IsZero() bool { return r.ID == uuid.Nil }
