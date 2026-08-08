package postgres_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/internal/shared/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/infrastructure/sqlite"
	"vfinancy/backend/internal/features/sync"
	syncpostgres "vfinancy/backend/internal/features/sync/postgres"
)

const migrationDir = "../../../../../backend/migrations/sqlite"

func newTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"), database.Options{
		MaxOpenConns: 1, MaxIdleConns: 1,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	log := logger.New("error", "text", "stdout")
	if err := migrations.NewRunner(migrationDir, db.DB, log, "sqlite").Up(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestReplicationRoundTrip(t *testing.T) {
	persistence.SetDialect(persistence.DialectSQLite)
	ctx := context.Background()
	local := newTestDB(t)
	remote := newTestDB(t)
	lr := syncpostgres.NewSyncRepository(local.DB)
	rr := syncpostgres.NewSyncRepository(remote.DB)

	companyID, err := lr.FirstCompanyID(ctx)
	if err != nil {
		t.Fatalf("first company: %v", err)
	}
	dev := &sync.Device{ID: "device-1", CompanyID: companyID, Name: "test", Platform: "test", Token: "tok-1", IsLocal: true, IsActive: true}
	if err := lr.RegisterDevice(ctx, dev); err != nil {
		t.Fatalf("register local device: %v", err)
	}
	if err := rr.RegisterDevice(ctx, &sync.Device{ID: "device-1", CompanyID: companyID, Name: "test", Platform: "test", Token: "tok-1", IsActive: true}); err != nil {
		t.Fatalf("register remote device: %v", err)
	}

	// Bump the local copy so it is strictly newer than the remote's
	// independently seeded copy (same data in both, but remote seeded
	// later).
	if _, err := local.ExecContext(ctx,
		`UPDATE companies SET legal_name = 'Local Corp', updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE id = $1`,
		companyID); err != nil {
		t.Fatalf("bump local company: %v", err)
	}

	meta := sync.LookupTable("companies")
	if meta == nil {
		t.Fatal("companies not in registry")
	}

	rows, wm, err := lr.RowsChangedSince(ctx, meta, 0)
	if err != nil {
		t.Fatalf("local rows since 0: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 company, got %d", len(rows))
	}
	if wm < rows[0].UpdatedAt {
		t.Fatalf("watermark %d behind row time %d", wm, rows[0].UpdatedAt)
	}

	for _, ch := range rows {
		applied, _, err := rr.ApplyRow(ctx, meta, ch.Payload, ch.UpdatedAt)
		if err != nil {
			t.Fatalf("remote apply: %v", err)
		}
		if !applied {
			t.Fatal("newer local company was not applied on the remote")
		}
	}

	var name string
	if err := remote.QueryRowContext(ctx, `SELECT legal_name FROM companies WHERE id = $1`, companyID).Scan(&name); err != nil {
		t.Fatalf("remote company: %v", err)
	}
	if name != "Local Corp" {
		t.Fatalf("remote company name = %q, want Local Corp", name)
	}

	got, _, err := rr.RowsChangedSince(ctx, meta, 0)
	if err != nil {
		t.Fatalf("remote rows since 0: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 company on remote, got %d", len(got))
	}

	stale := rows[0]
	stale.UpdatedAt = rows[0].UpdatedAt - 1
	applied, localTime, err := rr.ApplyRow(ctx, meta, stale.Payload, stale.UpdatedAt)
	if err != nil {
		t.Fatalf("stale apply: %v", err)
	}
	if applied {
		t.Fatal("stale row with older timestamp was applied")
	}
	if localTime == 0 {
		t.Fatal("LWW rejection should report the current local time")
	}
}

func TestTombstonePropagation(t *testing.T) {
	persistence.SetDialect(persistence.DialectSQLite)
	ctx := context.Background()
	local := newTestDB(t)
	remote := newTestDB(t)
	lr := syncpostgres.NewSyncRepository(local.DB)
	rr := syncpostgres.NewSyncRepository(remote.DB)

	// A brand-new row that is not referenced by any FK, present on both
	// sides (the earlier push already replicated it to the remote).
	insert := `INSERT INTO currencies (code, symbol, name, decimal_places, type, is_active) VALUES ('XXX', 'X', 'Test currency', 2, 'fiat', TRUE)`
	if _, err := local.ExecContext(ctx, insert); err != nil {
		t.Fatalf("insert local currency: %v", err)
	}
	if _, err := remote.ExecContext(ctx, insert); err != nil {
		t.Fatalf("insert remote currency: %v", err)
	}

	if _, err := local.ExecContext(ctx, `DELETE FROM currencies WHERE code = 'XXX'`); err != nil {
		t.Fatalf("delete currency: %v", err)
	}

	tombs, err := lr.TombstonesSince(ctx, "currencies", 0)
	if err != nil {
		t.Fatalf("local tombstones: %v", err)
	}
	if len(tombs) != 1 {
		t.Fatalf("expected 1 tombstone, got %d", len(tombs))
	}
	if tombs[0].RecordID != "XXX" {
		t.Fatalf("tombstone record id = %q, want XXX", tombs[0].RecordID)
	}

	meta := sync.LookupTable("currencies")
	if ft, _ := rr.FetchTime(ctx, meta, "XXX"); ft == 0 {
		t.Fatal("currency XXX missing on remote before delete")
	}
	applied, _, err := rr.ApplyTombstone(ctx, meta, tombs[0].RecordID, tombs[0].UpdatedAt)
	if err != nil {
		t.Fatalf("remote apply tombstone: %v", err)
	}
	if !applied {
		t.Fatal("remote tombstone not applied")
	}
	if err := rr.SetTombstone(ctx, &sync.Tombstone{TableName: "currencies", RecordID: tombs[0].RecordID, UpdatedAt: tombs[0].UpdatedAt}); err != nil {
		t.Fatalf("set tombstone: %v", err)
	}
	if ft, _ := rr.FetchTime(ctx, meta, "XXX"); ft != 0 {
		t.Fatal("currency XXX still present on remote after delete")
	}

	propagated, err := rr.TombstonesSince(ctx, "currencies", 0)
	if err != nil {
		t.Fatalf("propagated tombstones: %v", err)
	}
	if len(propagated) != 1 || propagated[0].RecordID != "XXX" {
		t.Fatalf("second device did not see the delete: %+v", propagated)
	}
}

func TestCompositePKTombstone(t *testing.T) {
	persistence.SetDialect(persistence.DialectSQLite)
	ctx := context.Background()
	db := newTestDB(t)
	repo := syncpostgres.NewSyncRepository(db.DB)

	var roleID, permissionCode string
	if err := db.QueryRowContext(ctx, `SELECT role_id, permission_code FROM role_permissions LIMIT 1`).Scan(&roleID, &permissionCode); err != nil {
		t.Fatalf("select role_permission: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1 AND permission_code = $2`, roleID, permissionCode); err != nil {
		t.Fatalf("delete role_permission: %v", err)
	}

	tombs, err := repo.TombstonesSince(ctx, "role_permissions", 0)
	if err != nil {
		t.Fatalf("tombstones: %v", err)
	}
	if len(tombs) != 1 {
		t.Fatalf("expected 1 tombstone, got %d", len(tombs))
	}

	meta := sync.LookupTable("role_permissions")
	expected, err := meta.RecordID(map[string]any{"role_id": roleID, "permission_code": permissionCode})
	if err != nil {
		t.Fatalf("record id: %v", err)
	}
	if tombs[0].RecordID != expected {
		t.Fatalf("tombstone record id = %q, want %q", tombs[0].RecordID, expected)
	}

	var vals []string
	if err := json.Unmarshal([]byte(tombs[0].RecordID), &vals); err != nil {
		t.Fatalf("composite id is not a JSON array: %v", err)
	}
	if len(vals) != 2 {
		t.Fatalf("composite id has %d values, want 2", len(vals))
	}
}

func TestCursorMonotonic(t *testing.T) {
	persistence.SetDialect(persistence.DialectSQLite)
	ctx := context.Background()
	db := newTestDB(t)
	repo := syncpostgres.NewSyncRepository(db.DB)

	companyID, err := repo.FirstCompanyID(ctx)
	if err != nil {
		t.Fatalf("first company: %v", err)
	}
	if err := repo.RegisterDevice(ctx, &sync.Device{ID: "device-1", CompanyID: companyID, Name: "t", Platform: "t", Token: "tok", IsLocal: true, IsActive: true}); err != nil {
		t.Fatalf("register: %v", err)
	}

	now := time.Now().UnixMilli()
	if err := repo.UpdateCursors(ctx, "device-1", map[string]int64{"companies": now}); err != nil {
		t.Fatalf("update cursor: %v", err)
	}
	// A stale watermark must not move the cursor backward.
	if err := repo.UpdateCursors(ctx, "device-1", map[string]int64{"companies": now - 500}); err != nil {
		t.Fatalf("update stale cursor: %v", err)
	}
	cursors, err := repo.GetCursors(ctx, "device-1")
	if err != nil {
		t.Fatalf("get cursors: %v", err)
	}
	if cursors["companies"] != now {
		t.Fatalf("cursor moved backward to %d, want %d", cursors["companies"], now)
	}
}
