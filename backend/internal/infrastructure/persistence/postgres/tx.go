package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/internal/domain/repositories"
)

// TxManager is the PostgreSQL implementation of
// repositories.TransactionManager. It uses the existing database.DB
// wrapper (which itself wraps *sql.DB) and the standard SQL
// transaction API.
type TxManager struct {
	db *database.DB
}

// NewTxManager returns a new transaction manager.
func NewTxManager(db *database.DB) *TxManager {
	return &TxManager{db: db}
}

// WithinTransaction begins a transaction, calls fn, and commits on
// success. If fn returns an error, the transaction is rolled back
// and the error is propagated. pgx errors are translated to
// domain-friendly errors via Translate.
func (m *TxManager) WithinTransaction(ctx context.Context, fn repositories.TxRunner) error {
	return m.db.WithTx(ctx, func(tx *database.Tx) error {
		tb := &txBox{tx: tx.Tx}
		uow := newUnitOfWork(tb)
		txCtx := repositories.ContextWithUnitOfWork(ctx, uow)
		if err := fn(txCtx); err != nil {
			return Translate(err)
		}
		return nil
	})
}

// txBox is the small wrapper that exposes the database/sql Tx
// methods a repository needs, so that the same repository code can
// run against *sql.DB or *sql.Tx without modification.
type txBox struct {
	tx *sql.Tx
}

func (t *txBox) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *txBox) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *txBox) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// dbBox wraps *sql.DB for repositories that run outside a transaction.
type dbBox struct {
	db *sql.DB
}

func (d *dbBox) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *dbBox) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *dbBox) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// Querier is the common interface implemented by txBox and dbBox. The
// repository implementations take a Querier so they can run against
// either the connection pool (auto-commit) or a transaction.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// errf wraps err so callers can use it in fmt.Errorf chains without
// re-importing fmt.
func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
