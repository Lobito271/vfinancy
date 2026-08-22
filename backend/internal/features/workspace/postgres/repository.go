package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/workspace"
)

type repository struct{ q persistence.Querier }

func NewRepository(db *sql.DB) *repository { return &repository{q: persistence.FromDB(db)} }

const companyColumns = `id, code, legal_name,
 COALESCE(trade_name, ''), tax_id, COALESCE(address, ''), COALESCE(phone, ''),
 COALESCE(email, ''),
 country_code, functional_currency_code, timezone, fiscal_year_start_month,
 is_active, created_at, updated_at, deleted_at`

func (r *repository) GetProfile(ctx context.Context) (*workspace.LocalProfile, error) {
	const q = `SELECT id, name, password_hash, password_enabled, failed_attempts,
 locked_until, active_company_id, theme, language, date_format, number_format,
 decimal_places, timezone, created_at, updated_at FROM local_profiles LIMIT 1`
	p := &workspace.LocalProfile{}
	var hash sql.NullString
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q).Scan(
		&p.ID, &p.Name, &hash, &p.PasswordEnabled, &p.FailedAttempts,
		&p.LockedUntil, &p.ActiveCompanyID, &p.Theme, &p.Language,
		&p.DateFormat, &p.NumberFormat, &p.DecimalPlaces, &p.Timezone,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, workspace.ErrProfileNotFound
		}
		return nil, persistence.Translate(err)
	}
	if hash.Valid {
		p.PasswordHash = hash.String
	}
	return p, nil
}

func (r *repository) CreateProfile(ctx context.Context, p *workspace.LocalProfile) error {
	const q = `INSERT INTO local_profiles
 (id, name, password_hash, password_enabled, failed_attempts, locked_until,
  active_company_id, theme, language, date_format, number_format, decimal_places,
  timezone, created_at, updated_at)
 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, p.ID, p.Name,
		p.PasswordHash, p.PasswordEnabled, p.FailedAttempts, p.LockedUntil,
		p.ActiveCompanyID, p.Theme, p.Language, p.DateFormat, p.NumberFormat,
		p.DecimalPlaces, p.Timezone, p.CreatedAt, p.UpdatedAt)
	return persistence.Translate(err)
}

func (r *repository) UpdateProfile(ctx context.Context, p *workspace.LocalProfile) error {
	const q = `UPDATE local_profiles SET name = $2, password_hash = $3,
 password_enabled = $4, failed_attempts = $5, locked_until = $6,
 active_company_id = $7, theme = $8, language = $9, date_format = $10,
 number_format = $11, decimal_places = $12, timezone = $13, updated_at = $14
 WHERE id = $1`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, p.ID, p.Name,
		p.PasswordHash, p.PasswordEnabled, p.FailedAttempts, p.LockedUntil,
		p.ActiveCompanyID, p.Theme, p.Language, p.DateFormat, p.NumberFormat,
		p.DecimalPlaces, p.Timezone, p.UpdatedAt)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *repository) ListCompanies(ctx context.Context) ([]*workspace.Company, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT `+companyColumns+` FROM companies WHERE deleted_at IS NULL ORDER BY legal_name`)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	return scanCompanies(rows)
}

func (r *repository) GetCompany(ctx context.Context, id uuid.UUID) (*workspace.Company, error) {
	c := &workspace.Company{}
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+companyColumns+` FROM companies WHERE id = $1`, id)
	if err := scanCompany(row, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *repository) CreateCompany(ctx context.Context, c *workspace.Company) error {
	const q = `INSERT INTO companies
 (id, code, legal_name, trade_name, tax_id, address, phone, email,
  country_code, functional_currency_code, timezone, fiscal_year_start_month,
  is_active, created_at, updated_at)
 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, c.ID, c.Code, c.LegalName,
		c.TradeName, c.TaxID, c.Address, c.Phone, c.Email, c.CountryCode,
		c.FunctionalCurrency, c.Timezone, c.FiscalYearStartMonth, c.IsActive,
		c.CreatedAt, c.UpdatedAt)
	return persistence.Translate(err)
}

func (r *repository) UpdateCompany(ctx context.Context, c *workspace.Company) error {
	const q = `UPDATE companies SET code = $2, legal_name = $3, trade_name = $4,
 tax_id = $5, address = $6, phone = $7, email = $8, country_code = $9,
 functional_currency_code = $10, timezone = $11, fiscal_year_start_month = $12,
 is_active = $13, updated_at = $14 WHERE id = $1`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, c.ID, c.Code, c.LegalName,
		c.TradeName, c.TaxID, c.Address, c.Phone, c.Email, c.CountryCode,
		c.FunctionalCurrency, c.Timezone, c.FiscalYearStartMonth, c.IsActive, c.UpdatedAt)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func scanCompany(row interface{ Scan(...any) error }, c *workspace.Company) error {
	if err := row.Scan(&c.ID, &c.Code, &c.LegalName, &c.TradeName, &c.TaxID,
		&c.Address, &c.Phone, &c.Email, &c.CountryCode, &c.FunctionalCurrency,
		&c.Timezone, &c.FiscalYearStartMonth, &c.IsActive, &c.CreatedAt,
		&c.UpdatedAt, &c.DeletedAt); err != nil {
		if persistence.IsPgNoRows(err) {
			return repositories.ErrNotFound
		}
		return persistence.Translate(err)
	}
	return nil
}

func scanCompanies(rows *sql.Rows) ([]*workspace.Company, error) {
	companies := make([]*workspace.Company, 0)
	err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		c := &workspace.Company{}
		if err := scanCompany(row, c); err != nil {
			return err
		}
		companies = append(companies, c)
		return nil
	})
	return companies, err
}

var _ workspace.Repository = (*repository)(nil)
