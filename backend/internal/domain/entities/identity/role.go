package identity

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Role is a named bundle of permissions scoped to a Company. System
// roles (admin, manager, etc.) are seeded and cannot be modified.
// Custom roles are managed at runtime.
type Role struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Code        valueobjects.ShortCode
	Name        valueobjects.FullName
	Description string
	Type        enums.RoleType
	IsActive    bool
	permissions map[string]Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// NewRoleOptions is the input to NewRole.
type NewRoleOptions struct {
	CompanyID   uuid.UUID
	Code        valueobjects.ShortCode
	Name        valueobjects.FullName
	Description string
	Type        enums.RoleType
}

// NewRole constructs a Role.
func NewRole(now time.Time, opts NewRoleOptions) (*Role, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if !opts.Type.Valid() {
		return nil, errors.Wrap(errors.ErrInvalidEnum, errField("role type is invalid"))
	}
	return &Role{
		ID:          uuid.New(),
		CompanyID:   opts.CompanyID,
		Code:        opts.Code,
		Name:        opts.Name,
		Description: opts.Description,
		Type:        opts.Type,
		IsActive:    true,
		permissions: make(map[string]Permission),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Grant adds a permission to the role. System roles reject this method
// (callers must edit the seed in SQL and re-apply the migration).
func (r *Role) Grant(p Permission) error {
	if r.Type == enums.RoleTypeSystem {
		return errors.Wrap(errors.ErrInvalidStateTransition, errField("system roles cannot be modified at runtime"))
	}
	r.permissions[p.Code] = p
	return nil
}

// Revoke removes a permission from the role. No-op if absent.
func (r *Role) Revoke(code string) {
	delete(r.permissions, code)
}

// Permissions returns the granted permission codes. Callers must not
// mutate the returned slice.
func (r *Role) Permissions() []string {
	out := make([]string, 0, len(r.permissions))
	for k := range r.permissions {
		out = append(out, k)
	}
	return out
}

// HasPermission checks whether this role has a given permission.
func (r *Role) HasPermission(code string) bool {
	_, ok := r.permissions[code]
	return ok
}

// Activate / Deactivate.
func (r *Role) Activate()   { r.IsActive = true }
func (r *Role) Deactivate() { r.IsActive = false }

// Rename updates the role's display name. System roles cannot be
// renamed.
func (r *Role) Rename(name valueobjects.FullName) error {
	if r.Type == enums.RoleTypeSystem {
		return errors.Wrap(errors.ErrInvalidStateTransition, errField("system roles cannot be renamed"))
	}
	r.Name = name
	return nil
}
