package persistence

import (
	"context"

	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/internal/domain/repositories"
)

// TxManager is the implementation of repositories.TransactionManager.
// It uses the database.DB wrapper (which itself wraps *sql.DB) and the
// standard SQL transaction API.
type TxManager struct {
	db *database.DB
}

// NewTxManager returns a new transaction manager.
func NewTxManager(db *database.DB) *TxManager {
	return &TxManager{db: db}
}

// WithinTransaction begins a transaction, calls fn, and commits on
// success. The context passed to fn carries the transaction-bound
// Querier so every repository that routes through persistence.Q runs
// its SQL on the transaction. If fn returns an error, the transaction
// is rolled back and the error is propagated. pgx errors are
// translated to domain-friendly errors via Translate.
//
// When fn is invoked while a transaction is already active in ctx, it
// joins that transaction instead of opening a new one: workflows that
// compose several service calls keep the whole operation on a single
// database transaction.
func (m *TxManager) WithinTransaction(ctx context.Context, fn repositories.TxRunner) error {
	if TxQuerierFromContext(ctx) != nil {
		return fn(ctx)
	}
	return m.db.WithTx(ctx, func(tx *database.Tx) error {
		ctx = WithTxQuerier(ctx, &txBox{tx: tx.Tx})
		if err := fn(ctx); err != nil {
			return Translate(err)
		}
		return nil
	})
}
