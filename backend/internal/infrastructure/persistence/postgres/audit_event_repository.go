package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/administration"
	"vfinancy/backend/internal/domain/repositories"
)

type auditEventRepository struct {
	q Querier
}

func newAuditEventRepository(db *sql.DB) *auditEventRepository {
	return &auditEventRepository{q: &dbBox{db: db}}
}

func newAuditEventRepositoryTx(tx *sql.Tx) *auditEventRepository {
	return &auditEventRepository{q: &txBox{tx: tx}}
}

const auditEventColumns = `
	id, company_id, user_id, session_id, event_type, target_type, target_id,
	description, metadata, ip_address, device, occurred_at
`

func (r *auditEventRepository) Create(ctx context.Context, e *administration.AuditEvent) error {
	const q = `INSERT INTO audit_events (
		id, company_id, user_id, session_id, event_type, target_type, target_id,
		description, metadata, ip_address, device, occurred_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	var meta []byte
	if e.Metadata != nil {
		meta = []byte(e.Metadata)
	}
	_, err := r.q.ExecContext(ctx, q,
		e.ID, e.CompanyID, nullIfEmptyUUID(e.UserID), nullIfEmptyUUID(e.SessionID),
		string(e.EventType), nullIfEmpty(e.TargetType), nullIfEmptyUUID(e.TargetID),
		e.Description, meta, e.IPAddress, e.Device, e.OccurredAt,
	)
	return Translate(err)
}

func (r *auditEventRepository) List(ctx context.Context, companyID uuid.UUID, filter repositories.AuditEventFilter) ([]*administration.AuditEvent, int, error) {
	var (
		clauses = []string{"company_id = $1"}
		args    []any
	)
	args = append(args, companyID)

	if filter.EventType != "" {
		clauses = append(clauses, fmt.Sprintf("event_type = $%d", len(args)+1))
		args = append(args, filter.EventType)
	}
	if filter.UserID != nil {
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", len(args)+1))
		args = append(args, *filter.UserID)
	}
	if filter.From != nil {
		clauses = append(clauses, fmt.Sprintf("occurred_at >= $%d", len(args)+1))
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		clauses = append(clauses, fmt.Sprintf("occurred_at <= $%d", len(args)+1))
		args = append(args, *filter.To)
	}

	where := joinClauses(clauses)

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.q.QueryRowContext(ctx, "SELECT count(*) FROM audit_events WHERE "+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, Translate(err)
	}

	limit, offset := limitOffset(filter.PageRequest, 25, 200)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)

	q := fmt.Sprintf("SELECT %s FROM audit_events WHERE %s ORDER BY occurred_at DESC LIMIT $%d OFFSET $%d",
		auditEventColumns, where, limitPos, offsetPos)
	rows, err := r.q.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, Translate(err)
	}
	out := make([]*administration.AuditEvent, 0, limit)
	if err := scanRows(rows, func(r *sql.Rows) error {
		e, err := scanAuditEventFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, e)
		return nil
	}); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *auditEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*administration.AuditEvent, error) {
	q := `SELECT ` + auditEventColumns + ` FROM audit_events WHERE id = $1`
	row := r.q.QueryRowContext(ctx, q, id)
	return scanAuditEvent(row)
}

func scanAuditEvent(row *sql.Row) (*administration.AuditEvent, error) {
	e := &administration.AuditEvent{}
	var (
		userID, sessionID, targetID sql.NullString
		eventType, targetType       string
		description, ipAddress, device string
		metadata                    []byte
	)
	err := row.Scan(
		&e.ID, &e.CompanyID, &userID, &sessionID, &eventType, &targetType, &targetID,
		&description, &metadata, &ipAddress, &device, &e.OccurredAt,
	)
	if err != nil {
		if isPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, Translate(err)
	}
	e.EventType = administration.AuditEventType(eventType)
	e.TargetType = targetType
	e.Description = description
	e.IPAddress = ipAddress
	e.Device = device
	if metadata != nil {
		e.Metadata = json.RawMessage(metadata)
	}
	if userID.Valid {
		id := masterdataParseUUID(userID.String)
		e.UserID = &id
	}
	if sessionID.Valid {
		id := masterdataParseUUID(sessionID.String)
		e.SessionID = &id
	}
	if targetID.Valid {
		id := masterdataParseUUID(targetID.String)
		e.TargetID = &id
	}
	return e, nil
}

func scanAuditEventFromRows(rows *sql.Rows) (*administration.AuditEvent, error) {
	e := &administration.AuditEvent{}
	var (
		userID, sessionID, targetID sql.NullString
		eventType, targetType       string
		description, ipAddress, device string
		metadata                    []byte
	)
	if err := rows.Scan(
		&e.ID, &e.CompanyID, &userID, &sessionID, &eventType, &targetType, &targetID,
		&description, &metadata, &ipAddress, &device, &e.OccurredAt,
	); err != nil {
		return nil, Translate(err)
	}
	e.EventType = administration.AuditEventType(eventType)
	e.TargetType = targetType
	e.Description = description
	e.IPAddress = ipAddress
	e.Device = device
	if metadata != nil {
		e.Metadata = json.RawMessage(metadata)
	}
	if userID.Valid {
		id := masterdataParseUUID(userID.String)
		e.UserID = &id
	}
	if sessionID.Valid {
		id := masterdataParseUUID(sessionID.String)
		e.SessionID = &id
	}
	if targetID.Valid {
		id := masterdataParseUUID(targetID.String)
		e.TargetID = &id
	}
	return e, nil
}

var _ repositories.AuditEventRepository = (*auditEventRepository)(nil)
