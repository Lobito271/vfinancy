// Package notifications implements the device-local notification feed.
// A background worker scans the inventory for batches past their
// maximum sale date ("producto en remate") and records an alert per
// batch. The frontend bell dropdown consumes the feed; notifications
// stay on the device and are never replicated like master data.
package notifications

import (
	"time"

	"github.com/google/uuid"
)

// TypeClearance marks a notification about stock that passed its
// maximum sale date and is still on hand.
const TypeClearance = "inventory.clearance"

// RecordTypeBatch identifies the referenced record as an inventory
// batch. Notification.RecordID then carries the batch id.
const RecordTypeBatch = "inventory_batch"

// Notification is a business alert for a single company. An
// unread notification has a nil ReadAt.
type Notification struct {
	ID         uuid.UUID
	CompanyID  uuid.UUID
	Type       string
	Title      string
	Message    string
	RecordType string
	RecordID   *uuid.UUID
	DedupKey   string
	ReadAt     *time.Time
	CreatedAt  time.Time
	DeletedAt  *time.Time
}
