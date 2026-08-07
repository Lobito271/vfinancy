package sync

import "time"

// Device identifies a registered replicating installation. Exactly one
// row has IsLocal = true: this desktop install.
type Device struct {
	ID        string
	CompanyID string
	Name      string
	Platform  string
	Token     string
	IsLocal   bool
	IsActive  bool
	LastSeen  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SyncCursor is the per (device, table) watermark of the last change
// processed, in milliseconds since the Unix epoch.
type SyncCursor struct {
	DeviceID      string
	TableName     string
	LastUpdatedAt time.Time
}

// Conflict is a short alias for SyncConflict used by the service and
// server code that builds conflict records.
type Conflict = SyncConflict

// SyncConflict is the audit record of a resolved LWW conflict. The
// losing side and both timestamps are kept.
type SyncConflict struct {
	ID              string
	DeviceID        *string
	TableName       string
	RecordID        string
	Operation       string
	LocalUpdatedAt  *time.Time
	RemoteUpdatedAt *time.Time
	Resolution      string
	Message         string
	CreatedAt       time.Time
}

// Change is one replicated row. Payload holds every column of the row
// keyed by column name; time columns are RFC3339 strings, decimals are
// exact decimal strings, everything else is a JSON-native value.
type Change struct {
	TableName string         `json:"table_name"`
	RecordID  string         `json:"record_id"`
	UpdatedAt int64          `json:"updated_at_ms"`
	Payload   map[string]any `json:"payload"`
}

// Tombstone is a hard-delete marker.
type Tombstone struct {
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	UpdatedAt int64  `json:"updated_at_ms"`
}

// Result reports how the server applied a pushed row or tombstone.
type Result struct {
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	Status    string `json:"status"` // applied | conflict
	Message   string `json:"message,omitempty"`
}

// Request is the payload of a single bidirectional sync call.
type Request struct {
	Cursors    map[string]int64 `json:"cursors"`
	Rows       []Change         `json:"rows"`
	Tombstones []Tombstone      `json:"tombstones"`
}

// Response is the payload returned by the sync server. Cursors is the
// watermark to persist after applying Rows and Tombstones.
type Response struct {
	Results    []Result         `json:"results"`
	Rows       []Change         `json:"rows"`
	Tombstones []Tombstone      `json:"tombstones"`
	Cursors    map[string]int64 `json:"cursors"`
	ServerTime int64            `json:"server_time_ms"`
}

// RegisterRequest registers a device with the server.
type RegisterRequest struct {
	CompanyID string `json:"company_id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
}

// RegisterResponse carries the server-assigned device credentials.
type RegisterResponse struct {
	DeviceID string `json:"device_id"`
	Token    string `json:"token"`
}
