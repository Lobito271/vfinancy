package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/sales"
)

type customerPaymentRepository struct {
	q persistence.Querier
}

// NewCustomerPaymentRepository returns a CustomerPaymentRepository
// backed by *sql.DB.
func NewCustomerPaymentRepository(db *sql.DB) *customerPaymentRepository {
	return &customerPaymentRepository{q: persistence.FromDB(db)}
}

const customerPaymentColumns = `
	id, customer_id, number, payment_date, amount, payment_method,
	reference, notes, status, created_at, updated_at, deleted_at, created_by, updated_by
`

const customerPaymentColumnsPrefixed = `
	cp.id, cp.customer_id, cp.number, cp.payment_date, cp.amount, cp.payment_method,
	cp.reference, cp.notes, cp.status, cp.created_at, cp.updated_at, cp.deleted_at,
	cp.created_by, cp.updated_by
`

func (r *customerPaymentRepository) Create(ctx context.Context, p *sales.CustomerPayment, allocations []sales.PaymentAllocation) error {
	const q = `INSERT INTO customer_payments (
		id, customer_id, number, payment_date, amount, payment_method,
		reference, notes, status, created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.ID, p.CustomerID, p.Number, p.PaymentDate, p.Amount.String(), p.PaymentMethod.String(),
		persistence.NullIfEmpty(p.Reference), persistence.NullIfEmpty(p.Notes), p.Status,
		p.CreatedAt, p.UpdatedAt, persistence.NullIfZeroTime(p.DeletedAt),
		persistence.NullIfEmptyUUID(p.CreatedBy), persistence.NullIfEmptyUUID(p.UpdatedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	return r.insertAllocations(ctx, allocations)
}

func (r *customerPaymentRepository) insertAllocations(ctx context.Context, allocations []sales.PaymentAllocation) error {
	const q = `INSERT INTO customer_payment_allocations (id, customer_payment_id, sale_id, allocated_amount, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	for _, a := range allocations {
		if _, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
			a.ID, a.CustomerPaymentID, a.SaleID, a.AllocatedAmount.String(), a.CreatedAt); err != nil {
			return persistence.Translate(err)
		}
	}
	return nil
}

func (r *customerPaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	const q = `UPDATE customer_payments SET status = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, status, time.Now().UTC(), id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *customerPaymentRepository) NextNumber(ctx context.Context) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	like := fmt.Sprintf("CP-%d-%%", year)
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM customer_payments WHERE number LIKE $1`, like).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("CP-%d-%05d", year, n+1), nil
}

func (r *customerPaymentRepository) List(ctx context.Context, filter sales.CustomerPaymentFilter) (repositories.Page[*sales.CustomerPayment], error) {
	var clauses []string
	var args []any
	clauses = append(clauses, "deleted_at IS NULL")
	if filter.CustomerID != nil {
		clauses = append(clauses, fmt.Sprintf("customer_id = $%d", len(args)+1))
		args = append(args, *filter.CustomerID)
	}
	if filter.SaleID != nil {
		clauses = append(clauses, fmt.Sprintf(`EXISTS (SELECT 1 FROM customer_payment_allocations a WHERE a.customer_payment_id = customer_payments.id AND a.sale_id = $%d)`, len(args)+1))
		args = append(args, *filter.SaleID)
	}
	if filter.From != nil {
		clauses = append(clauses, fmt.Sprintf("payment_date >= $%d", len(args)+1))
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		clauses = append(clauses, fmt.Sprintf("payment_date <= $%d", len(args)+1))
		args = append(args, *filter.To)
	}
	where := persistence.JoinClauses(clauses)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM customer_payments WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*sales.CustomerPayment]{}, persistence.Translate(err)
	}

	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM customer_payments WHERE %s ORDER BY payment_date DESC, number DESC LIMIT $%d OFFSET $%d",
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

func (r *customerPaymentRepository) ListForSale(ctx context.Context, saleID uuid.UUID) ([]*sales.CustomerPayment, error) {
	q := `SELECT ` + customerPaymentColumnsPrefixed + `
		FROM customer_payments cp
		JOIN customer_payment_allocations a ON a.customer_payment_id = cp.id
		WHERE a.sale_id = $1 AND cp.deleted_at IS NULL
		ORDER BY cp.payment_date DESC, cp.number DESC`
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

func scanCustomerPayment(row *sql.Row) (*sales.CustomerPayment, error) {
	p := &sales.CustomerPayment{}
	var (
		reference, notes, createdBy, updatedBy sql.NullString
		deletedAt                              sql.NullTime
		paymentDate                            time.Time
		amount, method, status                 string
	)
	if err := persistence.ScanRow(row,
		&p.ID, &p.CustomerID, &p.Number, &paymentDate, &amount, &method,
		&reference, &notes, &status,
		&p.CreatedAt, &p.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, err
	}
	p.PaymentDate = paymentDate
	return p, decodeCustomerPayment(p, reference, notes, createdBy, updatedBy, deletedAt, amount, method, status)
}

func scanCustomerPaymentFromRows(rows *sql.Rows) (*sales.CustomerPayment, error) {
	p := &sales.CustomerPayment{}
	var (
		reference, notes, createdBy, updatedBy sql.NullString
		deletedAt                              sql.NullTime
		paymentDate                            time.Time
		amount, method, status                 string
	)
	if err := rows.Scan(
		&p.ID, &p.CustomerID, &p.Number, &paymentDate, &amount, &method,
		&reference, &notes, &status,
		&p.CreatedAt, &p.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	p.PaymentDate = paymentDate
	return p, decodeCustomerPayment(p, reference, notes, createdBy, updatedBy, deletedAt, amount, method, status)
}

func decodeCustomerPayment(p *sales.CustomerPayment, reference, notes, createdBy, updatedBy sql.NullString, deletedAt sql.NullTime, amount, method, status string) error {
	if reference.Valid {
		p.Reference = reference.String
	}
	if notes.Valid {
		p.Notes = notes.String
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		p.DeletedAt = &t
	}
	var err error
	if p.Amount, err = persistence.ParseMoney(amount); err != nil {
		return err
	}
	p.PaymentMethod = persistence.ParsePaymentMethod(method)
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
