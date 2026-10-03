package persistence

import "sync"

// Dialect identifies the SQL dialect a *sql.DB is running against.
// The repository layer is dialect-portable by construction ($1
// placeholders, TRUE/FALSE literals and the types in the shared
// migrations all work on both engines); this tiny bootstrap value is
// only needed by the rare statement that has no portable spelling
// (e.g. text[] literals, which only PostgreSQL supports).
type Dialect string

const (
	DialectPostgres Dialect = "postgres"
	DialectSQLite   Dialect = "sqlite"
)

var (
	dialectMu      sync.RWMutex
	defaultDialect = DialectPostgres
)

// SetDialect records which engine the default *sql.DB belongs to. It
// must be called once at process bootstrap (app init, CLI, sync
// server) before any repository runs queries that consult the dialect.
func SetDialect(d Dialect) {
	dialectMu.Lock()
	defer dialectMu.Unlock()
	defaultDialect = d
}

// CurrentDialect returns the configured dialect.
func CurrentDialect() Dialect {
	dialectMu.RLock()
	defer dialectMu.RUnlock()
	return defaultDialect
}

// IsPostgres reports whether the process is running against PostgreSQL.
func IsPostgres() bool { return CurrentDialect() == DialectPostgres }

// IsSQLite reports whether the process is running against SQLite.
func IsSQLite() bool { return CurrentDialect() == DialectSQLite }

// Num casts a stored decimal column to a number so aggregates can run
// over it. Quantities and amounts are stored as decimal strings to
// avoid float drift in Go; casting inside SQL is only used for sums and
// comparisons the caller never routes through Money.
func Num(column string) string {
	if IsSQLite() {
		return "CAST(" + column + " AS REAL)"
	}
	return "CAST(" + column + " AS numeric)"
}

// Year and Month extract the calendar year and the 1-based month of a
// timestamp column for grouping. Timestamp storage differs per dialect:
// SQLite holds epoch milliseconds, PostgreSQL a native timestamptz.
func Year(column string) string {
	if IsSQLite() {
		return "CAST(strftime('%Y', " + column + " / 1000, 'unixepoch') AS INTEGER)"
	}
	return "CAST(EXTRACT(YEAR FROM " + column + ") AS INTEGER)"
}

// Month is the 1-based month counterpart of Year.
func Month(column string) string {
	if IsSQLite() {
		return "CAST(strftime('%m', " + column + " / 1000, 'unixepoch') AS INTEGER)"
	}
	return "CAST(EXTRACT(MONTH FROM " + column + ") AS INTEGER)"
}
