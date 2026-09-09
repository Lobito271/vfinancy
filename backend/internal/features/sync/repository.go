package sync

import (
	"context"
	"io/fs"
	"time"
)

// LocalStore reads and writes the replication bookkeeping and business
// rows of the local SQLite runtime database. Row payloads carry native
// Go values: time.Time timestamps, string uuids, dates and money, bool
// flags, int64 integers and nil for NULL.
type LocalStore interface {
	// Cursor returns the watermark of key ("<table>:push" or
	// "<table>:pull"), or the zero time when the cursor does not
	// exist.
	Cursor(ctx context.Context, key string) (time.Time, error)
	// SetCursor advances the watermark of key; it never moves
	// backwards.
	SetCursor(ctx context.Context, key string, at time.Time) error
	// Changes returns every row of table whose time column is newer
	// than since, ordered by that column ascending, including rows
	// whose deleted_at is set (soft deletes are rows, not tombstones).
	Changes(ctx context.Context, meta TableMeta, since time.Time) ([]map[string]any, error)
	// Tombstones returns the pending hard-delete markers of table.
	Tombstones(ctx context.Context, table string) ([]Tombstone, error)
	// UpsertRows applies pulled rows inside one transaction with
	// last-writer-wins: a row is skipped when it was deleted locally
	// (an unpushed tombstone exists) or when the local copy is strictly
	// newer. Divergences over existing rows are recorded in
	// sync_conflicts and returned.
	UpsertRows(ctx context.Context, meta TableMeta, rows []map[string]any) ([]Conflict, error)
	// ApplyDeletes removes the given primary keys from table inside
	// one transaction.
	ApplyDeletes(ctx context.Context, meta TableMeta, ids []string) error
	// ForgetTombstones removes tombstones that have been replicated to
	// the mirror.
	ForgetTombstones(ctx context.Context, table string, ids []string) error
}

// RemoteStore reads and writes the cloud PostgreSQL mirror over a
// direct pgx connection. The mirror has a single writer (this
// desktop install), so writes are unconditional: deletes flow one way,
// from the local database to the mirror.
type RemoteStore interface {
	// EnsureSchema connects to the mirror and applies pending
	// migrations from migrationsFS (the embedded
	// backend/migrations/postgres content).
	EnsureSchema(ctx context.Context, migrationsFS fs.FS) error
	// Changes returns every row of table whose time column is newer
	// than since, ordered ascending, in the same value shape as
	// LocalStore.Changes.
	Changes(ctx context.Context, meta TableMeta, since time.Time) ([]map[string]any, error)
	// UpsertRows writes rows into the mirror inside one transaction.
	UpsertRows(ctx context.Context, meta TableMeta, rows []map[string]any) error
	// ApplyDeletes removes the given primary keys from the mirror
	// inside one transaction.
	ApplyDeletes(ctx context.Context, meta TableMeta, ids []string) error
	// Close releases the mirror connection pool.
	Close() error
}
