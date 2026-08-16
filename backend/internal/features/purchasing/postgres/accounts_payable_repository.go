package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/features/purchasing"
	"vfinancy/backend/infrastructure/persistence"
)

type accountsPayableRepository struct {
	q persistence.Querier
}

func NewAccountsPayableRepository(db *sql.DB) *accountsPayableRepository {
	return &accountsPayableRepository{q: persistence.FromDB(db)}
}

func (r *accountsPayableRepository) GetOpenBalanceForSupplier(ctx context.Context, supplierID uuid.UUID) (string, error) {
	const q = `SELECT COALESCE(SUM(CAST(total AS REAL) - CAST(paid_amount AS REAL)), 0)
		FROM purchase_orders
		WHERE supplier_id = $1 AND deleted_at IS NULL
		  AND status IN ('pending', 'received', 'paid')`
	var balance float64
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, supplierID).Scan(&balance); err != nil {
		return "", persistence.Translate(err)
	}
	return decimal.NewFromFloat(balance).StringFixed(2), nil
}

func (r *accountsPayableRepository) ListAgingBucket(ctx context.Context, supplierID uuid.UUID) (map[string]string, error) {
	const q = `SELECT expected_date, (CAST(total AS REAL) - CAST(paid_amount AS REAL))
		FROM purchase_orders
		WHERE supplier_id = $1 AND deleted_at IS NULL
		  AND status IN ('pending', 'received', 'paid')
		  AND CAST(total AS REAL) > CAST(paid_amount AS REAL)`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, supplierID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	buckets := map[string]decimal.Decimal{
		"0-30":  decimal.Zero,
		"31-60": decimal.Zero,
		"61-90": decimal.Zero,
		"90+":   decimal.Zero,
	}
	now := time.Now().UTC()
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		var (
			expectedDate sql.NullTime
			balance      float64
		)
		if err := r.Scan(&expectedDate, &balance); err != nil {
			return persistence.Translate(err)
		}
		if !expectedDate.Valid || expectedDate.Time.IsZero() {
			return nil
		}
		days := int(now.Sub(expectedDate.Time).Hours() / 24)
		if days <= 0 {
			return nil
		}
		key := "90+"
		switch {
		case days <= 30:
			key = "0-30"
		case days <= 60:
			key = "31-60"
		case days <= 90:
			key = "61-90"
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

var _ purchasing.AccountsPayableRepository = (*accountsPayableRepository)(nil)
