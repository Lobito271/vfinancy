package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"vfinancy/backend/internal/shared/logger"
)

type Migration struct {
	Version  int
	Name     string
	UpSQL    string
	DownSQL  string
	Filename string
}

type Runner struct {
	dir      string
	db       *sql.DB
	log      *logger.Logger
	table    string
	dialect  string
}

// NewRunner returns a runner that applies the migrations in dir to db.
// dialect is "postgres" or "sqlite" and selects the bookkeeping-table
// DDL; the migration files themselves must match the dialect.
func NewRunner(dir string, db *sql.DB, log *logger.Logger, dialect string) *Runner {
	if dialect == "" {
		dialect = "postgres"
	}
	return &Runner{dir: dir, db: db, log: log, table: "schema_migrations", dialect: dialect}
}

func (r *Runner) EnsureTable(ctx context.Context) error {
	createTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version    BIGINT PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, r.table)
	if r.dialect == "sqlite" {
		createTable = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version    INTEGER PRIMARY KEY,
				name       TEXT NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
			)`, r.table)
	}
	_, err := r.db.ExecContext(ctx, createTable)
	if err != nil {
		return fmt.Errorf("migrations: ensure table: %w", err)
	}
	return nil
}

func (r *Runner) Load() ([]Migration, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("migrations: read dir %q: %w", r.dir, err)
	}

	byVersion := map[int]Migration{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		version, mname, direction, err := parseFilename(name)
		if err != nil {
			r.log.Warn("skipping migration file", "file", name, "error", err.Error())
			continue
		}

		path := filepath.Join(r.dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("migrations: read %q: %w", path, err)
		}

		m, ok := byVersion[version]
		if !ok {
			m = Migration{Version: version, Name: mname, Filename: name}
			byVersion[version] = m
		} else {
			m.Filename = name
		}
		switch direction {
		case "up":
			m.UpSQL = string(data)
		case "down":
			m.DownSQL = string(data)
		}
		byVersion[version] = m
	}

	out := make([]Migration, 0, len(byVersion))
	for _, m := range byVersion {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

func parseFilename(name string) (version int, mname, direction string, err error) {
	var suffix string
	switch {
	case strings.HasSuffix(name, ".up.sql"):
		suffix = ".up"
		direction = "up"
	case strings.HasSuffix(name, ".down.sql"):
		suffix = ".down"
		direction = "down"
	default:
		return 0, "", "", fmt.Errorf("invalid migration name %q: must end in .up.sql or .down.sql", name)
	}
	stem := strings.TrimSuffix(name, suffix+".sql")
	idx := strings.IndexByte(stem, '_')
	if idx <= 0 || idx == len(stem)-1 {
		return 0, "", "", fmt.Errorf("invalid migration name %q: expected VERSION_NAME.up|down.sql", name)
	}
	versionStr := stem[:idx]
	rest := stem[idx+1:]
	v, err := strconv.Atoi(versionStr)
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid version in %q: %w", name, err)
	}
	if rest == "" {
		return 0, "", "", fmt.Errorf("invalid migration name %q: missing descriptive name", name)
	}
	return v, rest, direction, nil
}

func (r *Runner) applied(ctx context.Context) (map[int]bool, error) {
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf("SELECT version FROM %s", r.table))
	if err != nil {
		return nil, fmt.Errorf("migrations: list applied: %w", err)
	}
	defer rows.Close()
	out := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func (r *Runner) Up(ctx context.Context) error {
	if err := r.EnsureTable(ctx); err != nil {
		return err
	}
	migs, err := r.Load()
	if err != nil {
		return err
	}
	done, err := r.applied(ctx)
	if err != nil {
		return err
	}

	applied := 0
	for _, m := range migs {
		if done[m.Version] {
			continue
		}
		if m.UpSQL == "" {
			return fmt.Errorf("migrations: missing .up.sql for version %d (%s)", m.Version, m.Name)
		}
		r.log.Info("applying migration", "version", m.Version, "name", m.Name)
		if err := r.runInTx(ctx, m.UpSQL); err != nil {
			return fmt.Errorf("migrations: apply %d: %w", m.Version, err)
		}
		if _, err := r.db.ExecContext(ctx,
			fmt.Sprintf("INSERT INTO %s (version, name, applied_at) VALUES ($1, $2, $3)", r.table),
			m.Version, m.Name, time.Now().UTC(),
		); err != nil {
			return fmt.Errorf("migrations: record %d: %w", m.Version, err)
		}
		applied++
	}
	r.log.Info("migrations complete", "applied", applied, "total", len(migs))
	return nil
}

func (r *Runner) Down(ctx context.Context, steps int) error {
	if err := r.EnsureTable(ctx); err != nil {
		return err
	}
	migs, err := r.Load()
	if err != nil {
		return err
	}
	done, err := r.applied(ctx)
	if err != nil {
		return err
	}

	reversed := make([]Migration, 0, len(migs))
	for i := len(migs) - 1; i >= 0; i-- {
		if done[migs[i].Version] {
			reversed = append(reversed, migs[i])
		}
	}
	if steps <= 0 || steps > len(reversed) {
		steps = len(reversed)
	}

	for i := 0; i < steps; i++ {
		m := reversed[i]
		if m.DownSQL == "" {
			return fmt.Errorf("migrations: missing .down.sql for version %d (%s)", m.Version, m.Name)
		}
		r.log.Info("reverting migration", "version", m.Version, "name", m.Name)
		if err := r.runInTx(ctx, m.DownSQL); err != nil {
			return fmt.Errorf("migrations: revert %d: %w", m.Version, err)
		}
		if _, err := r.db.ExecContext(ctx,
			fmt.Sprintf("DELETE FROM %s WHERE version = $1", r.table),
			m.Version,
		); err != nil {
			return fmt.Errorf("migrations: unrecord %d: %w", m.Version, err)
		}
	}
	return nil
}

func (r *Runner) runInTx(ctx context.Context, sqlText string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, sqlText); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			return fmt.Errorf("rollback: %w (original: %v)", rbErr, err)
		}
		return err
	}
	return tx.Commit()
}

func (r *Runner) Status(ctx context.Context) error {
	if err := r.EnsureTable(ctx); err != nil {
		return err
	}
	migs, err := r.Load()
	if err != nil {
		return err
	}
	done, err := r.applied(ctx)
	if err != nil {
		return err
	}
	for _, m := range migs {
		state := "pending"
		if done[m.Version] {
			state = "applied"
		}
		r.log.Info("migration", "version", m.Version, "name", m.Name, "status", state)
	}
	return nil
}

func (r *Runner) FromFS(fsys fs.FS) ([]Migration, error) {
	byVersion := map[int]Migration{}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}
		name := filepath.Base(path)
		version, mname, direction, err := parseFilename(name)
		if err != nil {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		m, ok := byVersion[version]
		if !ok {
			m = Migration{Version: version, Name: mname, Filename: name}
		}
		switch direction {
		case "up":
			m.UpSQL = string(data)
		case "down":
			m.DownSQL = string(data)
		}
		byVersion[version] = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]Migration, 0, len(byVersion))
	for _, m := range byVersion {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}
