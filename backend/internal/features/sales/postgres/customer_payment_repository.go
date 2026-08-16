package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/sales"
)

type customerPaymentRepository struct {
	q persistence.Querier
}

func NewCustomerPaymentRepository(db *sql.DB) *customerPaymentRepository {
	return &customerPaymentRepository{q: persistence.FromDB(db)}
}

const customerPaymentColumns = `
	id, company_id, customer_id, branch_id, number, payment_date, amount,
	payment_method, reference, bank_account_id, cash_register_id, currency_code,
	exchange_rate, notes, status, created_at, updated_at, deleted_at,
	created_by, updated_by
`

func (r *customerPaymentRepository) Create(ctx context.Context, p *sales.CustomerPayment) error {
	const q = `INSERT INTO customer_payments (
		id, company_id, customer_id, branch_id, number, payment_date, amount,
		payment_method, reference, bank_account_id, cash_register_id, currency_code,
		exchange_rate, notes, status, created_at, updated_at, deleted_at,
		created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
		$16, $17, $18, $19, $20)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.ID, p.CompanyID, p.CustomerID, nil, p.Number, p.PaymentDate,
		p.Amount.String(), p.Method.String(), persistence.NullIfEmpty(p.Reference),
		persistence.NullIfEmptyUUID(p.BankAccountID), persistence.NullIfEmptyUUID(p.CashRegisterID),
		p.CurrencyCode.String(), p.ExchangeRate.String(), persistence.NullIfEmpty(p.Notes),
		p.Status, p.CreatedAt, p.UpdatedAt, nil,
		persistence.NullIfEmptyUUID(p.CreatedBy), persistence.NullIfEmptyUUID(p.UpdatedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	return r.insertAllocations(ctx, p)
}

func (r *customerPaymentRepository) insertAllocations(ctx context.Context, p *sales.CustomerPayment) error {
	const q = `INSERT INTO customer_payment_allocations (customer_payment_id, sale_id, allocated_amount)
		VALUES ($1, $2, $3)`
	for _, a := range p.Allocations() {
		if _, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, p.ID, a.SaleID, a.Amount.String()); err != nil {
			return persistence.Translate(err)
		}
	}
	return nil
}

func (r *customerPaymentRepository) Update(ctx context.Context, p *sales.CustomerPayment) error {
	const q = `UPDATE customer_payments SET
		payment_date = $1, amount = $2, payment_method = $3, reference = $4,
		bank_account_id = $5, cash_register_id = $6, notes = $7, status = $8,
		updated_at = $9, updated_by = $10
	 WHERE id = $11 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.PaymentDate, p.Amount.String(), p.Method.String(),
		persistence.NullIfEmpty(p.Reference),
		persistence.NullIfEmptyUUID(p.BankAccountID), persistence.NullIfEmptyUUID(p.CashRegisterID),
		persistence.NullIfEmpty(p.Notes), p.Status,
		time.Now().UTC(), persistence.NullIfEmptyUUID(p.UpdatedBy), p.ID,
	)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return r.insertAllocations(ctx, p)
}

func (r *customerPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*sales.CustomerPayment, error) {
	q := `SELECT ` + customerPaymentColumns + ` FROM customer_payments WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanCustomerPayment(row)
}

func (r *customerPaymentRepository) List(ctx context.Context, filter sales.CustomerPaymentFilter) (repositories.Page[*sales.CustomerPayment], error) {
	var (
		clauses = []string{"deleted_at IS NULL"}
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.CustomerID != nil {
		clauses = append(clauses, fmt.Sprintf("customer_id = $%d", len(args)+1))
		args = append(args, *filter.CustomerID)
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if !filter.PayRange.From.IsZero() {
		clauses = append(clauses, fmt.Sprintf("payment_date >= $%d", len(args)+1))
		args = append(args, filter.PayRange.From)
	}
	if !filter.PayRange.To.IsZero() {
		clauses = append(clauses, fmt.Sprintf("payment_date <= $%d", len(args)+1))
		args = append(args, filter.PayRange.To)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := persistence.JoinClauses(clauses)
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM customer_payments WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*sales.CustomerPayment]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM customer_payments WHERE %s ORDER BY payment_date DESC LIMIT $%d OFFSET $%d",
			customerPaymentColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*sales.CustomerPayment]{}, persistence.Translate(err)
	}
	out := make([]*sales.CustomerPayment, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		p, err := scanCustomerPaymentFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return repositories.Page[*sales.CustomerPayment]{}, err
	}
	return repositories.Page[*sales.CustomerPayment]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *customerPaymentRepository) ListAllocationsForSale(ctx context.Context, saleID uuid.UUID) ([]*sales.CustomerPayment, error) {
	const q = `SELECT cp.id, cp.company_id, cp.customer_id, cp.branch_id, cp.number,
		cp.payment_date, cp.amount, cp.payment_method, cp.reference, cp.bank_account_id,
		cp.cash_register_id, cp.currency_code, cp.exchange_rate, cp.notes, cp.status,
		cp.created_at, cp.updated_at, cp.deleted_at, cp.created_by, cp.updated_by
		FROM customer_payments cp
		JOIN customer_payment_allocations a ON a.customer_payment_id = cp.id
		WHERE a.sale_id = $1 AND cp.deleted_at IS NULL
		ORDER BY cp.payment_date DESC`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, saleID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*sales.CustomerPayment, 0, 8)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		p, err := scanCustomerPaymentFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *customerPaymentRepository) GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	like := fmt.Sprintf("COB-%d-%%", year)
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM customer_payments WHERE company_id = $1 AND number LIKE $2`,
		companyID, like).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("COB-%d-%05d", year, n+1), nil
}

func scanCustomerPayment(row *sql.Row) (*sales.CustomerPayment, error) {
	p := &sales.CustomerPayment{}
	var (
		branchID, bankAccountID, cashRegisterID, reference, notes sql.NullString
		createdBy, updatedBy                                      sql.NullString
		deletedAt                                                 sql.NullTime
		amount, method, currency, rate, status                    string
	)
	err := persistence.ScanRow(row,
		&p.ID, &p.CompanyID, &p.CustomerID, &branchID, &p.Number,
		&p.PaymentDate, &amount, &method, &reference, &bankAccountID, &cashRegisterID,
		&currency, &rate, &notes, &status,
		&p.CreatedAt, &p.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	_ = branchID
	_ = deletedAt
	if err := decodeCustomerPayment(p, bankAccountID, cashRegisterID, reference, notes, createdBy, updatedBy, amount, method, currency, rate, status); err != nil {
		return nil, err
	}
	return p, nil
}

func scanCustomerPaymentFromRows(rows *sql.Rows) (*sales.CustomerPayment, error) {
	p := &sales.CustomerPayment{}
	var (
		branchID, bankAccountID, cashRegisterID, reference, notes sql.NullString
		createdBy, updatedBy                                      sql.NullString
		deletedAt                                                 sql.NullTime
		amount, method, currency, rate, status                    string
	)
	if err := rows.Scan(
		&p.ID, &p.CompanyID, &p.CustomerID, &branchID, &p.Number,
		&p.PaymentDate, &amount, &method, &reference, &bankAccountID, &cashRegisterID,
		&currency, &rate, &notes, &status,
		&p.CreatedAt, &p.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	_ = branchID
	_ = deletedAt
	if err := decodeCustomerPayment(p, bankAccountID, cashRegisterID, reference, notes, createdBy, updatedBy, amount, method, currency, rate, status); err != nil {
		return nil, err
	}
	return p, nil
}

func decodeCustomerPayment(p *sales.CustomerPayment, bankAccountID, cashRegisterID, reference, notes, createdBy, updatedBy sql.NullString, amount, method, currency, rate, status string) error {
	if bankAccountID.Valid {
		id := persistence.ParseUUID(bankAccountID.String)
		p.BankAccountID = &id
	}
	if cashRegisterID.Valid {
		id := persistence.ParseUUID(cashRegisterID.String)
		p.CashRegisterID = &id
	}
	if reference.Valid {
		p.Reference = reference.String
	}
	if notes.Valid {
		p.Notes = notes.String
	}
	if v, err := persistence.ParseMoney(amount); err != nil {
		return err
	} else {
		p.Amount = v
	}
	p.Method = persistence.ParsePaymentMethod(method)
	cc, err := valueobjects.NewCurrencyCode(currency)
	if err != nil {
		return err
	}
	p.CurrencyCode = cc
	er, err := valueobjects.ExchangeRateFromString(rate)
	if err != nil {
		return err
	}
	p.ExchangeRate = er
	p.Status = status
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		p.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		p.UpdatedBy = &id
	}
	return nil
}

var _ sales.CustomerPaymentRepository = (*customerPaymentRepository)(nil)
