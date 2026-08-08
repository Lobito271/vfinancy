package sync

import (
	"encoding/json"
	"fmt"
)

// TableMeta describes a replicated table for the generic replication
// engine: how to identify a row (PKColumns), which column carries the
// LWW timestamp (TimeColumn), which columns hold timestamps that must
// round-trip between the two engines as RFC3339, and which columns are
// exact-precision decimals that travel as decimal strings.
type TableMeta struct {
	Name        string
	PKColumns   []string
	TimeColumn  string
	TimeColumns []string
	Decimals    []string
}

// SyncedTables is the authoritative list of replicated tables. audit_logs,
// audit_events and login_history are intentionally excluded: they are
// append-only forensic records of the local device and would flood the
// watermark scan. exchange_rates is excluded until its repository exists.
var SyncedTables = []*TableMeta{
	{Name: "companies", PKColumns: []string{"id"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at", "deleted_at"}},
	{Name: "branches", PKColumns: []string{"id"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at", "deleted_at"}},
	{Name: "roles", PKColumns: []string{"id"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at", "deleted_at"}},
	{Name: "users", PKColumns: []string{"id"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at", "deleted_at", "locked_until", "last_login_at"}},
	{Name: "user_roles", PKColumns: []string{"id"}, TimeColumn: "assigned_at", TimeColumns: []string{"assigned_at", "expires_at"}},
	{Name: "user_profiles", PKColumns: []string{"id"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at"}},
	{Name: "user_sessions", PKColumns: []string{"id"}, TimeColumn: "last_activity_at", TimeColumns: []string{"created_at", "last_activity_at", "locked_at", "expires_at"}},
	{Name: "application_settings", PKColumns: []string{"id"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at"}},
	{Name: "taxes", PKColumns: []string{"id"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at", "deleted_at"}, Decimals: []string{"default_rate"}},
	{Name: "currencies", PKColumns: []string{"code"}, TimeColumn: "updated_at", TimeColumns: []string{"created_at", "updated_at"}},
	{Name: "countries", PKColumns: []string{"code"}, TimeColumn: "created_at", TimeColumns: []string{"created_at"}},
	{Name: "permissions", PKColumns: []string{"code"}, TimeColumn: "created_at", TimeColumns: []string{"created_at"}},
	{Name: "role_permissions", PKColumns: []string{"role_id", "permission_code"}, TimeColumn: "created_at", TimeColumns: []string{"created_at"}},
}

// LookupTable returns the metadata for a replicated table name.
func LookupTable(name string) *TableMeta {
	for _, m := range SyncedTables {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// IsTimeCol reports whether col is one of the table's timestamp columns.
func (m *TableMeta) IsTimeCol(col string) bool {
	for _, c := range m.TimeColumns {
		if c == col {
			return true
		}
	}
	return false
}

// IsDecimalCol reports whether col holds an exact-precision decimal.
func (m *TableMeta) IsDecimalCol(col string) bool {
	for _, c := range m.Decimals {
		if c == col {
			return true
		}
	}
	return false
}

// SinglePK reports whether the table is keyed by a single column.
func (m *TableMeta) SinglePK() bool {
	return len(m.PKColumns) == 1
}

// RecordID renders the primary-key values of a payload as the opaque
// record id used on the wire and in sync_tombstones: the bare value for
// single-column keys, a JSON array for composite keys.
func (m *TableMeta) RecordID(payload map[string]any) (string, error) {
	if m.SinglePK() {
		v, ok := payload[m.PKColumns[0]]
		if !ok || v == nil {
			return "", errf("record has no pk column %q", m.PKColumns[0])
		}
		return scalarString(v), nil
	}
	vals := make([]string, 0, len(m.PKColumns))
	for _, c := range m.PKColumns {
		v, ok := payload[c]
		if !ok || v == nil {
			return "", errf("record has no pk column %q", c)
		}
		vals = append(vals, scalarString(v))
	}
	b, err := json.Marshal(vals)
	if err != nil {
		return "", errf("marshal record id: %w", err)
	}
	return string(b), nil
}

// PKArgs turns an opaque record id into the primary-key values, in
// PKColumns order, to bind in a WHERE clause.
func (m *TableMeta) PKArgs(recordID string) ([]string, error) {
	if m.SinglePK() {
		return []string{recordID}, nil
	}
	var vals []string
	if err := json.Unmarshal([]byte(recordID), &vals); err != nil {
		return nil, errf("invalid composite record id %q: %w", recordID, err)
	}
	if len(vals) != len(m.PKColumns) {
		return nil, errf("composite record id %q has %d values, expected %d", recordID, len(vals), len(m.PKColumns))
	}
	return vals, nil
}

func scalarString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return ""
	}
}

func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
