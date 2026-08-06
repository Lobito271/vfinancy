package postgres

import (
	"database/sql"

	"vfinancy/backend/internal/domain/repositories"
)

// Common helpers shared by all repository implementations.

// scanRow scans a single row into a dest using the stdlib Scan API.
// It is intentionally minimal: callers pass the columns they need
// and a destination matching the column order.
func scanRow(row *sql.Row, dest ...any) error {
	if err := row.Scan(dest...); err != nil {
		if isPgNoRows(err) {
			return repositories.ErrNotFound
		}
		return Translate(err)
	}
	return nil
}

// scanRows scans every row in a result set. The callback is invoked
// for each row. The first error from the callback or the rows is
// returned; rows.Err() is checked at the end.
func scanRows(rows *sql.Rows, fn func(*sql.Rows) error) error {
	defer rows.Close()
	for rows.Next() {
		if err := fn(rows); err != nil {
			return err
		}
	}
	return Translate(rows.Err())
}

// inClause builds a placeholder list for an IN (?, ?, ?, ...) clause
// given a count. Returns the placeholder string and the bind
// arguments (nil) — callers pass their own args separately.
func inClause(n int) string {
	if n <= 0 {
		// Postgres requires at least one placeholder even if the
		// caller filters it out before calling; "IN ()" is a syntax
		// error. We use a dummy "NULL" predicate that is always
		// false and matches no rows.
		return "NULL"
	}
	out := make([]byte, 0, 2*n)
	for i := 0; i < n; i++ {
		if i > 0 {
			out = append(out, ',')
		}
		out = append(out, '?')
	}
	return string(out)
}

// limitOffset returns LIMIT and OFFSET strings for a PageRequest,
// applying a sane maximum to prevent runaway queries.
func limitOffset(p repositories.PageRequest, defaultLimit, maxLimit int) (limit, offset int) {
	limit = p.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset = p.Offset
	if offset < 0 {
		offset = 0
	}
	return
}
