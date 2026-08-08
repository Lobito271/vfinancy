// Package sqlite provides the SQLite connection used as the primary
// runtime database of the desktop app. SQLite is embedded (pure Go,
// no CGO) so the Wails app can be cross-compiled without a C toolchain.
//
// The connection is configured so timestamps are stored as INTEGER
// milliseconds since the Unix epoch (always comparable, timezone
// independent) and scanned back into time.Time via the driver's
// _inttotime / _texttotime options. Foreign keys are enforced and the
// WAL journal is used for better concurrency under the sync worker.
//
// Note: modernc.org/sqlite writes time.Time as INTEGER ms because of
// _time_integer_format=unix_milli, but its _inttotime read path uses the
// legacy mattn heuristic: INTEGER values with |v| < 1e12 are read as
// Unix seconds, everything else as ms. Real timestamps (epoch ms are
// ~1.7e12) therefore round-trip correctly; only synthetic sub-1e12
// values (e.g. a cursor of 1000) would be misread, so keep the code's
// time values on real clocks.
package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"vfinancy/backend/infrastructure/database"
)

const DriverName = "sqlite"

// DSN builds the driver-specific DSN for a SQLite file. The _txlock
// (immediate) option takes write locks up front so a background sync
// worker and foreground writes do not deadlock on "database is locked".
func DSN(path string) string {
	return fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)"+
			"&_pragma=busy_timeout(10000)"+
			"&_pragma=journal_mode(WAL)"+
			"&_pragma=synchronous(NORMAL)"+
			"&_time_integer_format=unix_milli"+
			"&_inttotime=1&_texttotime=1&_timezone=UTC&_txlock=immediate",
		path,
	)
}

// Open opens (creating if needed) the SQLite database at path. The
// parent directory is created when it does not exist. SQLite manages
// its own connection pool internally; keep the pool small to avoid
// WAL file contention.
func Open(path string, opts database.Options) (*database.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite: empty path")
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("sqlite: create dir %q: %w", dir, err)
		}
	}
	if opts.MaxOpenConns == 0 {
		opts.MaxOpenConns = 4
	}
	if opts.MaxIdleConns == 0 {
		opts.MaxIdleConns = 2
	}
	if opts.ConnMaxLifetime == 0 {
		opts.ConnMaxLifetime = 30 * time.Minute
	}
	db, err := database.Open(DriverName, DSN(path), opts)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: ping %q: %w", path, err)
	}
	return db, nil
}
