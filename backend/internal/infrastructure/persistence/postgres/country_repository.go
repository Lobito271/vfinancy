package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"vfinancy/backend/internal/domain/entities/masterdata"
	"vfinancy/backend/internal/domain/repositories"
)

type countryRepository struct {
	q Querier
}

func newCountryRepository(db *sql.DB) *countryRepository {
	return &countryRepository{q: &dbBox{db: db}}
}

func newCountryRepositoryTx(tx *sql.Tx) *countryRepository {
	return &countryRepository{q: &txBox{tx: tx}}
}

const countryColumns = `
	code, name, locale, currency_code, tax_id_label, personal_id_label,
	default_document_types, created_at
`

func (r *countryRepository) GetByCode(ctx context.Context, code string) (*masterdata.Country, error) {
	q := `SELECT ` + countryColumns + ` FROM countries WHERE code = $1`
	row := r.q.QueryRowContext(ctx, q, code)
	return scanCountry(row)
}

func (r *countryRepository) List(ctx context.Context) ([]*masterdata.Country, error) {
	q := `SELECT ` + countryColumns + ` FROM countries ORDER BY code`
	rows, err := r.q.QueryContext(ctx, q)
	if err != nil {
		return nil, Translate(err)
	}
	out := make([]*masterdata.Country, 0)
	if err := scanRows(rows, func(r *sql.Rows) error {
		c := &masterdata.Country{}
		var docTypes string
		if err := r.Scan(&c.Code, &c.Name, &c.Locale, &c.Currency, &c.TaxIDLabel, &c.PersonalIDLabel, &docTypes, &c.CreatedAt); err != nil {
			return Translate(err)
		}
		c.DefaultDocumentTypes = parseTextArray(docTypes)
		out = append(out, c)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *countryRepository) Upsert(ctx context.Context, c *masterdata.Country) error {
	var (
		placeholders []string
		args         []any
	)
	for i, dt := range c.DefaultDocumentTypes {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, dt)
	}
	arrayExpr := "ARRAY[]::text[]"
	if len(placeholders) > 0 {
		arrayExpr = fmt.Sprintf("ARRAY[%s]::text[]", strings.Join(placeholders, ","))
	}
	arrayPos := len(args) + 1
	args = append(args, c.Code)
	args = append(args, c.Name)
	args = append(args, c.Locale)
	args = append(args, c.Currency)
	args = append(args, c.TaxIDLabel)
	args = append(args, c.PersonalIDLabel)

	q := fmt.Sprintf(`INSERT INTO countries (code, name, locale, currency_code, tax_id_label, personal_id_label, default_document_types)
	VALUES ($%d, $%d, $%d, $%d, $%d, $%d, %s)
	ON CONFLICT (code) DO UPDATE SET
		name = EXCLUDED.name,
		locale = EXCLUDED.locale,
		currency_code = EXCLUDED.currency_code,
		tax_id_label = EXCLUDED.tax_id_label,
		personal_id_label = EXCLUDED.personal_id_label,
		default_document_types = EXCLUDED.default_document_types`,
		arrayPos, arrayPos+1, arrayPos+2, arrayPos+3, arrayPos+4, arrayPos+5, arrayExpr)
	_, err := r.q.ExecContext(ctx, q, args...)
	return Translate(err)
}

func scanCountry(row *sql.Row) (*masterdata.Country, error) {
	c := &masterdata.Country{}
	var docTypes string
	err := row.Scan(&c.Code, &c.Name, &c.Locale, &c.Currency, &c.TaxIDLabel, &c.PersonalIDLabel, &docTypes, &c.CreatedAt)
	if err != nil {
		if isPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, Translate(err)
	}
	c.DefaultDocumentTypes = parseTextArray(docTypes)
	return c, nil
}

func parseTextArray(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		return nil
	}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

var _ repositories.CountryRepository = (*countryRepository)(nil)
