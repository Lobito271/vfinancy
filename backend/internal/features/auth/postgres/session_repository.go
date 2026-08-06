package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/auth"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
)

type sessionRepository struct {
	q persistence.Querier
}

func NewSessionRepository(db *sql.DB) *sessionRepository {
	return &sessionRepository{q: persistence.FromDB(db)}
}


const sessionColumns = `
	id, user_id, token, ip_address, user_agent, device, is_active,
	locked_at, locked_reason, expires_at, last_activity_at, created_at
`

func (r *sessionRepository) Create(ctx context.Context, s *auth.UserSession) error {
	const q = `INSERT INTO user_sessions (
		id, user_id, token, ip_address, user_agent, device, is_active,
		locked_at, locked_reason, expires_at, last_activity_at, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.ID, s.UserID, s.Token, s.IPAddress, s.UserAgent, s.Device,
		s.IsActive, persistence.NullIfZeroTime(s.LockedAt), persistence.NullIfEmpty(s.LockedReason),
		s.ExpiresAt, s.LastActivityAt, s.CreatedAt,
	)
	return persistence.Translate(err)
}

func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*auth.UserSession, error) {
	q := `SELECT ` + sessionColumns + ` FROM user_sessions WHERE token = $1`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, token)
	return scanSession(row)
}

func (r *sessionRepository) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*auth.UserSession, error) {
	q := `SELECT ` + sessionColumns + ` FROM user_sessions WHERE user_id = $1 AND is_active = TRUE ORDER BY last_activity_at DESC`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, userID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*auth.UserSession, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		s, err := scanSessionFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, s)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *sessionRepository) Update(ctx context.Context, s *auth.UserSession) error {
	const q = `UPDATE user_sessions SET
		is_active = $1, locked_at = $2, locked_reason = $3,
		last_activity_at = $4
	 WHERE id = $5`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.IsActive, persistence.NullIfZeroTime(s.LockedAt), persistence.NullIfEmpty(s.LockedReason),
		s.LastActivityAt, s.ID,
	)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *sessionRepository) DeactivateAll(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE user_sessions SET is_active = FALSE WHERE user_id = $1 AND is_active = TRUE`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, userID)
	return persistence.Translate(err)
}

func (r *sessionRepository) Deactivate(ctx context.Context, sessionID uuid.UUID) error {
	const q = `UPDATE user_sessions SET is_active = FALSE WHERE id = $1 AND is_active = TRUE`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, sessionID)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *sessionRepository) CleanExpired(ctx context.Context) (int64, error) {
	const q = `DELETE FROM user_sessions WHERE expires_at < $1 AND is_active = FALSE`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, time.Now().UTC())
	if err != nil {
		return 0, persistence.Translate(err)
	}
	return res.RowsAffected()
}

func scanSession(row *sql.Row) (*auth.UserSession, error) {
	s := &auth.UserSession{}
	var (
		lockedAt     sql.NullTime
		lockedReason sql.NullString
	)
	err := persistence.ScanRow(row,
		&s.ID, &s.UserID, &s.Token, &s.IPAddress, &s.UserAgent, &s.Device,
		&s.IsActive, &lockedAt, &lockedReason,
		&s.ExpiresAt, &s.LastActivityAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lockedAt.Valid {
		t := lockedAt.Time
		s.LockedAt = &t
	}
	if lockedReason.Valid {
		s.LockedReason = lockedReason.String
	}
	return s, nil
}

func scanSessionFromRows(rows *sql.Rows) (*auth.UserSession, error) {
	s := &auth.UserSession{}
	var (
		lockedAt     sql.NullTime
		lockedReason sql.NullString
	)
	if err := rows.Scan(
		&s.ID, &s.UserID, &s.Token, &s.IPAddress, &s.UserAgent, &s.Device,
		&s.IsActive, &lockedAt, &lockedReason,
		&s.ExpiresAt, &s.LastActivityAt, &s.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if lockedAt.Valid {
		t := lockedAt.Time
		s.LockedAt = &t
	}
	if lockedReason.Valid {
		s.LockedReason = lockedReason.String
	}
	return s, nil
}
