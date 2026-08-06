package persistence

import (
	"context"
	"database/sql"
	"fmt"
)

// Querier is the common interface implemented by txBox and dbBox. The
// repository implementations take a Querier so they can run against
// either the connection pool (auto-commit) or a transaction.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
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

// FromDB returns a Querier bound to the connection pool (auto-commit).
func FromDB(db *sql.DB) Querier {
	return &dbBox{db: db}
}

// errf wraps err so callers can use it in fmt.Errorf chains without
// re-importing fmt.
func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
