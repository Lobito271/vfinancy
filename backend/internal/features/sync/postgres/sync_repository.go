// Package postgres implements the sync engine stores: LocalStore over
// the local SQLite runtime database and RemoteStore over the cloud
// PostgreSQL mirror. One SQL engine serves both dialects: statements
// use $N placeholders and portable native values, and the dialect only
// changes how time.Time arguments are encoded (millisecond integers on
// SQLite, timestamps on PostgreSQL).
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	stdsync "sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/features/sync"
)

// store runs the shared replication SQL against one database handle.
type store struct {
	db     *sql.DB
	sqlite bool
}

// Local is the sync LocalStore bound to the local SQLite runtime
// database.
type Local struct {
	store
}

// NewLocal returns a LocalStore over the SQLite runtime database.
func NewLocal(db *sql.DB) *Local {
	return &Local{store{db: db, sqlite: true}}
}

// Remote is the sync RemoteStore over the cloud PostgreSQL mirror. The
// connection is opened lazily on first use and survives network
// failures: the database/sql pool reconnects transparently.
type Remote struct {
	mu  stdsync.Mutex
	dsn string
	db  *sql.DB
	log *logger.Logger
}

// NewRemote returns a RemoteStore for the mirror at dsn.
func NewRemote(dsn string, log *logger.Logger) *Remote {
	return &Remote{dsn: dsn, log: log}
}

var (
	_ sync.LocalStore  = (*Local)(nil)
	_ sync.RemoteStore = (*Remote)(nil)
)

func (r *Remote) connect(ctx context.Context) (*store, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db == nil {
		db, err := sql.Open("pgx", r.dsn)
		if err != nil {
			return nil, fmt.Errorf("sync: open mirror: %w", err)
		}
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(2)
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("sync: connect mirror: %w", err)
		}
		r.db = db
	}
	return &store{db: r.db}, nil
}

// EnsureSchema applies pending mirror migrations.
func (r *Remote) EnsureSchema(ctx context.Context, migrationsFS fs.FS) error {
	st, err := r.connect(ctx)
	if err != nil {
		return err
	}
	if err := migrations.NewRunnerFS(migrationsFS, st.db, r.log, "postgres").Up(ctx); err != nil {
		return fmt.Errorf("sync: mirror schema: %w", err)
	}
	return nil
}

func (r *Remote) Changes(ctx context.Context, meta sync.TableMeta, since time.Time) ([]map[string]any, error) {
	st, err := r.connect(ctx)
	if err != nil {
		return nil, err
	}
	return st.Changes(ctx, meta, since)
}

func (r *Remote) UpsertRows(ctx context.Context, meta sync.TableMeta, rows []map[string]any) error {
	st, err := r.connect(ctx)
	if err != nil {
		return err
	}
	return st.writeRows(ctx, meta, rows)
}

func (r *Remote) ApplyDeletes(ctx context.Context, meta sync.TableMeta, ids []string) error {
	st, err := r.connect(ctx)
	if err != nil {
		return err
	}
	return st.ApplyDeletes(ctx, meta, ids)
}

func (r *Remote) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db == nil {
		return nil
	}
	err := r.db.Close()
	r.db = nil
	return err
}

func (s *store) timeVal(t time.Time) any {
	if s.sqlite {
		return t.UnixMilli()
	}
	return t.UTC()
}

func (s *store) Cursor(ctx context.Context, key string) (time.Time, error) {
	var at time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT last_updated_at FROM sync_cursors WHERE table_name = $1`, key).Scan(&at)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, persistence.Translate(err)
	}
	return at, nil
}

func (s *store) SetCursor(ctx context.Context, key string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sync_cursors (table_name, last_updated_at)
		VALUES ($1, $2)
		ON CONFLICT (table_name) DO UPDATE SET last_updated_at = excluded.last_updated_at
		WHERE excluded.last_updated_at > sync_cursors.last_updated_at`,
		key, s.timeVal(at))
	return persistence.Translate(err)
}

// ponytail: no batch limit — desktop tables are small; add one when a
// table grows past memory.
func (s *store) Changes(ctx context.Context, meta sync.TableMeta, since time.Time) ([]map[string]any, error) {
	q := fmt.Sprintf("SELECT * FROM %s WHERE %s > $1 ORDER BY %s ASC", meta.Name, meta.TimeColumn, meta.TimeColumn)
	rows, err := s.db.QueryContext(ctx, q, s.timeVal(since))
	if err != nil {
		return nil, persistence.Translate(err)
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		cols, vals, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		row, err := normalizeRow(meta, cols, vals)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *store) Tombstones(ctx context.Context, table string) ([]sync.Tombstone, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT record_id, updated_at FROM sync_tombstones
		WHERE table_name = $1
		ORDER BY updated_at ASC`, table)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	defer rows.Close()

	out := make([]sync.Tombstone, 0)
	for rows.Next() {
		var id string
		var at time.Time
		if err := rows.Scan(&id, &at); err != nil {
			return nil, persistence.Translate(err)
		}
		out = append(out, sync.Tombstone{Table: table, ID: id, UpdatedAt: at})
	}
	return out, rows.Err()
}

func (s *store) ForgetTombstones(ctx context.Context, table string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return persistence.Translate(err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM sync_tombstones WHERE table_name = $1 AND record_id = $2`, table, id); err != nil {
			return persistence.Translate(err)
		}
	}
	return persistence.Translate(tx.Commit())
}

// UpsertRows applies pulled rows with last-writer-wins inside one
// transaction and records the divergences it resolved.
func (s *store) UpsertRows(ctx context.Context, meta sync.TableMeta, rows []map[string]any) ([]sync.Conflict, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	defer func() { _ = tx.Rollback() }()

	conflicts := make([]sync.Conflict, 0)
	for _, row := range rows {
		id, err := meta.PKOf(row)
		if err != nil {
			return nil, err
		}
		skip, err := s.tombstoned(ctx, tx, meta.Name, id)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}
		remoteAt := timeValue(row[meta.TimeColumn])
		localAt, err := s.rowTime(ctx, tx, meta, id)
		if err != nil {
			return nil, err
		}
		if !sync.ResolveLWW(localAt, remoteAt) {
			conflicts = append(conflicts, sync.NewConflict(meta.Name, id,
				&localAt, &remoteAt, sync.ResolutionLocalWon, "local row is newer"))
			continue
		}
		if err := s.upsertRow(ctx, tx, meta, row); err != nil {
			return nil, err
		}
		if !localAt.IsZero() && !localAt.Equal(remoteAt) {
			conflicts = append(conflicts, sync.NewConflict(meta.Name, id,
				&localAt, &remoteAt, sync.ResolutionRemoteWon, "remote row is newer"))
		}
	}
	for _, c := range conflicts {
		if err := insertConflict(ctx, tx, s, c); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, persistence.Translate(err)
	}
	return conflicts, nil
}

// writeRows upserts rows unconditionally inside one transaction (the
// mirror path; the mirror has a single writer).
func (s *store) writeRows(ctx context.Context, meta sync.TableMeta, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return persistence.Translate(err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, row := range rows {
		if err := s.upsertRow(ctx, tx, meta, row); err != nil {
			return err
		}
	}
	return persistence.Translate(tx.Commit())
}

func (s *store) ApplyDeletes(ctx context.Context, meta sync.TableMeta, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return persistence.Translate(err)
	}
	defer func() { _ = tx.Rollback() }()
	q := fmt.Sprintf("DELETE FROM %s WHERE %s", meta.Name, pkWhere(meta, 1))
	for _, id := range ids {
		pks := meta.PKArgs(id)
		args := make([]any, len(pks))
		for i, v := range pks {
			args[i] = v
		}
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return persistence.Translate(err)
		}
	}
	return persistence.Translate(tx.Commit())
}

func (s *store) tombstoned(ctx context.Context, q persistence.Querier, table, id string) (bool, error) {
	var one int
	err := q.QueryRowContext(ctx,
		`SELECT 1 FROM sync_tombstones WHERE table_name = $1 AND record_id = $2`, table, id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, persistence.Translate(err)
	}
	return true, nil
}

func (s *store) rowTime(ctx context.Context, q persistence.Querier, meta sync.TableMeta, id string) (time.Time, error) {
	qry := fmt.Sprintf("SELECT %s FROM %s WHERE %s", meta.TimeColumn, meta.Name, pkWhere(meta, 1))
	pks := meta.PKArgs(id)
	args := make([]any, len(pks))
	for i, v := range pks {
		args[i] = v
	}
	var v any
	err := q.QueryRowContext(ctx, qry, args...).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, persistence.Translate(err)
	}
	return timeValue(v), nil
}

func (s *store) upsertRow(ctx context.Context, q persistence.Querier, meta sync.TableMeta, row map[string]any) error {
	cols := make([]string, 0, len(row))
	for c := range row {
		cols = append(cols, c)
	}
	sort.Strings(cols)
	args := make([]any, len(cols))
	for i, c := range cols {
		args[i] = row[c]
	}
	if _, err := q.ExecContext(ctx, buildUpsert(meta, cols), args...); err != nil {
		return persistence.Translate(err)
	}
	return nil
}

func insertConflict(ctx context.Context, q persistence.Querier, s *store, c sync.Conflict) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO sync_conflicts
			(id, table_name, record_id, local_updated_at, remote_updated_at, resolution, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.TableName, c.RecordID, nullTime(c.LocalUpdatedAt), nullTime(c.RemoteUpdatedAt),
		c.Resolution, c.Message, s.timeVal(c.CreatedAt))
	return persistence.Translate(err)
}

// normalizeRow converts one scanned row into the portable value shape:
// time.Time timestamps, string uuids/dates/money, bool flags, integers
// and nil for NULL.
func normalizeRow(meta sync.TableMeta, cols []string, vals []any) (map[string]any, error) {
	row := make(map[string]any, len(cols))
	for i, c := range cols {
		v := vals[i]
		switch {
		case c == meta.TimeColumn:
			t := timeValue(v)
			if t.IsZero() {
				return nil, fmt.Errorf("sync: %s: empty %s", meta.Name, c)
			}
			row[c] = t
		case meta.IsDateCol(c):
			if t, ok := v.(time.Time); ok {
				row[c] = t.Format("2006-01-02")
			} else {
				row[c] = plainValue(v)
			}
		case meta.IsBoolCol(c):
			row[c] = boolValue(v)
		default:
			row[c] = plainValue(v)
		}
	}
	return row, nil
}

func timeValue(v any) time.Time {
	switch x := v.(type) {
	case time.Time:
		return x
	case int64:
		return time.UnixMilli(x)
	case int32:
		return time.UnixMilli(int64(x))
	case []byte:
		return timeValue(string(x))
	case string:
		if t, err := time.Parse(time.RFC3339Nano, x); err == nil {
			return t
		}
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return time.UnixMilli(n)
		}
	}
	return time.Time{}
}

func boolValue(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case bool:
		return x
	case int64:
		return x != 0
	case int32:
		return x != 0
	case string:
		return x == "t" || x == "true" || x == "1"
	default:
		return false
	}
}

func plainValue(v any) any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case string, bool, int64, int32, float64, time.Time:
		return x
	case []byte:
		return string(x)
	case [16]byte:
		return uuid.UUID(x).String()
	default:
		return fmt.Sprintf("%v", x)
	}
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func scanRow(rows *sql.Rows) ([]string, []any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, nil, err
	}
	return cols, vals, nil
}

func buildUpsert(meta sync.TableMeta, cols []string) string {
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	sets := make([]string, 0, len(cols))
	for _, c := range cols {
		if isPK(meta, c) {
			continue
		}
		sets = append(sets, c+" = excluded."+c)
	}
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		meta.Name,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(meta.PKs, ", "),
		strings.Join(sets, ", "),
	)
}

func pkWhere(meta sync.TableMeta, start int) string {
	parts := make([]string, len(meta.PKs))
	for i, c := range meta.PKs {
		parts[i] = fmt.Sprintf("%s = $%d", c, start+i)
	}
	return strings.Join(parts, " AND ")
}

func isPK(meta sync.TableMeta, col string) bool {
	for _, c := range meta.PKs {
		if c == col {
			return true
		}
	}
	return false
}
