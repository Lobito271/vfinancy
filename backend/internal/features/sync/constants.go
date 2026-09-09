package sync

import "github.com/google/uuid"

// Conflict resolutions recorded in sync_conflicts, from the applying
// (local) side's perspective: LOCAL_WON kept the local copy,
// REMOTE_WON applied the incoming copy.
const (
	ResolutionLocalWon  = "LOCAL_WON"
	ResolutionRemoteWon = "REMOTE_WON"
)

func newID() string {
	return uuid.NewString()
}
