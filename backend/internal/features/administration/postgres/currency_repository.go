package postgres

import (
	"context"
	"database/sql"
	"time"

	"vfinancy/backend/internal/features/administration"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

type currencyRepository struct {
	q persistence.Querier
}

func NewCurrencyRepository(db *sql.DB) *currencyRepository {
	return &currencyRepository{q: persistence.FromDB(db)}
}


const currencyColumns = `
	code, symbol, name, decimal_places, type, is_active, created_at
`

func (r *currencyRepository) GetByCode(ctx context.Context, code string) (*administration.Currency, error) {
	q := `SELECT ` + currencyColumns + ` FROM currencies WHERE code = $1`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, code)
	return scanCurrency(row)
}

func (r *currencyRepository) List(ctx context.Context) ([]*administration.Currency, error) {
	q := `SELECT ` + currencyColumns + ` FROM currencies ORDER BY code`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*administration.Currency, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		c := &administration.Currency{}
		var code, currencyType string
		if err := r.Scan(&code, &c.Symbol, &c.Name, &c.DecimalPlaces, &currencyType, &c.IsActive, &c.CreatedAt); err != nil {
			return persistence.Translate(err)
		}
		c.Code = valueobjects.MustCurrencyCode(code)
		c.Type = enums.CurrencyType(currencyType)
		out = append(out, c)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *currencyRepository) Upsert(ctx context.Context, c *administration.Currency) error {
	const q = `INSERT INTO currencies (code, symbol, name, decimal_places, type, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (code) DO UPDATE SET
		symbol = EXCLUDED.symbol,
		name = EXCLUDED.name,
		decimal_places = EXCLUDED.decimal_places,
		type = EXCLUDED.type,
		is_active = EXCLUDED.is_active,
		updated_at = EXCLUDED.updated_at`
	now := time.Now().UTC()
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		c.Code.String(), c.Symbol, c.Name, c.DecimalPlaces, string(c.Type), c.IsActive, c.CreatedAt, now,
	)
	return persistence.Translate(err)
}

func scanCurrency(row *sql.Row) (*administration.Currency, error) {
	c := &administration.Currency{}
	var code, currencyType string
	err := row.Scan(&code, &c.Symbol, &c.Name, &c.DecimalPlaces, &currencyType, &c.IsActive, &c.CreatedAt)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, persistence.Translate(err)
	}
	c.Code = valueobjects.MustCurrencyCode(code)
	c.Type = enums.CurrencyType(currencyType)
	return c, nil
}

var _ administration.CurrencyRepository = (*currencyRepository)(nil)
