package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserRoleRepository persists the user × role junction (and the
// optional branch scope / expiry). Most operations live on
// RoleRepository (Grant / Revoke) but the user × role side is here.
type UserRoleRepository interface {
	// Assign persists a single role assignment. Implementations are
	// expected to upsert on (user_id, role_id, branch_id) so that
	// re-assigning the same role updates the expires_at.
	Assign(ctx context.Context, userID, roleID uuid.UUID, branchID *uuid.UUID, expiresAt *time.Time) error
	// Revoke removes a single assignment. The branchID must match
	// (or be nil for company-wide grants).
	Revoke(ctx context.Context, userID, roleID uuid.UUID, branchID *uuid.UUID) error
	// EffectiveRoles returns the role assignments currently active
	// (not expired) for a user. Used by the application layer to
	// compute the effective permission set.
	EffectiveRoles(ctx context.Context, userID uuid.UUID, at time.Time) ([]UserRoleAssignment, error)
}
