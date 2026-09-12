package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/administration"
)

type settingRepository struct {
	q persistence.Querier
}

func NewSettingRepository(db *sql.DB) *settingRepository {
	return &settingRepository{q: persistence.FromDB(db)}
}

const settingColumns = `id, key, value, created_at, updated_at`

func (r *settingRepository) Upsert(ctx context.Context, s *administration.ApplicationSetting) error {
	const q = `INSERT INTO application_settings (id, key, value, created_at, updated_at)
	 VALUES ($1, $2, $3, $4, $5)
	 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.ID, s.Key, []byte(s.Value), s.CreatedAt, s.UpdatedAt)
	return persistence.Translate(err)
}

func (r *settingRepository) GetByKey(ctx context.Context, key string) (*administration.ApplicationSetting, error) {
	const q = `SELECT ` + settingColumns + ` FROM application_settings WHERE key = $1`
	s := &administration.ApplicationSetting{}
	var value []byte
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, key).Scan(
		&s.ID, &s.Key, &value, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, persistence.Translate(err)
	}
	s.Value = json.RawMessage(value)
	return s, nil
}

func (r *settingRepository) List(ctx context.Context) ([]*administration.ApplicationSetting, error) {
	const q = `SELECT ` + settingColumns + ` FROM application_settings ORDER BY key`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*administration.ApplicationSetting, 0)
	err = persistence.ScanRows(rows, func(row *sql.Rows) error {
		s := &administration.ApplicationSetting{}
		var value []byte
		if err := row.Scan(&s.ID, &s.Key, &value, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return persistence.Translate(err)
		}
		s.Value = json.RawMessage(value)
		out = append(out, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _ administration.SettingRepository = (*settingRepository)(nil)
