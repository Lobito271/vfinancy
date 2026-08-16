package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/features/sales"
)

type accountsReceivableRepository struct {
	q persistence.Querier
}

func NewAccountsReceivableRepository(db *sql.DB) *accountsReceivableRepository {
	return &accountsReceivableRepository{q: persistence.FromDB(db)}
}

func (r *accountsReceivableRepository) GetOpenBalanceForCustomer(ctx context.Context, customerID uuid.UUID) (string, error) {
	var f float64
	err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COALESCE(SUM(CAST(total AS REAL) - CAST(paid_amount AS REAL)), 0)
		   FROM sales
		  WHERE customer_id = $1 AND deleted_at IS NULL
		    AND status IN ('pending', 'partial')`,
		customerID).Scan(&f)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return "0.00", nil
		}
		return "", persistence.Translate(err)
	}
	return decimal.NewFromFloat(f).StringFixed(2), nil
}

func (r *accountsReceivableRepository) ListAgingBucket(ctx context.Context, customerID uuid.UUID) (map[string]string, error) {
	const q = `SELECT due_date, (CAST(total AS REAL) - CAST(paid_amount AS REAL))
		   FROM sales
		  WHERE customer_id = $1 AND deleted_at IS NULL
		    AND status IN ('pending', 'partial')
		    AND CAST(total AS REAL) > CAST(paid_amount AS REAL)`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, customerID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	buckets := map[string]decimal.Decimal{
		"0-30":  decimal.Zero,
		"31-60": decimal.Zero,
		"61-90": decimal.Zero,
		"90+":   decimal.Zero,
	}
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		var (
			dueDate sql.NullTime
			balance float64
		)
		if err := r.Scan(&dueDate, &balance); err != nil {
			return persistence.Translate(err)
		}
		if !dueDate.Valid || dueDate.Time.IsZero() {
			return nil
		}
		daysOverdue := time.Since(dueDate.Time).Hours() / 24
		if daysOverdue <= 0 {
			return nil
		}
		var key string
		switch {
		case daysOverdue <= 30:
			key = "0-30"
		case daysOverdue <= 60:
			key = "31-60"
		case daysOverdue <= 90:
			key = "61-90"
		default:
			key = "90+"
		}
		buckets[key] = buckets[key].Add(decimal.NewFromFloat(balance))
		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]string{
		"0-30":  buckets["0-30"].StringFixed(2),
		"31-60": buckets["31-60"].StringFixed(2),
		"61-90": buckets["61-90"].StringFixed(2),
		"90+":   buckets["90+"].StringFixed(2),
	}, nil
}

var _ sales.AccountsReceivableRepository = (*accountsReceivableRepository)(nil)
