package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/treasury"
)

type creditCardRepository struct {
	q persistence.Querier
}

// NewCreditCardRepository returns a CreditCardRepository backed by the
// given database handle.
func NewCreditCardRepository(db *sql.DB) *creditCardRepository {
	return &creditCardRepository{q: persistence.FromDB(db)}
}

const creditCardColumns = `
	id, issuer, last_four, credit_limit, current_balance,
	cut_off_day, payment_due_day, is_active,
	created_at, updated_at, deleted_at, created_by, updated_by
`

const creditCardInsert = `INSERT INTO credit_cards (
	id, issuer, last_four, credit_limit, current_balance,
	cut_off_day, payment_due_day, is_active,
	created_at, updated_at, deleted_at, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

func (r *creditCardRepository) Create(ctx context.Context, c *treasury.CreditCard) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, creditCardInsert,
		c.ID, c.Issuer, c.LastFour,
		c.CreditLimit.String(), c.CurrentBalance.String(),
		c.CutOffDay, c.PaymentDueDay, c.IsActive,
		c.CreatedAt, c.UpdatedAt, persistence.NullIfZeroTime(c.DeletedAt),
		persistence.NullIfEmptyUUID(c.CreatedBy), persistence.NullIfEmptyUUID(c.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *creditCardRepository) Update(ctx context.Context, c *treasury.CreditCard) error {
	const q = `UPDATE credit_cards SET
		issuer = $1, last_four = $2, credit_limit = $3,
		current_balance = $4, cut_off_day = $5, payment_due_day = $6,
		is_active = $7, updated_at = $8, updated_by = $9
	 WHERE id = $10 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		c.Issuer, c.LastFour, c.CreditLimit.String(),
		c.CurrentBalance.String(), c.CutOffDay, c.PaymentDueDay,
		c.IsActive, time.Now().UTC(), persistence.NullIfEmptyUUID(c.UpdatedBy),
		c.ID,
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

func (r *creditCardRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE credit_cards SET deleted_at = $1, updated_at = $1, is_active = FALSE WHERE id = $2 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, now, id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *creditCardRepository) GetByID(ctx context.Context, id uuid.UUID) (*treasury.CreditCard, error) {
	q := `SELECT ` + creditCardColumns + ` FROM credit_cards WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanCreditCard(row)
}

func (r *creditCardRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*treasury.CreditCard, error) {
	// SQLite has no FOR UPDATE; its single-writer model plus the
	// BEGIN IMMEDIATE transaction gives the needed lock. Postgres
	// locks the row for the duration of the write transaction.
	lock := " FOR UPDATE"
	if persistence.IsSQLite() {
		lock = ""
	}
	q := `SELECT ` + creditCardColumns + ` FROM credit_cards WHERE id = $1 AND deleted_at IS NULL` + lock
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanCreditCard(row)
}

func (r *creditCardRepository) List(ctx context.Context) ([]*treasury.CreditCard, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT `+creditCardColumns+` FROM credit_cards WHERE deleted_at IS NULL ORDER BY issuer`)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*treasury.CreditCard, 0)
	if err := persistence.ScanRows(rows, func(rows *sql.Rows) error {
		c, err := scanCreditCardFromRows(rows)
		if err != nil {
			return err
		}
		out = append(out, c)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *creditCardRepository) CycleTotals(ctx context.Context, cardID uuid.UUID, from, to time.Time) (valueobjects.Money, valueobjects.Money, error) {
	fromStr, toStr := from.Format("2006-01-02"), to.Format("2006-01-02")
	const (
		qTotal = `SELECT COALESCE(SUM(cost_usd), 0) FROM purchase_orders
			WHERE credit_card_id = $1 AND deleted_at IS NULL
			AND order_date >= $2 AND order_date < $3 AND status <> 'cancelled'`
		qRefunds = `SELECT COALESCE(SUM(refund_amount), 0) FROM purchase_orders
			WHERE credit_card_id = $1 AND deleted_at IS NULL
			AND order_date >= $2 AND order_date < $3 AND status = 'cancelled'`
	)
	var totalStr, refundStr string
	q := persistence.Q(ctx, r.q)
	if err := q.QueryRowContext(ctx, qTotal, cardID, fromStr, toStr).Scan(&totalStr); err != nil {
		return valueobjects.Money{}, valueobjects.Money{}, persistence.Translate(err)
	}
	if err := q.QueryRowContext(ctx, qRefunds, cardID, fromStr, toStr).Scan(&refundStr); err != nil {
		return valueobjects.Money{}, valueobjects.Money{}, persistence.Translate(err)
	}
	total, err := persistence.ParseMoney(totalStr)
	if err != nil {
		return valueobjects.Money{}, valueobjects.Money{}, err
	}
	refunds, err := persistence.ParseMoney(refundStr)
	if err != nil {
		return valueobjects.Money{}, valueobjects.Money{}, err
	}
	return total, refunds, nil
}

func scanCreditCard(row *sql.Row) (*treasury.CreditCard, error) {
	c := &treasury.CreditCard{}
	var (
		creditLimit, currentBalance string
		createdBy, updatedBy        sql.NullString
		deletedAt                   sql.NullTime
	)
	err := persistence.ScanRow(row,
		&c.ID, &c.Issuer, &c.LastFour, &creditLimit, &currentBalance,
		&c.CutOffDay, &c.PaymentDueDay, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	if err := decodeCreditCard(c, creditLimit, currentBalance, createdBy, updatedBy, deletedAt); err != nil {
		return nil, err
	}
	return c, nil
}

func scanCreditCardFromRows(rows *sql.Rows) (*treasury.CreditCard, error) {
	c := &treasury.CreditCard{}
	var (
		creditLimit, currentBalance string
		createdBy, updatedBy        sql.NullString
		deletedAt                   sql.NullTime
	)
	if err := rows.Scan(
		&c.ID, &c.Issuer, &c.LastFour, &creditLimit, &currentBalance,
		&c.CutOffDay, &c.PaymentDueDay, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodeCreditCard(c, creditLimit, currentBalance, createdBy, updatedBy, deletedAt); err != nil {
		return nil, err
	}
	return c, nil
}

func decodeCreditCard(c *treasury.CreditCard, creditLimit, currentBalance string, createdBy, updatedBy sql.NullString, deletedAt sql.NullTime) error {
	lim, err := persistence.ParseMoney(creditLimit)
	if err != nil {
		return err
	}
	c.CreditLimit = lim
	bal, err := persistence.ParseMoney(currentBalance)
	if err != nil {
		return err
	}
	c.CurrentBalance = bal
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		c.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		c.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		c.DeletedAt = &t
	}
	return nil
}

var _ treasury.CreditCardRepository = (*creditCardRepository)(nil)
