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
	"errors"
	"fmt"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
)

// Application-level service errors. These are returned by the feature
// services and mapped to UI messages by the bindings layer.
var (
	// ErrValidation is returned when the request is missing
	// required fields or has an inconsistent shape.
	ErrValidation = derrors.New("VALIDATION", "services: invalid request")

	// ErrNotFound is returned when a referenced entity does not
	// exist. Wraps repositories.ErrNotFound.
	ErrNotFound = derrors.Wrap(repositories.ErrNotFound, nil)

	// ErrConflict is returned when an operation would violate a
	// business invariant (e.g. trying to approve a paid purchase).
	ErrConflict = derrors.New("CONFLICT", "services: conflict")

	// ErrInternal wraps unexpected errors from the lower layers.
	// The original error is wrapped for debugging.
	ErrInternal = derrors.New("INTERNAL", "services: internal error")

	// ErrCustomerBlocked — the customer is in a state (blocked or
	// inactive) that disallows the requested operation.
	ErrCustomerBlocked = derrors.Wrap(derrors.ErrCustomerInactive, nil)

	// ErrSupplierBlocked — symmetric to ErrCustomerBlocked for suppliers.
	ErrSupplierBlocked = derrors.Wrap(derrors.ErrSupplierInactive, nil)
)

// Errorf wraps an application-level sentinel with a message, preserving
// the error chain: fmt.Errorf("%w: %s", base, msg). Services use it to
// attach a field hint or request detail to a sentinel before returning.
func Errorf(base error, msg string) error {
	return fmt.Errorf("%w: %s", base, msg)
}

// MapError translates service/domain errors into the appropriate
// application-level sentinel. A service calls this on the final return
// value of a multi-step operation before returning to the bindings.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) || derrors.IsCode(err, "NOT_FOUND") {
		return ErrNotFound
	}
	if derrors.IsCode(err, "DUPLICATE") {
		return fmt.Errorf("%w: %s", ErrConflict, err.Error())
	}
	if derrors.IsCode(err, "INSUFFICIENT_STOCK") ||
		derrors.IsCode(err, "INVALID_PAYMENT") ||
		derrors.IsCode(err, "NEGATIVE_QUANTITY") ||
		derrors.IsCode(err, "NEGATIVE_MONEY") ||
		derrors.IsCode(err, "PURCHASE_CANCELLED") ||
		derrors.IsCode(err, "SALE_ALREADY_PAID") ||
		derrors.IsCode(err, "ALREADY_RECEIVED") ||
		derrors.IsCode(err, "ALREADY_RECONCILED") {
		return fmt.Errorf("%w: %s", ErrConflict, err.Error())
	}
	if derrors.IsCode(err, "REQUIRED") || derrors.IsCode(err, "INVALID_FORMAT") {
		return fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}
	return wrapAsInternal(err)
}

// wrapAsInternal wraps a low-level error as an internal error. The
// original is preserved via derrors.Wrap so callers can errors.As the
// underlying cause.
func wrapAsInternal(err error) error {
	if err == nil {
		return nil
	}
	return derrors.Wrap(ErrInternal, err)
}
