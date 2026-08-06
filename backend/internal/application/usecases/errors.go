// Package usecases is the application layer. It contains the use
// cases that orchestrate business workflows: each use case is a
// typed function that takes a Request, runs the relevant services
// inside a transaction, and returns a Response.
//
// Conventions
//
//   * Each use case is a struct with constructor-injected dependencies.
//   * The single entry point is the Execute method:
//
//         func (uc *CreateSaleUseCase) Execute(ctx context.Context, req CreateSaleRequest) (*CreateSaleResponse, error)
//
//   * Multi-step operations run inside a single transaction. Rollback
//     is automatic on any error.
//   * Use cases never contain business rules. They validate
//     workflow consistency (required fields, existence) and delegate
//     to services for everything else.
//   * Use cases never touch SQL or pgx. They talk to services.
//   * Use cases log workflow start/finish events with the input
//     identifiers and a duration.
//   * Errors are application-level. They wrap service/domain errors
//     and are returned as the second value of Execute.
package usecases

import (
	derrors "vfinancy/backend/internal/domain/errors"
)

// Application-level workflow errors. These are mapped from the
// underlying domain / service errors by the use case.
//
// We use the domain error sentinel pattern so the application /
// interface layer can match by code with errors.IsCode.
var (
	// ErrValidation is returned when the request is missing
	// required fields or has an inconsistent shape.
	ErrValidation = derrors.New("VALIDATION", "usecases: invalid request")

	// ErrNotFound is returned when a referenced entity does not
	// exist. Wraps repositories.ErrNotFound.
	ErrNotFound = derrors.Wrap(derrors.ErrNotFound, nil)

	// ErrConflict is returned when a workflow step would violate
	// a business invariant (e.g. trying to approve a paid purchase).
	ErrConflict = derrors.New("CONFLICT", "usecases: workflow conflict")

	// ErrUnauthorized is returned when the user is not allowed to
	// perform the workflow. Phase 1.5 leaves the body empty; Phase
	// 2 will replace it with the RBAC check.
	ErrUnauthorized = derrors.New("UNAUTHORIZED", "usecases: not allowed")

	// ErrInternal wraps unexpected errors from the lower layers.
	// The original error is wrapped for debugging.
	ErrInternal = derrors.New("INTERNAL", "usecases: internal error")
)
