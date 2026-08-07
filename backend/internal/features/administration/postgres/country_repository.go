package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"vfinancy/backend/internal/features/administration"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
)

type countryRepository struct {
	q persistence.Querier
}

func NewCountryRepository(db *sql.DB) *countryRepository {
	return &countryRepository{q: persistence.FromDB(db)}
}


const countryColumns = `
	code, name, locale, currency_code, tax_id_label, personal_id_label,
	default_document_types, created_at
`

func (r *countryRepository) GetByCode(ctx context.Context, code string) (*administration.Country, error) {
	q := `SELECT ` + countryColumns + ` FROM countries WHERE code = $1`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, code)
	return scanCountry(row)
}

func (r *countryRepository) List(ctx context.Context) ([]*administration.Country, error) {
	q := `SELECT ` + countryColumns + ` FROM countries ORDER BY code`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*administration.Country, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		c := &administration.Country{}
		var docTypes string
		if err := r.Scan(&c.Code, &c.Name, &c.Locale, &c.Currency, &c.TaxIDLabel, &c.PersonalIDLabel, &docTypes, &c.CreatedAt); err != nil {
			return persistence.Translate(err)
		}
		c.DefaultDocumentTypes = parseTextArray(docTypes)
		out = append(out, c)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *countryRepository) Upsert(ctx context.Context, c *administration.Country) error {
	// default_document_types is text[] on PostgreSQL and a {a,b,c}
	// literal string on SQLite. The {..} text form is what the shared
	// scan helper (parseTextArray) reads back, so both dialects stay
	// byte-compatible at the repository level.
	var (
		args    []any
		arrayExpr string
	)
	if persistence.IsSQLite() {
		arrayExpr = persistence.ArrayLiteral(c.DefaultDocumentTypes)
	} else {
		arrayExpr = "ARRAY[]::text[]"
		if len(c.DefaultDocumentTypes) > 0 {
			placeholders := make([]string, len(c.DefaultDocumentTypes))
			for i, dt := range c.DefaultDocumentTypes {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				args = append(args, dt)
			}
			arrayExpr = fmt.Sprintf("ARRAY[%s]::text[]", strings.Join(placeholders, ","))
		}
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
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, args...)
	return persistence.Translate(err)
}

func scanCountry(row *sql.Row) (*administration.Country, error) {
	c := &administration.Country{}
	var docTypes string
	err := row.Scan(&c.Code, &c.Name, &c.Locale, &c.Currency, &c.TaxIDLabel, &c.PersonalIDLabel, &docTypes, &c.CreatedAt)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, persistence.Translate(err)
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

var _ administration.CountryRepository = (*countryRepository)(nil)
