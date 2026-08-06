package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/administration"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

type taxRepository struct {
	q persistence.Querier
}

func NewTaxRepository(db *sql.DB) *taxRepository {
	return &taxRepository{q: persistence.FromDB(db)}
}


const taxColumns = `
	id, company_id, code, name, short_name, country_code, default_rate,
	is_inclusive, is_percentage, category, is_active, created_at, updated_at
`

func (r *taxRepository) Create(ctx context.Context, t *administration.Tax) error {
	const q = `INSERT INTO taxes (
		id, company_id, code, name, short_name, country_code, default_rate,
		is_inclusive, is_percentage, category, is_active, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		t.ID, nil, t.Code, t.Name, t.ShortName, t.CountryCode, t.DefaultRate.String(),
		t.IsInclusive, t.IsPercentage, string(t.Category), t.IsActive, t.CreatedAt, t.UpdatedAt,
	)
	return persistence.Translate(err)
}

func (r *taxRepository) Update(ctx context.Context, t *administration.Tax) error {
	const q = `UPDATE taxes SET
		code = $1, name = $2, short_name = $3, country_code = $4, default_rate = $5,
		is_inclusive = $6, is_percentage = $7, category = $8, is_active = $9,
		updated_at = $10
	WHERE id = $11 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		t.Code, t.Name, t.ShortName, t.CountryCode, t.DefaultRate.String(),
		t.IsInclusive, t.IsPercentage, string(t.Category), t.IsActive,
		time.Now().UTC(), t.ID,
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

func (r *taxRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE taxes SET deleted_at = $1, updated_at = $2, is_active = FALSE WHERE id = $3 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, now, now, id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *taxRepository) GetByID(ctx context.Context, id uuid.UUID) (*administration.Tax, error) {
	q := `SELECT ` + taxColumns + ` FROM taxes WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanTax(row)
}

func (r *taxRepository) GetByCode(ctx context.Context, companyID *uuid.UUID, code string) (*administration.Tax, error) {
	const q = `SELECT ` + taxColumns + ` FROM taxes WHERE code = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, code)
	return scanTax(row)
}

func (r *taxRepository) List(ctx context.Context, companyID *uuid.UUID) ([]*administration.Tax, error) {
	const q = `SELECT ` + taxColumns + ` FROM taxes WHERE deleted_at IS NULL ORDER BY code`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*administration.Tax, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		t, err := scanTaxFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, t)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func scanTax(row *sql.Row) (*administration.Tax, error) {
	t := &administration.Tax{}
	var (
		companyID   sql.NullString
		code, name  string
		shortName   string
		countryCode string
		rate        string
		category    string
	)
	err := row.Scan(
		&t.ID, &companyID, &code, &name, &shortName, &countryCode, &rate,
		&t.IsInclusive, &t.IsPercentage, &category, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, persistence.Translate(err)
	}
	t.Code = code
	t.Name = name
	t.ShortName = shortName
	t.CountryCode = countryCode
	t.DefaultRate, _ = valueobjects.PercentageFromString(rate)
	t.Category = enums.TaxCategory(category)
	return t, nil
}

func scanTaxFromRows(rows *sql.Rows) (*administration.Tax, error) {
	t := &administration.Tax{}
	var (
		companyID   sql.NullString
		code, name  string
		shortName   string
		countryCode string
		rate        string
		category    string
	)
	if err := rows.Scan(
		&t.ID, &companyID, &code, &name, &shortName, &countryCode, &rate,
		&t.IsInclusive, &t.IsPercentage, &category, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	t.Code = code
	t.Name = name
	t.ShortName = shortName
	t.CountryCode = countryCode
	t.DefaultRate, _ = valueobjects.PercentageFromString(rate)
	t.Category = enums.TaxCategory(category)
	return t, nil
}

var _ administration.TaxRepository = (*taxRepository)(nil)
