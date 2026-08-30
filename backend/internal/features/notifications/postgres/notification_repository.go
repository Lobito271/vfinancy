package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/notifications"
)

var _ notifications.Repository = (*notificationRepository)(nil)

type notificationRepository struct {
	q persistence.Querier
}

func NewNotificationRepository(db *sql.DB) *notificationRepository {
	return &notificationRepository{q: persistence.FromDB(db)}
}

const notificationColumns = `
	id, company_id, type, title, message, record_type, record_id,
	dedup_key, read_at, created_at, deleted_at
`

// Insert ... ON CONFLICT is valid on both modernc/sqlite and
// postgres; collisions are silently skipped so re-scans never
// duplicate a notification for the same batch.
func (r *notificationRepository) CreateBatch(ctx context.Context, notes []*notifications.Notification) (int, error) {
	if len(notes) == 0 {
		return 0, nil
	}
	now := time.Now().UTC()
	const base = `INSERT INTO notifications (
		id, company_id, type, title, message, record_type, record_id,
		dedup_key, read_at, created_at, deleted_at
	) VALUES `
	ph := make([]string, 0, len(notes))
	args := make([]any, 0, len(notes)*10)
	for _, n := range notes {
		ph = append(ph, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(args)+1, len(args)+2, len(args)+3, len(args)+4, len(args)+5,
			len(args)+6, len(args)+7, len(args)+8, len(args)+9, len(args)+10, len(args)+11))
		id := n.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		args = append(args,
			id, n.CompanyID, n.Type, n.Title, n.Message,
			persistence.NullIfEmpty(n.RecordType),
			persistence.NullIfEmptyUUID(n.RecordID),
			n.DedupKey, persistence.NullIfZeroTime(n.ReadAt), now,
			persistence.NullIfZeroTime(n.DeletedAt),
		)
	}
	q := base + strings.Join(ph, ", ") +
		` ON CONFLICT (company_id, type, dedup_key) DO NOTHING`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, args...)
	if err != nil {
		return 0, persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *notificationRepository) List(ctx context.Context, filter notifications.ListFilter) (repositories.Page[*notifications.Notification], error) {
	var (
		clauses = []string{"company_id = $1", "deleted_at IS NULL"}
		args    = []any{filter.CompanyID}
	)
	if filter.UnreadOnly {
		clauses = append(clauses, "read_at IS NULL")
	}
	where := persistence.JoinClauses(clauses)
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM notifications WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*notifications.Notification]{}, persistence.Translate(err)
	}

	args = append(args, limit, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT "+notificationColumns+" FROM notifications WHERE %s ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d",
			where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*notifications.Notification]{}, persistence.Translate(err)
	}
	out := make([]*notifications.Notification, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		n, err := scanNotification(r)
		if err != nil {
			return err
		}
		out = append(out, n)
		return nil
	}); err != nil {
		return repositories.Page[*notifications.Notification]{}, err
	}
	return repositories.Page[*notifications.Notification]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *notificationRepository) CountUnread(ctx context.Context, companyID uuid.UUID) (int, error) {
	const q = `SELECT count(*) FROM notifications
		WHERE company_id = $1 AND read_at IS NULL AND deleted_at IS NULL`
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID).Scan(&n); err != nil {
		return 0, persistence.Translate(err)
	}
	return n, nil
}

func (r *notificationRepository) MarkRead(ctx context.Context, companyID uuid.UUID, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	phs := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids)+2)
	args = append(args, companyID)
	for _, id := range ids {
		args = append(args, id)
		phs = append(phs, fmt.Sprintf("$%d", len(args)))
	}
	q := fmt.Sprintf(`UPDATE notifications SET read_at = $%d
		WHERE company_id = $1 AND read_at IS NULL AND deleted_at IS NULL AND id IN (%s)`,
		len(args)+1, strings.Join(phs, ", "))
	args = append(args, time.Now().UTC())
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, args...)
	if err != nil {
		return 0, persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, companyID uuid.UUID) (int, error) {
	const q = `UPDATE notifications SET read_at = $1
		WHERE company_id = $2 AND read_at IS NULL AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, time.Now().UTC(), companyID)
	if err != nil {
		return 0, persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *notificationRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	const q = `UPDATE notifications SET deleted_at = $1
		WHERE company_id = $2 AND id = $3 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, time.Now().UTC(), companyID, id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

// DeleteStaleClearance removes the unread clearance alerts whose batch
// stopped being on clearance (stock ran out, written off, ...). Read
// alerts stay as history.
func (r *notificationRepository) DeleteStaleClearance(ctx context.Context, companyID uuid.UUID, keep []string) (int, error) {
	q := `UPDATE notifications SET deleted_at = $1
		WHERE company_id = $2 AND type = $3 AND record_id IS NOT NULL
		  AND read_at IS NULL AND deleted_at IS NULL`
	args := []any{time.Now().UTC(), companyID, notifications.TypeClearance}
	if len(keep) > 0 {
		phs := make([]string, 0, len(keep))
		for _, id := range keep {
			args = append(args, id)
			phs = append(phs, fmt.Sprintf("$%d", len(args)))
		}
		q += ` AND record_id NOT IN (` + strings.Join(phs, ", ") + `)`
	}
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, args...)
	if err != nil {
		return 0, persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func scanNotification(r *sql.Rows) (*notifications.Notification, error) {
	n := &notifications.Notification{}
	var (
		recordType        sql.NullString
		recordID          sql.NullString
		readAt, deletedAt sql.NullTime
	)
	if err := r.Scan(
		&n.ID, &n.CompanyID, &n.Type, &n.Title, &n.Message,
		&recordType, &recordID,
		&n.DedupKey, &readAt, &n.CreatedAt, &deletedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if recordType.Valid {
		n.RecordType = recordType.String
	}
	if recordID.Valid {
		id := persistence.ParseUUID(recordID.String)
		n.RecordID = &id
	}
	if readAt.Valid {
		t := readAt.Time
		n.ReadAt = &t
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		n.DeletedAt = &t
	}
	return n, nil
}
