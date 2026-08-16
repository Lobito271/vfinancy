package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/sales"
)

type customerAdvanceRepository struct {
	q persistence.Querier
}

func NewCustomerAdvanceRepository(db *sql.DB) *customerAdvanceRepository {
	return &customerAdvanceRepository{q: persistence.FromDB(db)}
}

const customerAdvanceColumns = `
	id, company_id, customer_id, number, advance_date, amount, remaining,
	payment_method, reference, bank_account_id, currency_code, exchange_rate,
	notes, status, created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *customerAdvanceRepository) Create(ctx context.Context, a *sales.CustomerAdvance) error {
	const q = `INSERT INTO customer_advances (
		id, company_id, customer_id, number, advance_date, amount, remaining,
		payment_method, reference, bank_account_id, currency_code, exchange_rate,
		notes, status, created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
		$16, $17, $18, $19)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		a.ID, a.CompanyID, a.CustomerID, a.Number, a.AdvanceDate,
		a.Amount.String(), a.Remaining().String(), a.Method.String(), nil,
		persistence.NullIfEmptyUUID(a.BankAccountID),
		a.CurrencyCode.String(), a.ExchangeRate.String(),
		persistence.NullIfEmpty(a.Notes), a.Status,
		a.CreatedAt, a.UpdatedAt, nil,
		persistence.NullIfEmptyUUID(a.CreatedBy), persistence.NullIfEmptyUUID(a.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *customerAdvanceRepository) Update(ctx context.Context, a *sales.CustomerAdvance) error {
	const q = `UPDATE customer_advances SET
		amount = $1, remaining = $2, payment_method = $3, bank_account_id = $4,
		notes = $5, status = $6, updated_at = $7, updated_by = $8
	 WHERE id = $9 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		a.Amount.String(), a.Remaining().String(), a.Method.String(),
		persistence.NullIfEmptyUUID(a.BankAccountID),
		persistence.NullIfEmpty(a.Notes), a.Status,
		time.Now().UTC(), persistence.NullIfEmptyUUID(a.UpdatedBy), a.ID,
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

func (r *customerAdvanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*sales.CustomerAdvance, error) {
	q := `SELECT ` + customerAdvanceColumns + ` FROM customer_advances WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanCustomerAdvance(row)
}

func (r *customerAdvanceRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*sales.CustomerAdvance, error) {
	q := `SELECT ` + customerAdvanceColumns + `
		FROM customer_advances
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY advance_date DESC`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, customerID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*sales.CustomerAdvance, 0, 8)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		a, err := scanCustomerAdvanceFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, a)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *customerAdvanceRepository) ListApplicationsForSale(ctx context.Context, saleID uuid.UUID) ([]*sales.CustomerAdvance, error) {
	const q = `SELECT ca.id, ca.company_id, ca.customer_id, ca.number,
		ca.advance_date, ca.amount, ca.remaining, ca.payment_method, ca.reference,
		ca.bank_account_id, ca.currency_code, ca.exchange_rate, ca.notes, ca.status,
		ca.created_at, ca.updated_at, ca.deleted_at, ca.created_by, ca.updated_by
		FROM customer_advances ca
		JOIN customer_advance_applications a ON a.customer_advance_id = ca.id
		WHERE a.sale_id = $1 AND ca.deleted_at IS NULL
		ORDER BY ca.advance_date DESC`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, saleID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*sales.CustomerAdvance, 0, 8)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		a, err := scanCustomerAdvanceFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, a)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func scanCustomerAdvance(row *sql.Row) (*sales.CustomerAdvance, error) {
	a := &sales.CustomerAdvance{}
	var (
		bankAccountID, reference, notes  sql.NullString
		createdBy, updatedBy             sql.NullString
		deletedAt                        sql.NullTime
		amount, remaining, method        string
		currency, rate, status           string
	)
	err := persistence.ScanRow(row,
		&a.ID, &a.CompanyID, &a.CustomerID, &a.Number,
		&a.AdvanceDate, &amount, &remaining, &method, &reference, &bankAccountID,
		&currency, &rate, &notes, &status,
		&a.CreatedAt, &a.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	_ = reference
	_ = remaining
	_ = deletedAt
	if err := decodeCustomerAdvance(a, bankAccountID, notes, createdBy, updatedBy, amount, method, currency, rate, status); err != nil {
		return nil, err
	}
	return a, nil
}

func scanCustomerAdvanceFromRows(rows *sql.Rows) (*sales.CustomerAdvance, error) {
	a := &sales.CustomerAdvance{}
	var (
		bankAccountID, reference, notes  sql.NullString
		createdBy, updatedBy             sql.NullString
		deletedAt                        sql.NullTime
		amount, remaining, method        string
		currency, rate, status           string
	)
	if err := rows.Scan(
		&a.ID, &a.CompanyID, &a.CustomerID, &a.Number,
		&a.AdvanceDate, &amount, &remaining, &method, &reference, &bankAccountID,
		&currency, &rate, &notes, &status,
		&a.CreatedAt, &a.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	_ = reference
	_ = remaining
	_ = deletedAt
	if err := decodeCustomerAdvance(a, bankAccountID, notes, createdBy, updatedBy, amount, method, currency, rate, status); err != nil {
		return nil, err
	}
	return a, nil
}

func decodeCustomerAdvance(a *sales.CustomerAdvance, bankAccountID, notes, createdBy, updatedBy sql.NullString, amount, method, currency, rate, status string) error {
	if bankAccountID.Valid {
		id := persistence.ParseUUID(bankAccountID.String)
		a.BankAccountID = &id
	}
	if notes.Valid {
		a.Notes = notes.String
	}
	if v, err := persistence.ParseMoney(amount); err != nil {
		return err
	} else {
		a.Amount = v
	}
	a.Method = persistence.ParsePaymentMethod(method)
	cc, err := valueobjects.NewCurrencyCode(currency)
	if err != nil {
		return err
	}
	a.CurrencyCode = cc
	er, err := valueobjects.ExchangeRateFromString(rate)
	if err != nil {
		return err
	}
	a.ExchangeRate = er
	a.Status = status
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		a.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		a.UpdatedBy = &id
	}
	return nil
}

var _ sales.CustomerAdvanceRepository = (*customerAdvanceRepository)(nil)
