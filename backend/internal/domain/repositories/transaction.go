package repositories

import "context"

// TxRunner is the closure invoked inside a transaction. The closure
// receives a context.Context that carries the active transaction;
// services and repositories use the same context so the entire
// workflow stays on a single transaction.
type TxRunner func(ctx context.Context) error

// TransactionManager begins, commits and rolls back transactions. The
// implementation lives in the infrastructure layer.
//
// The interface is defined here so the application layer can take a
// dependency on it without importing any infrastructure package.
type TransactionManager interface {
	// WithinTransaction begins a new transaction, calls fn, and
	// commits on success. If fn returns an error, the transaction
	// is rolled back and the error is propagated. If commit fails
	// the error is wrapped with ErrTx.
	WithinTransaction(ctx context.Context, fn TxRunner) error
}
