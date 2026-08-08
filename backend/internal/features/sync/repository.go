package sync

import (
	"context"
)

// Repository is the persistence contract for the sync engine. The
// concrete implementation in sync/postgres runs identically against the
// local SQLite database and the cloud PostgreSQL mirror: all queries
// use $N placeholders and the repository branches on
// persistence.IsSQLite() only for time-argument encoding.
type Repository interface {
	// --- devices ---
	GetLocalDevice(ctx context.Context) (*Device, error)
	GetDeviceByToken(ctx context.Context, token string) (*Device, error)
	RegisterDevice(ctx context.Context, d *Device) error
	TouchDeviceSeen(ctx context.Context, deviceID string, at int64) error

	// --- cursors ---
	GetCursors(ctx context.Context, deviceID string) (map[string]int64, error)
	UpdateCursors(ctx context.Context, deviceID string, cursors map[string]int64) error

	// --- conflicts ---
	LogConflict(ctx context.Context, c *SyncConflict) error

	// --- row access (generic, registry-driven) ---
	// RowsChangedSince returns every row of table whose TimeColumn is
	// strictly greater than after (ms epoch), payloads built via the
	// table metadata, plus the new watermark.
	RowsChangedSince(ctx context.Context, meta *TableMeta, after int64) ([]Change, int64, error)
	// TombstonesSince returns hard-delete markers of table newer than
	// after (ms epoch).
	TombstonesSince(ctx context.Context, table string, after int64) ([]Tombstone, error)
	// FetchTime returns the TimeColumn value of a row, or 0 if absent.
	FetchTime(ctx context.Context, meta *TableMeta, recordID string) (int64, error)
	// ApplyRow inserts or updates a row with LWW: returns false (and
	// the current time) when the local row is newer than the change.
	ApplyRow(ctx context.Context, meta *TableMeta, payload map[string]any, updatedAt int64) (applied bool, localTime int64, err error)
	// ApplyTombstone deletes a row with LWW: returns false when the
	// local row is newer than the tombstone.
	ApplyTombstone(ctx context.Context, meta *TableMeta, recordID string, updatedAt int64) (applied bool, localTime int64, err error)
	// DeleteRow hard-deletes a row regardless of LWW (used to reconcile
	// after the server already resolved it).
	DeleteRow(ctx context.Context, meta *TableMeta, recordID string) error
	// SetTombstone upserts a tombstone marker.
	SetTombstone(ctx context.Context, t *Tombstone) error
	// PurgeTombstones removes markers of table whose updated_at is not
	// newer than keep (ms epoch); returns the number removed.
	PurgeTombstones(ctx context.Context, table string, keep int64) (int64, error)
	// FirstCompanyID returns the company that owns this install.
	FirstCompanyID(ctx context.Context) (string, error)
}
