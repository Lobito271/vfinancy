// Package apperrors holds shared application-layer error sentinels and
// helpers used across the feature services.
//
// The per-service codes used in the services (REQUIRED, EMPTY_DOCUMENT,
// INVALID_PAYMENT, ...) are constructed as typed errors via
// domain/errors.New so errors.IsCode can match them; this package only
// carries the sentinels and helpers that are shared verbatim by more
// than one service.
package apperrors

import (
	"fmt"

	derrors "vfinancy/backend/internal/domain/errors"
)

// Application-level service errors. These are returned by the feature
// services and mapped to UI messages by the bindings layer.
var (
	// ErrValidation is returned when the request is missing
	// required fields or has an inconsistent shape.
	ErrValidation = derrors.New("VALIDATION", "services: invalid request")

	// ErrConflict is returned when an operation would violate a
	// business invariant (e.g. trying to approve a paid purchase).
	ErrConflict = derrors.New("CONFLICT", "services: conflict")

	// ErrCustomerBlocked — the customer is in a state (blocked or
	// inactive) that disallows the requested operation.
	ErrCustomerBlocked = derrors.Wrap(derrors.ErrCustomerInactive, nil)
)

// Errorf wraps an application-level sentinel with a message, preserving
// the error chain: fmt.Errorf("%w: %s", base, msg). Services use it to
// attach a field hint or request detail to a sentinel before returning.
func Errorf(base error, msg string) error {
	return fmt.Errorf("%w: %s", base, msg)
}
