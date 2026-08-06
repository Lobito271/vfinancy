package identity

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// User is a system user. Passwords are stored as Argon2id hashes
// produced by the application layer; the domain never sees the
// plaintext.
type User struct {
	ID                uuid.UUID
	CompanyID         uuid.UUID
	DefaultBranchID   *uuid.UUID
	Username          valueobjects.ShortCode
	Email             valueobjects.Email
	FullName          valueobjects.FullName
	PasswordHash      string
	MustChangePassword bool
	FailedLoginAttempts int
	LockedUntil       *time.Time
	LastLoginAt       *time.Time
	LastLoginIP       string
	IsActive          bool
	roles             []RoleAssignment
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	CreatedBy         *uuid.UUID
	UpdatedBy         *uuid.UUID
}

// RoleAssignment captures a (role, optional branch, optional expiry)
// tuple assigned to a user. The branch_id is nil for company-wide
// grants. The expires_at is nil for non-expiring grants.
type RoleAssignment struct {
	RoleID    uuid.UUID
	BranchID  *uuid.UUID
	ExpiresAt *time.Time
}

// NewUserOptions is the input to NewUser.
type NewUserOptions struct {
	CompanyID       uuid.UUID
	DefaultBranchID *uuid.UUID
	Username        valueobjects.ShortCode
	Email           valueobjects.Email
	FullName        valueobjects.FullName
	PasswordHash    string
}

// NewUser constructs a User. The user starts active with no failed
// attempts and must_change_password set to true (the application should
// force a password reset on the first login).
func NewUser(now time.Time, opts NewUserOptions) (*User, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if opts.PasswordHash == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("password hash is required"))
	}
	return &User{
		ID:                  uuid.New(),
		CompanyID:           opts.CompanyID,
		DefaultBranchID:     opts.DefaultBranchID,
		Username:            opts.Username,
		Email:               opts.Email,
		FullName:            opts.FullName,
		PasswordHash:        opts.PasswordHash,
		MustChangePassword:  true,
		FailedLoginAttempts: 0,
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// Status returns the current lifecycle status of the user. Inactive
// is the dominant state (an admin explicitly disabled the user) and
// takes precedence over the temporary lockout window.
func (u *User) Status(now time.Time) enums.UserStatus {
	if !u.IsActive {
		return enums.UserStatusInactive
	}
	if u.LockedUntil != nil && u.LockedUntil.After(now) {
		return enums.UserStatusLocked
	}
	return enums.UserStatusActive
}

// IsLocked reports whether the user is currently locked out.
func (u *User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && u.LockedUntil.After(now)
}

// RecordSuccessfulLogin resets the failed-attempt counter and updates
// the last-login metadata. Called by the application after a successful
// authentication.
func (u *User) RecordSuccessfulLogin(at time.Time, ip string) {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.LastLoginAt = &at
	u.LastLoginIP = ip
	u.MustChangePassword = false
}

// RecordFailedLogin increments the failed-attempt counter. When the
// counter reaches maxAttempts, the user is locked out for the given
// duration. Returns the new failed-attempt count.
func (u *User) RecordFailedLogin(maxAttempts int, lockoutDuration time.Duration, at time.Time) int {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= maxAttempts {
		until := at.Add(lockoutDuration)
		u.LockedUntil = &until
	}
	return u.FailedLoginAttempts
}

// Unlock clears the lockout window. Used by the application when an
// admin resets a user's password or manually unblocks the account.
func (u *User) Unlock() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
}

// Activate / Deactivate.
func (u *User) Activate()   { u.IsActive = true }
func (u *User) Deactivate() { u.IsActive = false }

// ChangeEmail updates the user's email address. Username and email
// are immutable in practice for compliance reasons; the application
// layer should never expose this method to end users.
func (u *User) ChangeEmail(email valueobjects.Email) {
	u.Email = email
}

// ChangePassword replaces the password hash. Resets must_change_password.
func (u *User) ChangePassword(hash string) {
	u.PasswordHash = hash
	u.MustChangePassword = false
}

// RequirePasswordChange flags the user so the next login forces a
// password reset. Used after an admin-triggered password reset.
func (u *User) RequirePasswordChange() {
	u.MustChangePassword = true
}

// AssignRole grants a role to the user. An optional branch scopes the
// grant; a nil branch means the grant is company-wide. An optional
// expiry makes the grant time-bound.
func (u *User) AssignRole(roleID uuid.UUID, branchID *uuid.UUID, expiresAt *time.Time) {
	u.roles = append(u.roles, RoleAssignment{
		RoleID:    roleID,
		BranchID:  branchID,
		ExpiresAt: expiresAt,
	})
}

// RevokeRole removes a role assignment. If branchID is nil, only
// company-wide assignments are removed; otherwise only the matching
// branch-scoped one.
func (u *User) RevokeRole(roleID uuid.UUID, branchID *uuid.UUID) {
	out := u.roles[:0]
	for _, r := range u.roles {
		if r.RoleID == roleID && sameUUIDPointer(r.BranchID, branchID) {
			continue
		}
		out = append(out, r)
	}
	u.roles = out
}

// EffectiveRoles returns the list of role IDs the user currently
// holds, after applying expiry. Expired grants are filtered out.
func (u *User) EffectiveRoles(now time.Time) []RoleAssignment {
	out := make([]RoleAssignment, 0, len(u.roles))
	for _, r := range u.roles {
		if r.ExpiresAt != nil && !r.ExpiresAt.After(now) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func sameUUIDPointer(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
