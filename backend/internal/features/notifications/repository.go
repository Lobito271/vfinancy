package notifications

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// ListFilter constrains the notification listing.
type ListFilter struct {
	CompanyID  uuid.UUID
	UnreadOnly bool
	repositories.PageRequest
}

// Repository persists the device-local notification feed.
type Repository interface {
	// CreateBatch inserts the notifications that do not collide on
	// (company_id, type, dedup_key), skipping duplicate rows. It
	// returns the number of newly inserted notifications.
	CreateBatch(ctx context.Context, notes []*Notification) (int, error)

	List(ctx context.Context, filter ListFilter) (repositories.Page[*Notification], error)
	CountUnread(ctx context.Context, companyID uuid.UUID) (int, error)

	// MarkRead marks the given notifications read. Unknown or already
	// read rows are ignored. Returns the number updated.
	MarkRead(ctx context.Context, companyID uuid.UUID, ids []uuid.UUID) (int, error)
	// MarkAllRead marks every unread notification of the company read.
	MarkAllRead(ctx context.Context, companyID uuid.UUID) (int, error)

	// Delete soft-deletes a single notification.
	Delete(ctx context.Context, companyID, id uuid.UUID) error

	// DeleteStaleClearance soft-deletes the unread clearance
	// notifications whose referenced batch is no longer in `keep`
	// (i.e. the stock left clearance). Returns the number removed.
	DeleteStaleClearance(ctx context.Context, companyID uuid.UUID, keep []string) (int, error)
}
