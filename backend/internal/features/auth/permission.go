package auth

import "time"

// Permission is a single grantable action in the global catalog. It is
// a value object on the wire (Permission.code is the FK), so this type
// is intentionally minimal: the authoritative list lives in the
// database and is seeded by an SQL migration.
type Permission struct {
	Code        string
	Module      string
	Action      string
	Description string
	CreatedAt   time.Time
}
