package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/administration"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
)

type settingRepository struct {
	q persistence.Querier
}

func NewSettingRepository(db *sql.DB) *settingRepository {
	return &settingRepository{q: persistence.FromDB(db)}
}


const settingColumns = `
	id, company_id, key, value, category, label, description, is_public,
	created_at, updated_at, updated_by
`

func (r *settingRepository) Upsert(ctx context.Context, s *administration.ApplicationSetting) error {
	const q = `INSERT INTO application_settings (
		id, company_id, key, value, category, label, description, is_public,
		created_at, updated_at, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (company_id, key) DO UPDATE SET
		value = EXCLUDED.value,
		category = EXCLUDED.category,
		label = EXCLUDED.label,
		description = EXCLUDED.description,
		is_public = EXCLUDED.is_public,
		updated_at = EXCLUDED.updated_at,
		updated_by = EXCLUDED.updated_by`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.ID, s.CompanyID, s.Key, []byte(s.Value), s.Category, s.Label, s.Description, s.IsPublic,
		s.CreatedAt, s.UpdatedAt, persistence.NullIfEmptyUUID(s.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *settingRepository) GetByKey(ctx context.Context, companyID uuid.UUID, key string) (*administration.ApplicationSetting, error) {
	q := `SELECT ` + settingColumns + ` FROM application_settings WHERE company_id = $1 AND key = $2`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, key)
	return scanSetting(row)
}

func (r *settingRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]*administration.ApplicationSetting, error) {
	q := `SELECT ` + settingColumns + ` FROM application_settings WHERE company_id = $1 ORDER BY category, key`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, companyID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	return scanSettings(rows)
}

func (r *settingRepository) ListByCategory(ctx context.Context, companyID uuid.UUID, category string) ([]*administration.ApplicationSetting, error) {
	q := `SELECT ` + settingColumns + ` FROM application_settings WHERE company_id = $1 AND category = $2 ORDER BY key`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, companyID, category)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	return scanSettings(rows)
}

func (r *settingRepository) Delete(ctx context.Context, companyID uuid.UUID, key string) error {
	const q = `DELETE FROM application_settings WHERE company_id = $1 AND key = $2`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, companyID, key)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func scanSetting(row *sql.Row) (*administration.ApplicationSetting, error) {
	s := &administration.ApplicationSetting{}
	var (
		value     []byte
		updatedBy sql.NullString
	)
	err := row.Scan(
		&s.ID, &s.CompanyID, &s.Key, &value, &s.Category, &s.Label, &s.Description, &s.IsPublic,
		&s.CreatedAt, &s.UpdatedAt, &updatedBy,
	)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, persistence.Translate(err)
	}
	s.Value = json.RawMessage(value)
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		s.UpdatedBy = &id
	}
	return s, nil
}

func scanSettings(rows *sql.Rows) ([]*administration.ApplicationSetting, error) {
	out := make([]*administration.ApplicationSetting, 0)
	err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		s := &administration.ApplicationSetting{}
		var (
			value     []byte
			updatedBy sql.NullString
		)
		if err := r.Scan(
			&s.ID, &s.CompanyID, &s.Key, &value, &s.Category, &s.Label, &s.Description, &s.IsPublic,
			&s.CreatedAt, &s.UpdatedAt, &updatedBy,
		); err != nil {
			return persistence.Translate(err)
		}
		s.Value = json.RawMessage(value)
		if updatedBy.Valid {
			id := persistence.ParseUUID(updatedBy.String)
			s.UpdatedBy = &id
		}
		out = append(out, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _ administration.SettingRepository = (*settingRepository)(nil)


