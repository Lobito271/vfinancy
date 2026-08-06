package enums

// RoleType is the broad classification of a role. System roles (admin,
// manager, etc.) cannot be edited or deleted; custom roles can be.
type RoleType string

const (
	// RoleTypeSystem — seeded by the application, protected.
	RoleTypeSystem RoleType = "system"
	// RoleTypeCustom — created by an administrator at runtime.
	RoleTypeCustom RoleType = "custom"
)

func (r RoleType) Valid() bool {
	return r == RoleTypeSystem || r == RoleTypeCustom
}

func (r RoleType) String() string { return string(r) }
