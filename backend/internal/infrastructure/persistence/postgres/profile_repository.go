package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/identity"
	"vfinancy/backend/internal/domain/repositories"
)

type profileRepository struct {
	q Querier
}

func newProfileRepository(db *sql.DB) *profileRepository {
	return &profileRepository{q: &dbBox{db: db}}
}

func newProfileRepositoryTx(tx *sql.Tx) *profileRepository {
	return &profileRepository{q: &txBox{tx: tx}}
}

const profileColumns = `
	id, user_id, avatar_url, theme, language, date_format, number_format,
	decimal_places, timezone, created_at, updated_at
`

func (r *profileRepository) Create(ctx context.Context, p *identity.UserProfile) error {
	const q = `INSERT INTO user_profiles (
		id, user_id, avatar_url, theme, language, date_format, number_format,
		decimal_places, timezone, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.q.ExecContext(ctx, q,
		p.ID, p.UserID, nullIfEmpty(p.AvatarURL), p.Theme, p.Language,
		p.DateFormat, p.NumberFormat, p.DecimalPlaces, p.Timezone,
		p.CreatedAt, p.UpdatedAt,
	)
	return Translate(err)
}

func (r *profileRepository) Update(ctx context.Context, p *identity.UserProfile) error {
	const q = `UPDATE user_profiles SET
		avatar_url = $1, theme = $2, language = $3, date_format = $4,
		number_format = $5, decimal_places = $6, timezone = $7, updated_at = $8
	 WHERE id = $9`
	res, err := r.q.ExecContext(ctx, q,
		nullIfEmpty(p.AvatarURL), p.Theme, p.Language, p.DateFormat,
		p.NumberFormat, p.DecimalPlaces, p.Timezone, time.Now().UTC(),
		p.ID,
	)
	if err != nil {
		return Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *profileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*identity.UserProfile, error) {
	q := `SELECT ` + profileColumns + ` FROM user_profiles WHERE user_id = $1`
	row := r.q.QueryRowContext(ctx, q, userID)
	return scanProfile(row)
}

func scanProfile(row *sql.Row) (*identity.UserProfile, error) {
	p := &identity.UserProfile{}
	var avatarURL sql.NullString
	err := scanRow(row,
		&p.ID, &p.UserID, &avatarURL, &p.Theme, &p.Language,
		&p.DateFormat, &p.NumberFormat, &p.DecimalPlaces, &p.Timezone,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if avatarURL.Valid {
		p.AvatarURL = avatarURL.String
	}
	return p, nil
}
