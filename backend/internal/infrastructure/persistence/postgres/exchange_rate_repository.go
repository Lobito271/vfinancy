package postgres

import (
	"context"
	"database/sql"
	"time"

	"vfinancy/backend/internal/domain/repositories"
)

type exchangeRateRepository struct {
	q Querier
}

func newExchangeRateRepository(db *sql.DB) *exchangeRateRepository {
	return &exchangeRateRepository{q: &dbBox{db: db}}
}

func newExchangeRateRepositoryTx(tx *sql.Tx) *exchangeRateRepository {
	return &exchangeRateRepository{q: &txBox{tx: tx}}
}

func (r *exchangeRateRepository) Upsert(ctx context.Context, from, to string, rate string, effectiveDate string, source string) error {
	const q = `INSERT INTO exchange_rates (company_id, from_currency, to_currency, rate, rate_date, source, created_at)
	VALUES (NULL, $1, $2, $3, $4, $5, $6)
	ON CONFLICT (from_currency, to_currency, rate_date) DO UPDATE SET
		rate = EXCLUDED.rate,
		source = EXCLUDED.source,
		created_at = EXCLUDED.created_at`
	_, err := r.q.ExecContext(ctx, q, from, to, rate, effectiveDate, source, time.Now().UTC())
	return Translate(err)
}

func (r *exchangeRateRepository) GetForDate(ctx context.Context, from, to string, date string) (string, error) {
	const q = `SELECT rate FROM exchange_rates
	WHERE from_currency = $1 AND to_currency = $2 AND rate_date <= $3
	ORDER BY rate_date DESC LIMIT 1`
	var rate string
	err := r.q.QueryRowContext(ctx, q, from, to, date).Scan(&rate)
	if err != nil {
		if isPgNoRows(err) {
			return "", repositories.ErrNotFound
		}
		return "", Translate(err)
	}
	return rate, nil
}

func (r *exchangeRateRepository) GetLatest(ctx context.Context, from, to string) (string, error) {
	const q = `SELECT rate FROM exchange_rates
	WHERE from_currency = $1 AND to_currency = $2
	ORDER BY rate_date DESC LIMIT 1`
	var rate string
	err := r.q.QueryRowContext(ctx, q, from, to).Scan(&rate)
	if err != nil {
		if isPgNoRows(err) {
			return "", repositories.ErrNotFound
		}
		return "", Translate(err)
	}
	return rate, nil
}

var _ repositories.ExchangeRateRepository = (*exchangeRateRepository)(nil)
