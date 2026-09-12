package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/treasury"
)

type exchangeRateRepository struct {
	q persistence.Querier
}

// NewExchangeRateRepository returns an ExchangeRateRepository backed by
// the given database handle.
func NewExchangeRateRepository(db *sql.DB) *exchangeRateRepository {
	return &exchangeRateRepository{q: persistence.FromDB(db)}
}

func (r *exchangeRateRepository) Upsert(ctx context.Context, from, to valueobjects.CurrencyCode, rateDate time.Time, rate valueobjects.Money, source string) error {
	const q = `INSERT INTO exchange_rates (
		id, from_currency, to_currency, rate_date, rate, source, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (from_currency, to_currency, rate_date) DO UPDATE SET
		rate = EXCLUDED.rate,
		source = EXCLUDED.source,
		created_at = EXCLUDED.created_at`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		uuid.New(), from.String(), to.String(),
		rateDate.Format("2006-01-02"), rate.String(), source, time.Now().UTC(),
	)
	return persistence.Translate(err)
}

func (r *exchangeRateRepository) Latest(ctx context.Context, from, to valueobjects.CurrencyCode) (*treasury.ExchangeRateSnapshot, error) {
	const q = `SELECT from_currency, to_currency, rate_date, rate, source FROM exchange_rates
	WHERE from_currency = $1 AND to_currency = $2
	ORDER BY rate_date DESC LIMIT 1`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, from.String(), to.String())
	snap := &treasury.ExchangeRateSnapshot{}
	var (
		fromStr, toStr, rate string
		rateDate             any
	)
	err := persistence.ScanRow(row, &fromStr, &toStr, &rateDate, &rate, &snap.Source)
	if err != nil {
		return nil, err
	}
	if snap.From, err = valueobjects.NewCurrencyCode(fromStr); err != nil {
		return nil, err
	}
	if snap.To, err = valueobjects.NewCurrencyCode(toStr); err != nil {
		return nil, err
	}
	date, err := decodeRateDate(rateDate)
	if err != nil {
		return nil, err
	}
	snap.RateDate = date
	money, err := persistence.ParseMoney(rate)
	if err != nil {
		return nil, err
	}
	snap.Rate = money
	return snap, nil
}

// decodeRateDate normalises a rate_date value: SQLite stores DATE
// columns as TEXT, PostgreSQL as time.Time.
func decodeRateDate(v any) (time.Time, error) {
	switch t := v.(type) {
	case time.Time:
		return t, nil
	case string:
		var formats = []string{"2006-01-02", "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05"}
		for _, f := range formats {
			if parsed, err := time.Parse(f, t); err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, fmt.Errorf("exchange rate: invalid rate_date %q", t)
	default:
		return time.Time{}, fmt.Errorf("exchange rate: unsupported rate_date type %T", v)
	}
}

var _ treasury.ExchangeRateRepository = (*exchangeRateRepository)(nil)
