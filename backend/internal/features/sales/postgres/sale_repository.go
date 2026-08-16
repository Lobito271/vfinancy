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

type saleRepository struct {
	q persistence.Querier
}

func NewSaleRepository(db *sql.DB) *saleRepository {
	return &saleRepository{q: persistence.FromDB(db)}
}

const saleColumns = `
	id, company_id, branch_id, customer_id, number, sale_date, due_date, status,
	subtotal, tax_amount, total, paid_amount, discount_amount, cost_total, profit,
	currency_code, exchange_rate, notes, cancelled_at, cancelled_reason,
	created_at, updated_at, deleted_at, created_by, updated_by
`

const saleItemColumns = `
	id, sale_id, product_id, inventory_batch_id, line_number, quantity,
	unit_price, discount_percent, discount_amount, tax_rate, tax_amount,
	line_total, cost_snapshot, description, created_at
`

const insertSaleItem = `INSERT INTO sale_items (
	id, sale_id, product_id, inventory_batch_id, line_number, quantity,
	unit_price, discount_percent, discount_amount, tax_rate, tax_amount,
	line_total, cost_snapshot, description, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

func (r *saleRepository) Create(ctx context.Context, s *sales.Sale) error {
	const q = `INSERT INTO sales (
		id, company_id, branch_id, customer_id, number, sale_date, due_date, status,
		subtotal, tax_amount, total, paid_amount, discount_amount, cost_total, profit,
		currency_code, exchange_rate, notes, cancelled_at, cancelled_reason,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, COALESCE($3, (SELECT id FROM branches WHERE company_id = $2 AND is_default = 1 AND deleted_at IS NULL LIMIT 1)), $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
		$16, $17, $18, $19, $20, $21, $22, $23, $24, $25)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.ID, s.CompanyID, s.BranchID, s.CustomerID, s.Number, s.CreatedAt,
		persistence.NullIfZeroTime(s.DueDate), s.Status.String(),
		s.Subtotal.String(), s.TaxAmount.String(), s.Total.String(), s.Paid.String(),
		s.DiscountAmount.String(), s.CostTotal.String(), s.Profit.String(),
		s.CurrencyCode.String(), s.ExchangeRate.String(),
		persistence.NullIfEmpty(s.Notes),
		persistence.NullIfZeroTime(s.CancelledAt), persistence.NullIfEmpty(s.CancelledReason),
		s.CreatedAt, s.UpdatedAt, nil,
		persistence.NullIfEmptyUUID(s.CreatedBy), persistence.NullIfEmptyUUID(s.UpdatedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	for _, li := range s.Items {
		_, err := persistence.Q(ctx, r.q).ExecContext(ctx, insertSaleItem,
			li.ID, li.SaleID, li.ProductID, nil, li.LineNumber,
			li.Quantity.String(), li.UnitPrice.String(),
			li.DiscountPercent.String(), li.DiscountAmount.String(),
			li.TaxRate.String(), li.TaxAmount.String(),
			li.LineTotal().String(), li.CostSnapshot.String(),
			persistence.NullIfEmpty(li.Description), li.CreatedAt,
		)
		if err != nil {
			return persistence.Translate(err)
		}
	}
	return nil
}

func (r *saleRepository) Update(ctx context.Context, s *sales.Sale) error {
	const q = `UPDATE sales SET
		branch_id = COALESCE($1, branch_id), customer_id = $2, due_date = $3, status = $4,
		subtotal = $5, tax_amount = $6, total = $7, paid_amount = $8,
		discount_amount = $9, cost_total = $10, profit = $11, notes = $12,
		cancelled_at = $13, cancelled_reason = $14, updated_at = $15, updated_by = $16
	 WHERE id = $17 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.BranchID, s.CustomerID, persistence.NullIfZeroTime(s.DueDate), s.Status.String(),
		s.Subtotal.String(), s.TaxAmount.String(), s.Total.String(), s.Paid.String(),
		s.DiscountAmount.String(), s.CostTotal.String(), s.Profit.String(),
		persistence.NullIfEmpty(s.Notes),
		persistence.NullIfZeroTime(s.CancelledAt), persistence.NullIfEmpty(s.CancelledReason),
		time.Now().UTC(), persistence.NullIfEmptyUUID(s.UpdatedBy), s.ID,
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

func (r *saleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE sales SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
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

func (r *saleRepository) GetByID(ctx context.Context, id uuid.UUID) (*sales.Sale, error) {
	q := `SELECT ` + saleColumns + ` FROM sales WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	s, err := scanSale(row)
	if err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *saleRepository) GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*sales.Sale, error) {
	q := `SELECT ` + saleColumns + `
		FROM sales
		WHERE company_id = $1 AND number = $2 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, number)
	s, err := scanSale(row)
	if err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *saleRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, `SELECT 1 FROM sales WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if persistence.IsPgNoRows(err) {
			return false, nil
		}
		return false, persistence.Translate(err)
	}
	return n == 1, nil
}

func (r *saleRepository) List(ctx context.Context, filter sales.SaleFilter) (repositories.Page[*sales.Sale], error) {
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
	if filter.BranchID != nil {
		clauses = append(clauses, fmt.Sprintf("branch_id = $%d", len(args)+1))
		args = append(args, *filter.BranchID)
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if !filter.IssueRange.From.IsZero() {
		clauses = append(clauses, fmt.Sprintf("sale_date >= $%d", len(args)+1))
		args = append(args, filter.IssueRange.From)
	}
	if !filter.IssueRange.To.IsZero() {
		clauses = append(clauses, fmt.Sprintf("sale_date <= $%d", len(args)+1))
		args = append(args, filter.IssueRange.To)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := persistence.JoinClauses(clauses)
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM sales WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*sales.Sale]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM sales WHERE %s ORDER BY sale_date DESC, number DESC LIMIT $%d OFFSET $%d",
			saleColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*sales.Sale]{}, persistence.Translate(err)
	}
	out := make([]*sales.Sale, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		s, err := scanSaleFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, s)
		return nil
	}); err != nil {
		return repositories.Page[*sales.Sale]{}, err
	}
	return repositories.Page[*sales.Sale]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *saleRepository) GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	like := fmt.Sprintf("V-%d-%%", year)
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM sales WHERE company_id = $1 AND number LIKE $2`,
		companyID, like).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("V-%d-%05d", year, n+1), nil
}

func (r *saleRepository) loadItems(ctx context.Context, s *sales.Sale) error {
	q := `SELECT ` + saleItemColumns + ` FROM sale_items WHERE sale_id = $1 ORDER BY line_number`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, s.ID)
	if err != nil {
		return persistence.Translate(err)
	}
	items := make([]*sales.SaleItem, 0, 8)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		li, err := scanSaleItem(r)
		if err != nil {
			return err
		}
		items = append(items, li)
		return nil
	}); err != nil {
		return err
	}
	s.Items = items
	return nil
}

func scanSale(row *sql.Row) (*sales.Sale, error) {
	s := &sales.Sale{}
	var (
		branchID, createdBy, updatedBy                sql.NullString
		dueDate, cancelledAt, deletedAt               sql.NullTime
		notes, reason                                 sql.NullString
		saleDate                                      time.Time
		status, currency, rate                        string
		subtotal, taxAmount, total, paid, discount, cost, profit string
	)
	err := persistence.ScanRow(row,
		&s.ID, &s.CompanyID, &branchID, &s.CustomerID, &s.Number, &saleDate, &dueDate, &status,
		&subtotal, &taxAmount, &total, &paid, &discount, &cost, &profit,
		&currency, &rate, &notes, &cancelledAt, &reason,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	_ = saleDate
	_ = deletedAt
	if err := decodeSale(s, branchID, createdBy, updatedBy, dueDate, cancelledAt, notes, reason, status, currency, rate, subtotal, taxAmount, total, paid, discount, cost, profit); err != nil {
		return nil, err
	}
	return s, nil
}

func scanSaleFromRows(rows *sql.Rows) (*sales.Sale, error) {
	s := &sales.Sale{}
	var (
		branchID, createdBy, updatedBy                sql.NullString
		dueDate, cancelledAt, deletedAt               sql.NullTime
		notes, reason                                 sql.NullString
		saleDate                                      time.Time
		status, currency, rate                        string
		subtotal, taxAmount, total, paid, discount, cost, profit string
	)
	if err := rows.Scan(
		&s.ID, &s.CompanyID, &branchID, &s.CustomerID, &s.Number, &saleDate, &dueDate, &status,
		&subtotal, &taxAmount, &total, &paid, &discount, &cost, &profit,
		&currency, &rate, &notes, &cancelledAt, &reason,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	_ = saleDate
	_ = deletedAt
	if err := decodeSale(s, branchID, createdBy, updatedBy, dueDate, cancelledAt, notes, reason, status, currency, rate, subtotal, taxAmount, total, paid, discount, cost, profit); err != nil {
		return nil, err
	}
	return s, nil
}

func decodeSale(s *sales.Sale, branchID, createdBy, updatedBy sql.NullString, dueDate, cancelledAt sql.NullTime, notes, reason sql.NullString, status, currency, rate string, subtotal, taxAmount, total, paid, discount, cost, profit string) error {
	if branchID.Valid {
		id := persistence.ParseUUID(branchID.String)
		s.BranchID = &id
	}
	cc, err := valueobjects.NewCurrencyCode(currency)
	if err != nil {
		return err
	}
	s.CurrencyCode = cc
	er, err := valueobjects.ExchangeRateFromString(rate)
	if err != nil {
		return err
	}
	s.ExchangeRate = er
	if dueDate.Valid {
		s.DueDate = &dueDate.Time
	}
	s.Status = persistence.ParseSaleStatus(status)
	if v, err := persistence.ParseMoney(subtotal); err != nil {
		return err
	} else {
		s.Subtotal = v
	}
	if v, err := persistence.ParseMoney(taxAmount); err != nil {
		return err
	} else {
		s.TaxAmount = v
	}
	if v, err := persistence.ParseMoney(total); err != nil {
		return err
	} else {
		s.Total = v
	}
	if v, err := persistence.ParseMoney(paid); err != nil {
		return err
	} else {
		s.Paid = v
	}
	if v, err := persistence.ParseMoney(discount); err != nil {
		return err
	} else {
		s.DiscountAmount = v
	}
	if v, err := persistence.ParseMoney(cost); err != nil {
		return err
	} else {
		s.CostTotal = v
	}
	if v, err := persistence.ParseMoney(profit); err != nil {
		return err
	} else {
		s.Profit = v
	}
	if notes.Valid {
		s.Notes = notes.String
	}
	if cancelledAt.Valid {
		t := cancelledAt.Time
		s.CancelledAt = &t
	}
	if reason.Valid {
		s.CancelledReason = reason.String
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		s.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		s.UpdatedBy = &id
	}
	return nil
}

func scanSaleItem(rows *sql.Rows) (*sales.SaleItem, error) {
	li := &sales.SaleItem{}
	var (
		batchID, description               sql.NullString
		lineTotal                          string
		quantity, unitPrice                string
		discountPercent, discountAmount    string
		taxRate, taxAmount, costSnapshot   string
	)
	if err := rows.Scan(
		&li.ID, &li.SaleID, &li.ProductID, &batchID, &li.LineNumber,
		&quantity, &unitPrice, &discountPercent, &discountAmount,
		&taxRate, &taxAmount, &lineTotal, &costSnapshot, &description, &li.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	_ = batchID
	_ = lineTotal
	q, err := valueobjects.QuantityFromString(quantity)
	if err != nil {
		return nil, err
	}
	li.Quantity = q
	if v, err := persistence.ParseMoney(unitPrice); err != nil {
		return nil, err
	} else {
		li.UnitPrice = v
	}
	dp, err := valueobjects.PercentageFromString(discountPercent)
	if err != nil {
		return nil, err
	}
	li.DiscountPercent = dp
	if v, err := persistence.ParseMoney(discountAmount); err != nil {
		return nil, err
	} else {
		li.DiscountAmount = v
	}
	tr, err := valueobjects.PercentageFromString(taxRate)
	if err != nil {
		return nil, err
	}
	li.TaxRate = tr
	if v, err := persistence.ParseMoney(taxAmount); err != nil {
		return nil, err
	} else {
		li.TaxAmount = v
	}
	if v, err := persistence.ParseMoney(costSnapshot); err != nil {
		return nil, err
	} else {
		li.CostSnapshot = v
	}
	if description.Valid {
		li.Description = description.String
	}
	return li, nil
}

var _ sales.SalesRepository = (*saleRepository)(nil)
