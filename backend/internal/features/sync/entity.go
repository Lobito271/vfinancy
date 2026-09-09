package sync

import "time"

// Conflict is the audit record of a resolved last-writer-wins
// divergence between the local row and the mirror row. The losing side
// and both timestamps are kept.
type Conflict struct {
	ID              string
	TableName       string
	RecordID        string
	LocalUpdatedAt  *time.Time
	RemoteUpdatedAt *time.Time
	Resolution      string
	Message         string
	CreatedAt       time.Time
}

// Tombstone is a pending local hard-delete marker written by the
// sync delete triggers. It is applied to the mirror and then removed.
type Tombstone struct {
	Table     string
	ID        string
	UpdatedAt time.Time
}

// NewConflict builds a conflict audit record stamped with the current
// time.
func NewConflict(table, recordID string, local, remote *time.Time, resolution, message string) Conflict {
	return Conflict{
		ID:              newID(),
		TableName:       table,
		RecordID:        recordID,
		LocalUpdatedAt:  local,
		RemoteUpdatedAt: remote,
		Resolution:      resolution,
		Message:         message,
		CreatedAt:       time.Now().UTC(),
	}
}
