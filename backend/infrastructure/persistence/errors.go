// Package postgres is the PostgreSQL implementation of the
// repositories defined in internal/domain/repositories. The package
// is the ONLY place that imports pgx / database/sql and is the ONLY
// place that knows the schema column names.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"vfinancy/backend/internal/domain/repositories"
)

// pgErrorCode is the SQLSTATE class that distinguishes the kind of
// constraint violation or warning the database is reporting. We only
// care about a handful of codes for now.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
	pgNotNullViolation    = "23502"
	pgSerializationFail   = "40001"
	pgDeadlock            = "40P01"
)

// Translate maps a pgx / database/sql error to a domain-friendly error.
// Returns nil if the input is nil; returns the original error
// unchanged if it is neither a *pgconn.PgError nor a known database
// network error.
//
// Mapping rules:
//   * 23505 unique_violation    → repositories.ErrDuplicate
//   * 23503 foreign_key_violation → repositories.ErrForeignKey
//   * 23514 check_violation      → repositories.ErrCheckConstraint
//   * 23502 not_null_violation   → repositories.ErrCheckConstraint
//   * 40001 / 40P01              → wrapped repositories.ErrTx
//   * network/connection errors  → repositories.ErrConnection
//   * context cancellation       → unchanged (caller decides)
//   * everything else             → unchanged
//
// SQLite (modernc.org/sqlite) reports constraint failures as plain
// strings ("constraint failed: FOREIGN KEY constraint failed (787)");
// those are matched by message text and mapped to the same sentinels.
func Translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case pgUniqueViolation:
			return fmt.Errorf("%w: %s", repositories.ErrDuplicate, pg.Message)
		case pgForeignKeyViolation:
			return fmt.Errorf("%w: %s", repositories.ErrForeignKey, pg.Message)
		case pgCheckViolation, pgNotNullViolation:
			return fmt.Errorf("%w: %s", repositories.ErrCheckConstraint, pg.Message)
		case pgSerializationFail, pgDeadlock:
			return fmt.Errorf("%w: %s", repositories.ErrTx, pg.Message)
		default:
			return err
		}
	}
	if errors.Is(err, sql.ErrConnDone) || errors.Is(err, sql.ErrTxDone) {
		return fmt.Errorf("%w: %s", repositories.ErrConnection, err.Error())
	}
	if code := sqliteConstraintCode(err); code != nil {
		return fmt.Errorf("%w: %s", code, err.Error())
	}
	return err
}

// sqliteConstraintCode maps a SQLite constraint error (modernc.org/sqlite)
// to the corresponding repository sentinel. SQLite reports constraint
// failures as plain error strings such as
// "constraint failed: FOREIGN KEY constraint failed (787)", so the
// message text is matched here. Returns nil for non-SQLite errors.
func sqliteConstraintCode(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return repositories.ErrForeignKey
	case strings.Contains(msg, "UNIQUE constraint failed"):
		return repositories.ErrDuplicate
	case strings.Contains(msg, "CHECK constraint failed"),
		strings.Contains(msg, "NOT NULL constraint failed"):
		return repositories.ErrCheckConstraint
	default:
		return nil
	}
}

// isPgNoRows reports whether the error is the standard "no rows in
// result set" error. Repositories return repositories.ErrNotFound
// in that case; this helper centralises the test.
func IsPgNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
