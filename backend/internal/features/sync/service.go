// Package sync replicates business rows between the local SQLite
// runtime database (the single source of truth) and a cloud PostgreSQL
// mirror over a direct pgx connection. A background worker calls
// RunOnce on a ticker: per table it pulls mirror rows and applies them
// with last-writer-wins, pushes local changes, then pushes local
// hard-deletes. Every error is returned to the worker, which logs it
// and retries on the next tick, so a sync failure never takes the app
// down.
package sync

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"time"
)

const (
	cursorPush = ":push"
	cursorPull = ":pull"
)

// Config wires the sync service to the cloud mirror.
type Config struct {
	Enabled      bool
	DSN          string
	PollInterval time.Duration
	MigrationsFS fs.FS
}

// Service performs bidirectional watermark replication between the
// local store and the mirror.
type Service struct {
	local       LocalStore
	remote      RemoteStore
	cfg         Config
	log         *slog.Logger
	schemaReady bool
}

// NewService returns a replication service bound to both stores.
func NewService(local LocalStore, remote RemoteStore, cfg Config, log *slog.Logger) *Service {
	return &Service{local: local, remote: remote, cfg: cfg, log: log}
}

// RunOnce performs one full replication exchange. The first error
// aborts the pass; the caller logs it and retries on the next tick.
func (s *Service) RunOnce(ctx context.Context) error {
	if err := s.ensureSchema(ctx); err != nil {
		return err
	}
	for _, meta := range SyncedTables() {
		if err := s.pullTable(ctx, meta); err != nil {
			return fmt.Errorf("sync: pull %s: %w", meta.Name, err)
		}
		if err := s.pushTable(ctx, meta); err != nil {
			return fmt.Errorf("sync: push %s: %w", meta.Name, err)
		}
		if err := s.pushDeletes(ctx, meta); err != nil {
			return fmt.Errorf("sync: push deletes %s: %w", meta.Name, err)
		}
	}
	return nil
}

// TestConnection verifies the mirror is reachable and its schema is
// current. The settings UI calls it before saving the configuration.
func (s *Service) TestConnection(ctx context.Context) error {
	if err := s.remote.EnsureSchema(ctx, s.cfg.MigrationsFS); err != nil {
		return fmt.Errorf("sync: test connection: %w", err)
	}
	s.schemaReady = true
	return nil
}

// Close releases the mirror connection pool.
func (s *Service) Close() {
	_ = s.remote.Close()
}

// ResolveLWW reports whether an incoming copy of a row must overwrite
// the local copy: absent or equally old local rows accept the incoming
// value, strictly newer local rows win.
func ResolveLWW(local, remote time.Time) bool {
	return !local.After(remote)
}

func (s *Service) ensureSchema(ctx context.Context) error {
	if s.schemaReady {
		return nil
	}
	if err := s.remote.EnsureSchema(ctx, s.cfg.MigrationsFS); err != nil {
		return fmt.Errorf("sync: ensure mirror schema: %w", err)
	}
	s.schemaReady = true
	return nil
}

func (s *Service) pullTable(ctx context.Context, meta TableMeta) error {
	since, err := s.local.Cursor(ctx, meta.Name+cursorPull)
	if err != nil {
		return err
	}
	rows, err := s.remote.Changes(ctx, meta, since)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	conflicts, err := s.local.UpsertRows(ctx, meta, rows)
	if err != nil {
		return err
	}
	for _, c := range conflicts {
		s.log.Warn("sync: conflict",
			"table", c.TableName,
			"record", c.RecordID,
			"resolution", c.Resolution,
			"message", c.Message,
		)
	}
	return s.local.SetCursor(ctx, meta.Name+cursorPull, rowWatermark(meta, rows))
}

func (s *Service) pushTable(ctx context.Context, meta TableMeta) error {
	since, err := s.local.Cursor(ctx, meta.Name+cursorPush)
	if err != nil {
		return err
	}
	rows, err := s.local.Changes(ctx, meta, since)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	if err := s.remote.UpsertRows(ctx, meta, rows); err != nil {
		return err
	}
	s.log.Debug("sync: rows pushed", "table", meta.Name, "count", len(rows))
	return s.local.SetCursor(ctx, meta.Name+cursorPush, rowWatermark(meta, rows))
}

func (s *Service) pushDeletes(ctx context.Context, meta TableMeta) error {
	tombs, err := s.local.Tombstones(ctx, meta.Name)
	if err != nil {
		return err
	}
	if len(tombs) == 0 {
		return nil
	}
	ids := make([]string, len(tombs))
	for i, tb := range tombs {
		ids[i] = tb.ID
	}
	if err := s.remote.ApplyDeletes(ctx, meta, ids); err != nil {
		return err
	}
	if err := s.local.ForgetTombstones(ctx, meta.Name, ids); err != nil {
		return err
	}
	s.log.Debug("sync: deletes pushed", "table", meta.Name, "count", len(ids))
	return nil
}

// rowWatermark returns the time column of the last row; rows arrive
// ordered ascending by that column.
func rowWatermark(meta TableMeta, rows []map[string]any) time.Time {
	last := rows[len(rows)-1]
	if t, ok := last[meta.TimeColumn].(time.Time); ok {
		return t
	}
	return time.Time{}
}
