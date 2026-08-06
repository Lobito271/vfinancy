// Typed errors raised by the auth services. Codes are stable so the
// application layer can match them with errors.IsCode without
// inspecting message strings.
package auth

import (
	derrors "vfinancy/backend/internal/domain/errors"
)

var (
	// ErrAuthFailed is returned when credentials are rejected. The
	// message is deliberately generic so callers cannot tell whether
	// the username or the password was wrong.
	ErrAuthFailed = derrors.New("AUTH_FAILED", "invalid credentials")

	// ErrAuthInactive is returned when the user account is disabled.
	ErrAuthInactive = derrors.New("AUTH_INACTIVE", "user account is inactive")

	// ErrAuthLocked is returned when the account is temporarily
	// locked after too many failed attempts.
	ErrAuthLocked = derrors.New("AUTH_LOCKED", "user account is locked")

	// ErrAuthNotFound is returned when a user referenced by ID does
	// not exist.
	ErrAuthNotFound = derrors.New("AUTH_NOT_FOUND", "user not found")

	// ErrSessionInvalid is returned when a session token exists but
	// is inactive or expired.
	ErrSessionInvalid = derrors.New("SESSION_INVALID", "session is invalid or expired")

	// ErrSessionNotFound is returned when no session matches the
	// given token.
	ErrSessionNotFound = derrors.New("SESSION_NOT_FOUND", "session not found")

	// ErrSessionFailed wraps unexpected session-operation failures.
	ErrSessionFailed = derrors.New("SESSION_FAILED", "session operation failed")

	// ErrPasswordWeak is returned when a new password fails the
	// strength requirements. Code INVALID_FORMAT maps to a
	// validation error in the application layer.
	ErrPasswordWeak = derrors.New("INVALID_FORMAT", "password does not meet strength requirements")
)
