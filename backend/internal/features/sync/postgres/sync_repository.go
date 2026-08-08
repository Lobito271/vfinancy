package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/sync"
)

// syncRepository is the persistence implementation of the replication
// engine. Every query uses $N placeholders so the same code runs on the
// local SQLite database and the cloud PostgreSQL mirror; the only
// dialect divergence is the encoding of time arguments (INTEGER ms on
// SQLite, time.Time on PostgreSQL).
type syncRepository struct {
	q persistence.Querier
}

// NewSyncRepository returns a repository bound to a connection pool.
func NewSyncRepository(db *sql.DB) *syncRepository {
	return &syncRepository{q: persistence.FromDB(db)}
}

// --- devices ---

const deviceColumns = `
	id, company_id, name, platform, token, is_local, is_active,
	last_seen_at, created_at, updated_at
`

func (r *syncRepository) GetLocalDevice(ctx context.Context) (*sync.Device, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+deviceColumns+` FROM sync_devices WHERE is_local = TRUE LIMIT 1`)
	return scanDevice(row)
}

func (r *syncRepository) GetDeviceByToken(ctx context.Context, token string) (*sync.Device, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+deviceColumns+` FROM sync_devices WHERE token = $1 AND is_active = TRUE LIMIT 1`, token)
	return scanDevice(row)
}

func (r *syncRepository) RegisterDevice(ctx context.Context, d *sync.Device) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, `
		INSERT INTO sync_devices
			(id, company_id, name, platform, token, is_local, is_active, last_seen_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)`,
		d.ID, d.CompanyID, d.Name, d.Platform, d.Token, d.IsLocal, d.IsActive, nullTime(d.LastSeen), timeArg(d.CreatedAt.UnixMilli()))
	return persistence.Translate(err)
}

func (r *syncRepository) TouchDeviceSeen(ctx context.Context, deviceID string, at int64) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`UPDATE sync_devices SET last_seen_at = $1, updated_at = $1 WHERE id = $2`, timeArg(at), deviceID)
	return persistence.Translate(err)
}

// --- cursors ---

func (r *syncRepository) GetCursors(ctx context.Context, deviceID string) (map[string]int64, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT table_name, last_updated_at FROM sync_cursors WHERE device_id = $1`, deviceID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	defer rows.Close()
	out := make(map[string]int64)
	for rows.Next() {
		var table string
		var last time.Time
		if err := rows.Scan(&table, &last); err != nil {
			return nil, persistence.Translate(err)
		}
		out[table] = last.UnixMilli()
	}
	return out, rows.Err()
}

func (r *syncRepository) UpdateCursors(ctx context.Context, deviceID string, cursors map[string]int64) error {
	for table, wm := range cursors {
		_, err := persistence.Q(ctx, r.q).ExecContext(ctx, `
			INSERT INTO sync_cursors (device_id, table_name, last_updated_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (device_id, table_name) DO UPDATE SET last_updated_at = excluded.last_updated_at
			WHERE excluded.last_updated_at > sync_cursors.last_updated_at`,
			deviceID, table, timeArg(wm))
		if err != nil {
			return persistence.Translate(err)
		}
	}
	return nil
}

// --- conflicts ---

func (r *syncRepository) LogConflict(ctx context.Context, c *sync.Conflict) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, `
		INSERT INTO sync_conflicts
			(id, device_id, table_name, record_id, operation,
			 local_updated_at, remote_updated_at, resolution, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		c.ID, nullString(c.DeviceID), c.TableName, c.RecordID, c.Operation,
		nullTimePtr(c.LocalUpdatedAt), nullTimePtr(c.RemoteUpdatedAt), c.Resolution, c.Message, timeArg(c.CreatedAt.UnixMilli()))
	return persistence.Translate(err)
}

// --- row access ---

func (r *syncRepository) RowsChangedSince(ctx context.Context, meta *sync.TableMeta, after int64) ([]sync.Change, int64, error) {
	q := fmt.Sprintf(`SELECT * FROM %s WHERE %s > $1 ORDER BY %s ASC`, meta.Name, meta.TimeColumn, meta.TimeColumn)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, timeArg(after))
	if err != nil {
		return nil, after, persistence.Translate(err)
	}
	defer rows.Close()
	out := make([]sync.Change, 0)
	watermark := after
	for rows.Next() {
		cols, vals, err := scanAnyRow(rows)
		if err != nil {
			return nil, after, persistence.Translate(err)
		}
		payload, timeMs := rowToPayload(meta, cols, vals)
		if timeMs > watermark {
			watermark = timeMs
		}
		recordID, err := meta.RecordID(payload)
		if err != nil {
			return nil, after, persistence.Translate(err)
		}
		out = append(out, sync.Change{
			TableName: meta.Name,
			RecordID:  recordID,
			UpdatedAt: timeMs,
			Payload:   payload,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, after, persistence.Translate(err)
	}
	return out, watermark, nil
}

func (r *syncRepository) TombstonesSince(ctx context.Context, table string, after int64) ([]sync.Tombstone, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, `
		SELECT table_name, record_id, updated_at
		FROM sync_tombstones
		WHERE table_name = $1 AND updated_at > $2
		ORDER BY updated_at ASC`, table, timeArg(after))
	if err != nil {
		return nil, persistence.Translate(err)
	}
	defer rows.Close()
	out := make([]sync.Tombstone, 0)
	for rows.Next() {
		var t sync.Tombstone
		var last time.Time
		if err := rows.Scan(&t.TableName, &t.RecordID, &last); err != nil {
			return nil, persistence.Translate(err)
		}
		t.UpdatedAt = last.UnixMilli()
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *syncRepository) FetchTime(ctx context.Context, meta *sync.TableMeta, recordID string) (int64, error) {
	pkValues, err := meta.PKArgs(recordID)
	if err != nil {
		return 0, err
	}
	q := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, meta.TimeColumn, meta.Name, pkWhere(meta, 1))
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, toAny(pkValues)...)
	var v any
	if err := row.Scan(&v); err != nil {
		if persistence.IsPgNoRows(err) {
			return 0, nil
		}
		return 0, persistence.Translate(err)
	}
	return scanTime(v), nil
}

func (r *syncRepository) ApplyRow(ctx context.Context, meta *sync.TableMeta, payload map[string]any, updatedAt int64) (bool, int64, error) {
	recordID, err := meta.RecordID(payload)
	if err != nil {
		return false, 0, err
	}
	localTime, err := r.FetchTime(ctx, meta, recordID)
	if err != nil {
		return false, 0, err
	}
	if localTime > updatedAt {
		return false, localTime, nil
	}
	cols, args := payloadToArgs(meta, payload)
	q := buildUpsert(meta, cols)
	if _, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, args...); err != nil {
		return false, localTime, persistence.Translate(err)
	}
	return true, localTime, nil
}

func (r *syncRepository) ApplyTombstone(ctx context.Context, meta *sync.TableMeta, recordID string, updatedAt int64) (bool, int64, error) {
	localTime, err := r.FetchTime(ctx, meta, recordID)
	if err != nil {
		return false, 0, err
	}
	if localTime == 0 {
		return true, 0, nil
	}
	if localTime > updatedAt {
		return false, localTime, nil
	}
	pkValues, err := meta.PKArgs(recordID)
	if err != nil {
		return false, localTime, err
	}
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s`, meta.Name, pkWhere(meta, 1))
	if _, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, toAny(pkValues)...); err != nil {
		return false, localTime, persistence.Translate(err)
	}
	return true, localTime, nil
}

func (r *syncRepository) DeleteRow(ctx context.Context, meta *sync.TableMeta, recordID string) error {
	pkValues, err := meta.PKArgs(recordID)
	if err != nil {
		return err
	}
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s`, meta.Name, pkWhere(meta, 1))
	if _, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, toAny(pkValues)...); err != nil {
		return persistence.Translate(err)
	}
	return nil
}

func (r *syncRepository) SetTombstone(ctx context.Context, t *sync.Tombstone) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, `
		INSERT INTO sync_tombstones (table_name, record_id, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (table_name, record_id) DO UPDATE SET updated_at = excluded.updated_at`,
		t.TableName, t.RecordID, timeArg(t.UpdatedAt))
	return persistence.Translate(err)
}

func (r *syncRepository) PurgeTombstones(ctx context.Context, table string, keep int64) (int64, error) {
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`DELETE FROM sync_tombstones WHERE table_name = $1 AND updated_at <= $2`, table, timeArg(keep))
	if err != nil {
		return 0, persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *syncRepository) FirstCompanyID(ctx context.Context) (string, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, `
		SELECT id FROM companies WHERE is_active = TRUE ORDER BY created_at LIMIT 1`)
	var id string
	if err := row.Scan(&id); err != nil {
		if persistence.IsPgNoRows(err) {
			return "", repositories.ErrNotFound
		}
		return "", persistence.Translate(err)
	}
	return id, nil
}

var _ sync.Repository = (*syncRepository)(nil)

// --- generic SQL helpers (dialect-safe) ---

// timeArg encodes a ms epoch timestamp for the active dialect.
func timeArg(ms int64) any {
	if persistence.IsSQLite() {
		return ms
	}
	return time.UnixMilli(ms).UTC()
}

// scanTime coerces an arbitrary scanned value to ms since epoch.
func scanTime(v any) int64 {
	switch x := v.(type) {
	case time.Time:
		return x.UnixMilli()
	case int64:
		return x
	case int32:
		return int64(x)
	case float64:
		return int64(x)
	case json.Number:
		if n, err := x.Int64(); err == nil {
			return n
		}
	case []byte:
		return scanTime(string(x))
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return n
		}
		if t, err := time.Parse(time.RFC3339Nano, x); err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

// rowToPayload turns one scanned row into a wire payload. Time columns
// become RFC3339 strings, decimal columns stay exact strings, and every
// other column is JSON-safe. The TimeColumn value is returned as ms.
func rowToPayload(meta *sync.TableMeta, cols []string, vals []any) (map[string]any, int64) {
	p := make(map[string]any, len(cols))
	var timeMs int64
	for i, c := range cols {
		v := vals[i]
		switch {
		case meta.IsTimeCol(c):
			ms := scanTime(v)
			if c == meta.TimeColumn {
				timeMs = ms
			}
			p[c] = time.UnixMilli(ms).UTC().Format(time.RFC3339Nano)
		case meta.IsDecimalCol(c):
			p[c] = decimalString(v)
		default:
			p[c] = jsonSafe(v)
		}
	}
	return p, timeMs
}

// payloadToArgs binds a payload to $N placeholders, converting time
// columns and decimal columns back to engine-native values.
func payloadToArgs(meta *sync.TableMeta, payload map[string]any) ([]string, []any) {
	cols := make([]string, 0, len(payload))
	for c := range payload {
		cols = append(cols, c)
	}
	sort.Strings(cols)
	args := make([]any, len(cols))
	for i, c := range cols {
		v := payload[c]
		if v == nil {
			args[i] = nil
			continue
		}
		switch {
		case meta.IsTimeCol(c):
			if s, ok := v.(string); ok {
				if ms, err := parseRFC3339Ms(s); err == nil {
					args[i] = timeArg(ms)
				} else {
					args[i] = s
				}
			} else {
				args[i] = timeArg(scanTime(v))
			}
		case meta.IsDecimalCol(c):
			args[i] = decimalString(v)
		default:
			args[i] = v
		}
	}
	return cols, args
}

// buildUpsert renders an idempotent INSERT ... ON CONFLICT upsert for a
// column list; the conflict target is the table primary key.
func buildUpsert(meta *sync.TableMeta, cols []string) string {
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(meta.Name)
	b.WriteString(" (")
	b.WriteString(strings.Join(cols, ", "))
	b.WriteString(") VALUES (")
	b.WriteString(strings.Join(placeholders, ", "))
	b.WriteString(") ON CONFLICT (")
	b.WriteString(strings.Join(meta.PKColumns, ", "))
	b.WriteString(") DO ")
	sets := make([]string, 0, len(cols))
	for _, c := range cols {
		if isPK(meta, c) {
			continue
		}
		sets = append(sets, c+" = excluded."+c)
	}
	if len(sets) == 0 {
		return b.String() + "NOTHING"
	}
	b.WriteString("UPDATE SET ")
	b.WriteString(strings.Join(sets, ", "))
	return b.String()
}

// pkWhere renders "pk1 = $N AND pk2 = $N+1" with placeholders starting
// at start (1-based).
func pkWhere(meta *sync.TableMeta, start int) string {
	parts := make([]string, len(meta.PKColumns))
	for i, c := range meta.PKColumns {
		parts[i] = fmt.Sprintf("%s = $%d", c, start+i)
	}
	return strings.Join(parts, " AND ")
}

func isPK(meta *sync.TableMeta, col string) bool {
	for _, c := range meta.PKColumns {
		if c == col {
			return true
		}
	}
	return false
}

func toAny(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

// --- scan helpers ---

func scanAnyRow(rows *sql.Rows) ([]string, []any, error) {
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

func scanDevice(row *sql.Row) (*sync.Device, error) {
	d := &sync.Device{}
	var last, created, updated time.Time
	err := row.Scan(&d.ID, &d.CompanyID, &d.Name, &d.Platform, &d.Token, &d.IsLocal, &d.IsActive, &last, &created, &updated)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, persistence.Translate(err)
	}
	if !last.IsZero() {
		d.LastSeen = &last
	}
	d.CreatedAt = created
	d.UpdatedAt = updated
	return d, nil
}

// --- value coercion helpers ---

func jsonSafe(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string, bool, int64, int, float64:
		return x
	case []byte:
		return string(x)
	case time.Time:
		return x.UTC().Format(time.RFC3339Nano)
	case json.Number:
		return x.String()
	default:
		return fmt.Sprintf("%v", x)
	}
}

func decimalString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case json.Number:
		return x.String()
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return fmt.Sprintf("%v", x)
	}
}

func parseRFC3339Ms(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

func nullString(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullTime(p *time.Time) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullTimePtr(p *time.Time) any {
	if p == nil {
		return nil
	}
	return *p
}
