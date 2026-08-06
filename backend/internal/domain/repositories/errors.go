package repositories

import (
	"context"
	"errors"

	derrors "vfinancy/backend/internal/domain/errors"
)

// Use the errors package's New function (which produces a DomainError).
// Aliased to avoid clashing with the standard library's errors.
var newDomainError = derrors.New

// Repository-level errors. These are the ONLY errors a repository
// implementation is allowed to return. Application code maps these to
// UI messages or HTTP status codes. PostgreSQL-specific errors are
// translated to these in the infrastructure layer.
//
// Each sentinel is a DomainError so callers can match by stable code
// (IsCode/IsAnyCode) regardless of the message.
var (
	// ErrNotFound is returned when an aggregate cannot be located by
	// the requested identifier.
	ErrNotFound = newDomainError("NOT_FOUND", "repositories: not found")

	// ErrDuplicate is returned when a unique constraint is violated
	// (e.g. inserting a customer with an existing document number).
	ErrDuplicate = newDomainError("DUPLICATE", "repositories: duplicate record")

	// ErrForeignKey is returned when a foreign-key constraint is
	// violated (e.g. inserting a sale for a non-existent customer).
	ErrForeignKey = newDomainError("FOREIGN_KEY_VIOLATION", "repositories: foreign key violation")

	// ErrCheckConstraint is returned when a CHECK constraint is
	// violated (e.g. a manual SQL insert that bypassed domain
	// validation).
	ErrCheckConstraint = newDomainError("CHECK_CONSTRAINT_VIOLATION", "repositories: check constraint")

	// ErrConnection is returned when the database is unreachable or
	// the connection is reset.
	ErrConnection = newDomainError("CONNECTION_FAILURE", "repositories: connection failure")

	// ErrTx is returned when a transaction fails to commit or
	// rollback. The application layer treats this as a hard error
	// and surfaces it to the caller.
	ErrTx = newDomainError("TX_ERROR", "repositories: transaction error")
)

// isContextError reports whether the error is a context cancellation
// or deadline. Such errors should not be mapped to ErrConnection.
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
