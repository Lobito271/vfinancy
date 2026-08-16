package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/purchasing"
	"vfinancy/backend/infrastructure/persistence"
)

type supplierPaymentRepository struct {
	q persistence.Querier
}

func NewSupplierPaymentRepository(db *sql.DB) *supplierPaymentRepository {
	return &supplierPaymentRepository{q: persistence.FromDB(db)}
}

const supplierPaymentColumns = `
	id, company_id, supplier_id, branch_id, number, payment_date,
	amount, payment_method, reference, bank_account_id, cash_register_id,
	credit_card_id, currency_code, exchange_rate, notes, status,
	created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *supplierPaymentRepository) Create(ctx context.Context, p *purchasing.SupplierPayment) error {
	const q = `INSERT INTO supplier_payments (
		id, company_id, supplier_id, branch_id, number, payment_date,
		amount, payment_method, reference, bank_account_id, cash_register_id,
		credit_card_id, currency_code, exchange_rate, notes, status,
		created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, NULL, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.ID, p.CompanyID, p.SupplierID, p.Number, p.PaymentDate,
		p.Amount.String(), p.Method.String(), persistence.NullIfEmpty(p.Reference),
		persistence.NullIfEmptyUUID(p.BankAccountID), persistence.NullIfEmptyUUID(p.CashRegisterID),
		persistence.NullIfEmptyUUID(p.CreditCardID), p.CurrencyCode.String(), p.ExchangeRate.String(),
		persistence.NullIfEmpty(p.Notes), p.Status, p.CreatedAt, p.UpdatedAt,
		persistence.NullIfEmptyUUID(p.CreatedBy), persistence.NullIfEmptyUUID(p.UpdatedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	return r.insertAllocations(ctx, p)
}

func (r *supplierPaymentRepository) insertAllocations(ctx context.Context, p *purchasing.SupplierPayment) error {
	const q = `INSERT INTO supplier_payment_allocations (supplier_payment_id, purchase_order_id, allocated_amount)
		VALUES ($1, $2, $3)`
	for _, a := range p.Allocations() {
		if _, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, p.ID, a.PurchaseOrderID, a.Amount.String()); err != nil {
			return persistence.Translate(err)
		}
	}
	return nil
}

func (r *supplierPaymentRepository) Update(ctx context.Context, p *purchasing.SupplierPayment) error {
	const q = `UPDATE supplier_payments SET
		payment_date = $1, amount = $2, payment_method = $3, reference = $4,
		bank_account_id = $5, cash_register_id = $6, credit_card_id = $7,
		notes = $8, status = $9, updated_at = $10, updated_by = $11
	 WHERE id = $12 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.PaymentDate, p.Amount.String(), p.Method.String(),
		persistence.NullIfEmpty(p.Reference),
		persistence.NullIfEmptyUUID(p.BankAccountID),
		persistence.NullIfEmptyUUID(p.CashRegisterID),
		persistence.NullIfEmptyUUID(p.CreditCardID),
		persistence.NullIfEmpty(p.Notes), p.Status, time.Now().UTC(),
		persistence.NullIfEmptyUUID(p.UpdatedBy), p.ID,
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

func (r *supplierPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*purchasing.SupplierPayment, error) {
	q := `SELECT ` + supplierPaymentColumns + ` FROM supplier_payments WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanSupplierPayment(row)
}

func (r *supplierPaymentRepository) List(ctx context.Context, filter purchasing.SupplierPaymentFilter) (repositories.Page[*purchasing.SupplierPayment], error) {
	var (
		clauses = []string{"deleted_at IS NULL"}
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.SupplierID != nil {
		clauses = append(clauses, fmt.Sprintf("supplier_id = $%d", len(args)+1))
		args = append(args, *filter.SupplierID)
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
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM supplier_payments WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*purchasing.SupplierPayment]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM supplier_payments WHERE %s ORDER BY payment_date DESC LIMIT $%d OFFSET $%d",
			supplierPaymentColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*purchasing.SupplierPayment]{}, persistence.Translate(err)
	}
	out := make([]*purchasing.SupplierPayment, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		p, err := scanSupplierPaymentFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return repositories.Page[*purchasing.SupplierPayment]{}, err
	}
	return repositories.Page[*purchasing.SupplierPayment]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *supplierPaymentRepository) ListAllocationsForPurchase(ctx context.Context, purchaseID uuid.UUID) ([]*purchasing.SupplierPayment, error) {
	q := `SELECT ` + supplierPaymentColumns + `
		FROM supplier_payments
		INNER JOIN supplier_payment_allocations a ON a.supplier_payment_id = supplier_payments.id
		WHERE a.purchase_order_id = $1 AND supplier_payments.deleted_at IS NULL`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, purchaseID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.SupplierPayment, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		p, err := scanSupplierPaymentFromRows(r)
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

func (r *supplierPaymentRepository) GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	like := fmt.Sprintf("PAG-%d-%%", year)
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM supplier_payments WHERE company_id = $1 AND number LIKE $2`,
		companyID, like).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("PAG-%d-%05d", year, n+1), nil
}

type supplierPaymentScan struct {
	branchID, reference, notes, createdBy, updatedBy     sql.NullString
	bankAccountID, cashRegisterID, creditCardID          sql.NullString
	deletedAt                                            sql.NullTime
	amount, method, currencyCode, exchangeRate, status   string
}

func scanSupplierPayment(row *sql.Row) (*purchasing.SupplierPayment, error) {
	p := &purchasing.SupplierPayment{}
	var s supplierPaymentScan
	err := persistence.ScanRow(row,
		&p.ID, &p.CompanyID, &p.SupplierID, &s.branchID, &p.Number, &p.PaymentDate,
		&s.amount, &s.method, &s.reference, &s.bankAccountID, &s.cashRegisterID,
		&s.creditCardID, &s.currencyCode, &s.exchangeRate, &s.notes, &s.status,
		&p.CreatedAt, &p.UpdatedAt, &s.deletedAt, &s.createdBy, &s.updatedBy,
	)
	if err != nil {
		return nil, err
	}
	if err := decodeSupplierPayment(p, &s); err != nil {
		return nil, err
	}
	return p, nil
}

func scanSupplierPaymentFromRows(rows *sql.Rows) (*purchasing.SupplierPayment, error) {
	p := &purchasing.SupplierPayment{}
	var s supplierPaymentScan
	if err := rows.Scan(
		&p.ID, &p.CompanyID, &p.SupplierID, &s.branchID, &p.Number, &p.PaymentDate,
		&s.amount, &s.method, &s.reference, &s.bankAccountID, &s.cashRegisterID,
		&s.creditCardID, &s.currencyCode, &s.exchangeRate, &s.notes, &s.status,
		&p.CreatedAt, &p.UpdatedAt, &s.deletedAt, &s.createdBy, &s.updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodeSupplierPayment(p, &s); err != nil {
		return nil, err
	}
	return p, nil
}

func decodeSupplierPayment(p *purchasing.SupplierPayment, s *supplierPaymentScan) error {
	if s.reference.Valid {
		p.Reference = s.reference.String
	}
	if s.bankAccountID.Valid {
		id := persistence.ParseUUID(s.bankAccountID.String)
		p.BankAccountID = &id
	}
	if s.cashRegisterID.Valid {
		id := persistence.ParseUUID(s.cashRegisterID.String)
		p.CashRegisterID = &id
	}
	if s.creditCardID.Valid {
		id := persistence.ParseUUID(s.creditCardID.String)
		p.CreditCardID = &id
	}
	if s.notes.Valid {
		p.Notes = s.notes.String
	}
	if s.createdBy.Valid {
		id := persistence.ParseUUID(s.createdBy.String)
		p.CreatedBy = &id
	}
	if s.updatedBy.Valid {
		id := persistence.ParseUUID(s.updatedBy.String)
		p.UpdatedBy = &id
	}
	p.Method = persistence.ParsePaymentMethod(s.method)
	p.Status = s.status
	var err error
	if p.CurrencyCode, err = valueobjects.NewCurrencyCode(s.currencyCode); err != nil {
		return err
	}
	if p.ExchangeRate, err = valueobjects.ExchangeRateFromString(s.exchangeRate); err != nil {
		return err
	}
	if p.Amount, err = persistence.ParseMoney(s.amount); err != nil {
		return err
	}
	return nil
}

var _ purchasing.SupplierPaymentRepository = (*supplierPaymentRepository)(nil)
