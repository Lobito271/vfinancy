package auth

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// RoleFilter is the input to RoleRepository.List.
type RoleFilter struct {
	CompanyID      *uuid.UUID
	Name           string
	Type           string
	IncludeDeleted bool
	repositories.PageRequest
}

// RoleRepository persists roles and the role → permission assignment.
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter RoleFilter) (repositories.Page[*Role], error)

	// Grant persists a single permission grant. Implementations are
	// expected to upsert (no error on duplicate).
	Grant(ctx context.Context, roleID uuid.UUID, permissionCode string) error
	// Revoke removes a single permission grant.
	Revoke(ctx context.Context, roleID uuid.UUID, permissionCode string) error
	// ReplacePermissions atomically replaces the full set of
	// permission codes for a role. Used by the application layer when
	// editing a role's permission list in the admin UI.
	ReplacePermissions(ctx context.Context, roleID uuid.UUID, codes []string) error
	// ListPermissionCodes returns the permission codes granted to a role.
	ListPermissionCodes(ctx context.Context, roleID uuid.UUID) ([]string, error)
}
