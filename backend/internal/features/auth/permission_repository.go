package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

)

// PermissionRepository persists the global permissions catalog. The
// catalog is seeded by an SQL migration and rarely mutated at runtime;
// this repository exists for read access from the application layer.
type PermissionRepository interface {
	GetByCode(ctx context.Context, code string) (*Permission, error)
	Exists(ctx context.Context, code string) (bool, error)
	ListAll(ctx context.Context) ([]*Permission, error)
	// ListCodesForRoles returns the union of permission codes granted
	// to any of the given role IDs. Used by the application layer to
	// build the effective permission set of a user.
	ListCodesForRoles(ctx context.Context, roleIDs []string) ([]string, error)
}

// UserRoleAssignment is the join row stored in user_roles.
type UserRoleAssignment struct {
	UserID    uuid.UUID
	RoleID    uuid.UUID
	BranchID  *uuid.UUID
	ExpiresAt *time.Time
}
