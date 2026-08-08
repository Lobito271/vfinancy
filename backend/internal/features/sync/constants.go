package sync

import "github.com/google/uuid"

// Conflict resolutions recorded in sync_conflicts. The writer always
// records from its own perspective: LOCAL_WON means the writer kept its
// own copy, REMOTE_WON means the incoming copy won.
const (
	ResolutionLocalWon  = "LOCAL_WON"
	ResolutionRemoteWon = "REMOTE_WON"
)

// Result statuses reported by the sync server per pushed item.
const (
	StatusApplied  = "applied"
	StatusConflict = "conflict"
	StatusFailed   = "failed"
)

// Wire protocol paths.
const (
	pathRegister = "/api/v1/devices/register"
	pathSync     = "/api/v1/sync"
)

func newID() string {
	return uuid.NewString()
}
