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

func NewCreditCardRepository(db *sql.DB) *creditCardRepository {
	return &creditCardRepository{q: persistence.FromDB(db)}
}

const creditCardColumns = `
	id, company_id, branch_id, issuer, last_four, card_holder,
	expiration_month, expiration_year, credit_limit, current_balance,
	cut_off_day, payment_due_day, currency_code, gl_account_id, is_active,
	created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *creditCardRepository) Create(ctx context.Context, c *treasury.CreditCard) error {
	const q = `INSERT INTO credit_cards (
		id, company_id, branch_id, issuer, last_four, card_holder,
		expiration_month, expiration_year, credit_limit, current_balance,
		cut_off_day, payment_due_day, currency_code, gl_account_id, is_active,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		c.ID, c.CompanyID, persistence.NullIfEmptyUUID(c.BranchID),
		c.Issuer, c.LastFour, c.CardHolder,
		c.ExpirationMonth, c.ExpirationYear, c.CreditLimit.String(), c.CurrentBalance.String(),
		c.CutOffDay, c.PaymentDueDay, c.CurrencyCode.String(), c.GLAccountID, c.IsActive,
		c.CreatedAt, c.UpdatedAt, persistence.NullIfZeroTime(c.DeletedAt),
		persistence.NullIfEmptyUUID(c.CreatedBy), persistence.NullIfEmptyUUID(c.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *creditCardRepository) Update(ctx context.Context, c *treasury.CreditCard) error {
	const q = `UPDATE credit_cards SET
		issuer = $1, last_four = $2, card_holder = $3,
		expiration_month = $4, expiration_year = $5, credit_limit = $6,
		current_balance = $7, cut_off_day = $8, payment_due_day = $9,
		currency_code = $10, gl_account_id = $11, is_active = $12,
		branch_id = $13, updated_at = $14, updated_by = $15
	 WHERE id = $16 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		c.Issuer, c.LastFour, c.CardHolder,
		c.ExpirationMonth, c.ExpirationYear, c.CreditLimit.String(),
		c.CurrentBalance.String(), c.CutOffDay, c.PaymentDueDay,
		c.CurrencyCode.String(), c.GLAccountID, c.IsActive,
		persistence.NullIfEmptyUUID(c.BranchID), time.Now().UTC(), persistence.NullIfEmptyUUID(c.UpdatedBy),
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

func (r *creditCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE credit_cards SET deleted_at = $1, updated_at = $2, is_active = FALSE WHERE id = $3 AND deleted_at IS NULL`
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

func (r *creditCardRepository) GetByID(ctx context.Context, id uuid.UUID) (*treasury.CreditCard, error) {
	q := `SELECT ` + creditCardColumns + ` FROM credit_cards WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanCreditCard(row)
}

func (r *creditCardRepository) List(ctx context.Context, companyID uuid.UUID) ([]*treasury.CreditCard, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT `+creditCardColumns+` FROM credit_cards WHERE company_id = $1 AND deleted_at IS NULL AND is_active = TRUE ORDER BY issuer`,
		companyID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*treasury.CreditCard, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		c, err := scanCreditCardFromRows(r)
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

func scanCreditCard(row *sql.Row) (*treasury.CreditCard, error) {
	c := &treasury.CreditCard{}
	var (
		branchID, createdBy, updatedBy            sql.NullString
		deletedAt                                 sql.NullTime
		currencyCode, creditLimit, currentBalance string
	)
	err := persistence.ScanRow(row,
		&c.ID, &c.CompanyID, &branchID, &c.Issuer, &c.LastFour, &c.CardHolder,
		&c.ExpirationMonth, &c.ExpirationYear, &creditLimit, &currentBalance,
		&c.CutOffDay, &c.PaymentDueDay, &currencyCode, &c.GLAccountID, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	if err := decodeCreditCard(c, branchID, createdBy, updatedBy, deletedAt, currencyCode, creditLimit, currentBalance); err != nil {
		return nil, err
	}
	return c, nil
}

func scanCreditCardFromRows(rows *sql.Rows) (*treasury.CreditCard, error) {
	c := &treasury.CreditCard{}
	var (
		branchID, createdBy, updatedBy            sql.NullString
		deletedAt                                 sql.NullTime
		currencyCode, creditLimit, currentBalance string
	)
	if err := rows.Scan(
		&c.ID, &c.CompanyID, &branchID, &c.Issuer, &c.LastFour, &c.CardHolder,
		&c.ExpirationMonth, &c.ExpirationYear, &creditLimit, &currentBalance,
		&c.CutOffDay, &c.PaymentDueDay, &currencyCode, &c.GLAccountID, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodeCreditCard(c, branchID, createdBy, updatedBy, deletedAt, currencyCode, creditLimit, currentBalance); err != nil {
		return nil, err
	}
	return c, nil
}

func decodeCreditCard(c *treasury.CreditCard, branchID, createdBy, updatedBy sql.NullString, deletedAt sql.NullTime, currencyCode, creditLimit, currentBalance string) error {
	if branchID.Valid {
		id := persistence.ParseUUID(branchID.String)
		c.BranchID = &id
	}
	cc, err := valueobjects.NewCurrencyCode(currencyCode)
	if err != nil {
		return err
	}
	c.CurrencyCode = cc
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
