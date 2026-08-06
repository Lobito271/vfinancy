package services

import (
	"context"
	"fmt"

	"vfinancy/backend/internal/domain/repositories"
)

// TxRunner is the application-side callback signature for transactional
// work. The application receives a UoW-bearing context and may call any
// repository or service whose access is mediated by the UoW. Returning
// a non-nil error rolls back the transaction.
type TxRunner func(ctx context.Context) error

// WithTransaction runs fn inside a database transaction. It is a
// thin adapter over repositories.TransactionManager that the services
// depend on, so the service code reads as `s.txm.WithinTransaction(...)`
// rather than `s.txManager.WithinTransaction(...)`.
type TxManager interface {
	WithinTransaction(ctx context.Context, fn TxRunner) error
}

// WithTransactionAdapter bridges repositories.TransactionManager to
// the application-side TxManager interface.
type WithTransactionAdapter struct {
	inner repositories.TransactionManager
}

// NewTxManager wraps a repositories.TransactionManager.
func NewTxManager(inner repositories.TransactionManager) *WithTransactionAdapter {
	if inner == nil {
		panic("services: nil txManager")
	}
	return &WithTransactionAdapter{inner: inner}
}

// WithinTransaction implements TxManager.
func (a *WithTransactionAdapter) WithinTransaction(ctx context.Context, fn TxRunner) error {
	return a.inner.WithinTransaction(ctx, func(ctx context.Context) error {
		return fn(ctx)
	})
}

// EnsureError formats an error with a fixed code for the rare cases
// where the application layer wants to construct a new domain error
// at runtime (e.g. aggregating several underlying failures).
func EnsureError(code, format string, args ...any) error {
	return fmt.Errorf("%s: %s", code, fmt.Sprintf(format, args...))
}
