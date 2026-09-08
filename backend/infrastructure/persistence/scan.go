package persistence

import (
	"database/sql"

	"vfinancy/backend/internal/domain/repositories"
)

// ScanRow scans a single row into dest using the stdlib Scan API. It
// is intentionally minimal: callers pass the columns they need and a
// destination matching the column order. NotFound rows are translated
// to repositories.ErrNotFound on PostgreSQL.
func ScanRow(row *sql.Row, dest ...any) error {
	if err := row.Scan(dest...); err != nil {
		if IsPgNoRows(err) {
			return repositories.ErrNotFound
		}
		return Translate(err)
	}
	return nil
}

// ScanRows scans every row in a result set. The callback is invoked
// for each row. The first error from the callback or the rows is
// returned; rows.Err() is checked at the end.
func ScanRows(rows *sql.Rows, fn func(*sql.Rows) error) error {
	defer rows.Close()
	for rows.Next() {
		if err := fn(rows); err != nil {
			return err
		}
	}
	return Translate(rows.Err())
}

// InClause builds the placeholder list for an IN (?, ?, ?, ...) clause
// given a count. Returns "NULL" for a count of zero or less.
func InClause(n int) string {
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

// LimitOffset returns the LIMIT and OFFSET for a PageRequest, applying
// a sane maximum to prevent runaway queries.
func LimitOffset(p repositories.PageRequest, defaultLimit, maxLimit int) (limit, offset int) {
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
