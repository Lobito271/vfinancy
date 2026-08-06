package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/administration"
	"vfinancy/backend/internal/domain/repositories"
)

type settingRepository struct {
	q Querier
}

func newSettingRepository(db *sql.DB) *settingRepository {
	return &settingRepository{q: &dbBox{db: db}}
}

func newSettingRepositoryTx(tx *sql.Tx) *settingRepository {
	return &settingRepository{q: &txBox{tx: tx}}
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
	_, err := r.q.ExecContext(ctx, q,
		s.ID, s.CompanyID, s.Key, []byte(s.Value), s.Category, s.Label, s.Description, s.IsPublic,
		s.CreatedAt, s.UpdatedAt, nullIfEmptyUUID(s.UpdatedBy),
	)
	return Translate(err)
}

func (r *settingRepository) GetByKey(ctx context.Context, companyID uuid.UUID, key string) (*administration.ApplicationSetting, error) {
	q := `SELECT ` + settingColumns + ` FROM application_settings WHERE company_id = $1 AND key = $2`
	row := r.q.QueryRowContext(ctx, q, companyID, key)
	return scanSetting(row)
}

func (r *settingRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]*administration.ApplicationSetting, error) {
	q := `SELECT ` + settingColumns + ` FROM application_settings WHERE company_id = $1 ORDER BY category, key`
	rows, err := r.q.QueryContext(ctx, q, companyID)
	if err != nil {
		return nil, Translate(err)
	}
	return scanSettings(rows)
}

func (r *settingRepository) ListByCategory(ctx context.Context, companyID uuid.UUID, category string) ([]*administration.ApplicationSetting, error) {
	q := `SELECT ` + settingColumns + ` FROM application_settings WHERE company_id = $1 AND category = $2 ORDER BY key`
	rows, err := r.q.QueryContext(ctx, q, companyID, category)
	if err != nil {
		return nil, Translate(err)
	}
	return scanSettings(rows)
}

func (r *settingRepository) Delete(ctx context.Context, companyID uuid.UUID, key string) error {
	const q = `DELETE FROM application_settings WHERE company_id = $1 AND key = $2`
	res, err := r.q.ExecContext(ctx, q, companyID, key)
	if err != nil {
		return Translate(err)
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
		if isPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, Translate(err)
	}
	s.Value = json.RawMessage(value)
	if updatedBy.Valid {
		id := masterdataParseUUID(updatedBy.String)
		s.UpdatedBy = &id
	}
	return s, nil
}

func scanSettings(rows *sql.Rows) ([]*administration.ApplicationSetting, error) {
	out := make([]*administration.ApplicationSetting, 0)
	err := scanRows(rows, func(r *sql.Rows) error {
		s := &administration.ApplicationSetting{}
		var (
			value     []byte
			updatedBy sql.NullString
		)
		if err := r.Scan(
			&s.ID, &s.CompanyID, &s.Key, &value, &s.Category, &s.Label, &s.Description, &s.IsPublic,
			&s.CreatedAt, &s.UpdatedAt, &updatedBy,
		); err != nil {
			return Translate(err)
		}
		s.Value = json.RawMessage(value)
		if updatedBy.Valid {
			id := masterdataParseUUID(updatedBy.String)
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

var _ repositories.SettingRepository = (*settingRepository)(nil)


