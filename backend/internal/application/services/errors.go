// Package services is the application service layer. It centralizes
// all business logic: input validation, state transitions, money
// math, journal-entry generation, and cross-aggregate orchestration.
// Repositories are dumb persistence; services are the brain.
//
// Conventions
//
//   * Every service exposes one public method per business
//     operation. The method name is a verb in the present tense:
//     Create, Update, Cancel, MarkAsPaid, etc.
//   * Input is a typed struct (Request / Input) and output is a
//     typed entity plus error. No positional arguments.
//   * Multi-step work runs inside a single transaction via
//     repositories.TransactionManager.WithinTransaction. A failure
//     in any step rolls back the whole operation.
//   * Errors are typed via the sentinel values in this package. The
//     application / interface layer can match by code.
//   * Services log structured events for every business outcome
//     (sale created, payment received, journal posted, etc.).
//   * Services never touch pgx, sql, http, or anything below
//     repositories; repositories never touch entities other than to
//     translate them.
package services

import (
	derrors "vfinancy/backend/internal/domain/errors"
)

// Service-level sentinel errors. Each has a stable Code() for matching.
//
// These wrap the domain's typed errors (ErrCustomerInactive,
// ErrPaymentExceedsBalance, ...) but expose them at a coarser
// granularity so the UI can show workflow-level messages without
// inspecting the wrapped error in most cases.
var (
	// ErrInsufficientStock — the requested operation would leave a
	// (product, warehouse) below zero quantity.
	ErrInsufficientStock = derrors.Wrap(derrors.ErrInsufficientStock, nil)

	// ErrInvalidPayment — the payment amount is zero, negative, or
	// otherwise unusable in the current workflow state.
	ErrInvalidPayment = derrors.Wrap(derrors.ErrInvalidPayment, nil)

	// ErrCustomerBlocked — the customer is in a state (blocked or
	// inactive) that disallows the requested operation.
	ErrCustomerBlocked = derrors.Wrap(derrors.ErrCustomerInactive, nil)

	// ErrSupplierBlocked — same as ErrCustomerBlocked but for suppliers.
	ErrSupplierBlocked = derrors.Wrap(derrors.ErrSupplierInactive, nil)

	// ErrInventoryUnavailable — the inventory row is missing,
	// soft-deleted, or in a status that disallows mutations.
	ErrInventoryUnavailable = derrors.New("INVENTORY_UNAVAILABLE", "inventory row is not available")

	// ErrJournalUnbalanced — the entry would not satisfy
	// SUM(debit) == SUM(credit).
	ErrJournalUnbalanced = derrors.Wrap(derrors.ErrUnbalancedJournalEntry, nil)

	// ErrDuplicatePurchase — a purchase has been approved twice.
	ErrDuplicatePurchase = derrors.New("DUPLICATE_PURCHASE", "purchase has already been approved")

	// ErrSaleCancelled — a payment is being applied to a cancelled
	// sale, or a similar operation on a terminal-state sale.
	ErrSaleCancelled = derrors.New("SALE_CANCELLED", "sale is cancelled")

	// ErrPurchaseCancelled — symmetric to ErrSaleCancelled for the
	// purchasing workflow.
	ErrPurchaseCancelled = derrors.Wrap(derrors.ErrPurchaseCancelled, nil)

	// ErrNegativeStock — an explicit guard raised by the inventory
	// service when an adjustment would drive the running balance
	// below zero. Distinct from ErrInsufficientStock, which is
	// raised by the sales workflow when reserving stock.
	ErrNegativeStock = derrors.New("NEGATIVE_STOCK", "stock would become negative")

	// ErrInventoryLocked — a batch in a non-active state (depleted
	// or written off) cannot be adjusted or consumed.
	ErrInventoryLocked = derrors.New("INVENTORY_LOCKED", "inventory batch is locked")

	// ErrInsufficientAdvance — a customer advance application would
	// drive the advance balance below zero.
	ErrInsufficientAdvance = derrors.Wrap(derrors.ErrAdvanceNegative, nil)
)
