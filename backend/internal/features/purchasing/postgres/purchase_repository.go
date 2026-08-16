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

type purchaseRepository struct {
	q persistence.Querier
}

func NewPurchaseRepository(db *sql.DB) *purchaseRepository {
	return &purchaseRepository{q: persistence.FromDB(db)}
}

const purchaseColumns = `
	id, company_id, supplier_id, branch_id, number, order_date,
	expected_date, received_date, status, subtotal, tax_amount, total,
	paid_amount, discount_amount, currency_code, exchange_rate, notes,
	cancelled_at, cancelled_reason, created_at, updated_at, deleted_at,
	created_by, updated_by
`

const purchaseItemColumns = `
	id, purchase_order_id, product_id, line_number, quantity_ordered,
	quantity_received, unit_cost, discount_percent, discount_amount,
	tax_rate, tax_amount, line_total, description, created_at
`

func (r *purchaseRepository) Create(ctx context.Context, p *purchasing.PurchaseOrder) error {
	const q = `INSERT INTO purchase_orders (
		id, company_id, supplier_id, branch_id, number, order_date,
		expected_date, received_date, status, subtotal, tax_amount, total,
		paid_amount, discount_amount, currency_code, exchange_rate, notes,
		cancelled_at, cancelled_reason, created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.ID, p.CompanyID, p.SupplierID, persistence.NullIfEmptyUUID(p.BranchID), p.Number, p.OrderDate,
		persistence.NullIfZeroTime(p.ExpectedDate), persistence.NullIfZeroTime(p.ReceivedDate),
		p.Status.String(), p.Subtotal.String(), p.TaxAmount.String(), p.Total.String(),
		p.Paid.String(), p.DiscountAmount.String(), p.CurrencyCode.String(), p.ExchangeRate.String(),
		persistence.NullIfEmpty(p.Notes), persistence.NullIfZeroTime(p.CancelledAt),
		persistence.NullIfEmpty(p.CancelledReason), p.CreatedAt, p.UpdatedAt,
		persistence.NullIfEmptyUUID(p.CreatedBy), persistence.NullIfEmptyUUID(p.UpdatedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	for _, li := range p.Items {
		if err := r.insertItem(ctx, p.ID, li); err != nil {
			return err
		}
	}
	return nil
}

func (r *purchaseRepository) insertItem(ctx context.Context, purchaseID uuid.UUID, li *purchasing.PurchaseOrderItem) error {
	const q = `INSERT INTO purchase_order_items (
		id, purchase_order_id, product_id, line_number, quantity_ordered,
		quantity_received, unit_cost, discount_percent, discount_amount,
		tax_rate, tax_amount, line_total, description, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	createdAt := li.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		li.ID, purchaseID, li.ProductID, li.LineNumber, li.Quantity.String(),
		"0.0000", li.UnitPrice.String(), li.DiscountPercent.String(),
		li.DiscountAmount.String(), li.TaxRate.String(), li.TaxAmount.String(),
		li.LineTotal().String(), persistence.NullIfEmpty(li.Description), createdAt,
	)
	return persistence.Translate(err)
}

func (r *purchaseRepository) Update(ctx context.Context, p *purchasing.PurchaseOrder) error {
	const q = `UPDATE purchase_orders SET
		branch_id = $1, expected_date = $2, received_date = $3, status = $4,
		subtotal = $5, tax_amount = $6, total = $7, paid_amount = $8,
		discount_amount = $9, notes = $10, cancelled_at = $11,
		cancelled_reason = $12, updated_at = $13, updated_by = $14
	 WHERE id = $15 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		persistence.NullIfEmptyUUID(p.BranchID), persistence.NullIfZeroTime(p.ExpectedDate),
		persistence.NullIfZeroTime(p.ReceivedDate), p.Status.String(),
		p.Subtotal.String(), p.TaxAmount.String(), p.Total.String(), p.Paid.String(),
		p.DiscountAmount.String(), persistence.NullIfEmpty(p.Notes),
		persistence.NullIfZeroTime(p.CancelledAt), persistence.NullIfEmpty(p.CancelledReason),
		time.Now().UTC(), persistence.NullIfEmptyUUID(p.UpdatedBy), p.ID,
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

func (r *purchaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`UPDATE purchase_orders SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`,
		now, now, id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *purchaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*purchasing.PurchaseOrder, error) {
	q := `SELECT ` + purchaseColumns + ` FROM purchase_orders WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	p, err := scanPurchaseOrder(row)
	if err != nil {
		return nil, err
	}
	items, err := r.scanItems(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Items = items
	return p, nil
}

func (r *purchaseRepository) GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*purchasing.PurchaseOrder, error) {
	q := `SELECT ` + purchaseColumns + `
		FROM purchase_orders
		WHERE company_id = $1 AND number = $2 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, number)
	p, err := scanPurchaseOrder(row)
	if err != nil {
		return nil, err
	}
	items, err := r.scanItems(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Items = items
	return p, nil
}

func (r *purchaseRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT 1 FROM purchase_orders WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if persistence.IsPgNoRows(err) {
			return false, nil
		}
		return false, persistence.Translate(err)
	}
	return n == 1, nil
}

func (r *purchaseRepository) List(ctx context.Context, filter purchasing.PurchaseFilter) (repositories.Page[*purchasing.PurchaseOrder], error) {
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
	if filter.BranchID != nil {
		clauses = append(clauses, fmt.Sprintf("branch_id = $%d", len(args)+1))
		args = append(args, *filter.BranchID)
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if !filter.IssueRange.From.IsZero() {
		clauses = append(clauses, fmt.Sprintf("order_date >= $%d", len(args)+1))
		args = append(args, filter.IssueRange.From)
	}
	if !filter.IssueRange.To.IsZero() {
		clauses = append(clauses, fmt.Sprintf("order_date <= $%d", len(args)+1))
		args = append(args, filter.IssueRange.To)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := persistence.JoinClauses(clauses)
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM purchase_orders WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*purchasing.PurchaseOrder]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM purchase_orders WHERE %s ORDER BY order_date DESC, number DESC LIMIT $%d OFFSET $%d",
			purchaseColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*purchasing.PurchaseOrder]{}, persistence.Translate(err)
	}
	out := make([]*purchasing.PurchaseOrder, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		p, err := scanPurchaseOrderFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return repositories.Page[*purchasing.PurchaseOrder]{}, err
	}
	return repositories.Page[*purchasing.PurchaseOrder]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *purchaseRepository) GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM purchase_orders WHERE company_id = $1 AND number LIKE $2`,
		companyID, fmt.Sprintf("PO-%d-%%", year)).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("PO-%d-%05d", year, n+1), nil
}

func (r *purchaseRepository) scanItems(ctx context.Context, purchaseID uuid.UUID) ([]*purchasing.PurchaseOrderItem, error) {
	q := `SELECT ` + purchaseItemColumns + ` FROM purchase_order_items WHERE purchase_order_id = $1`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, purchaseID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.PurchaseOrderItem, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		li, err := scanPurchaseOrderItem(r)
		if err != nil {
			return err
		}
		out = append(out, li)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

type purchaseScan struct {
	branchID, notes, cancelledReason, createdBy, updatedBy sql.NullString
	expectedDate, receivedDate, cancelledAt, deletedAt     sql.NullTime
	status, subtotal, taxAmount, total, paidAmount         string
	discountAmount, currencyCode, exchangeRate             string
}

func scanPurchaseOrder(row *sql.Row) (*purchasing.PurchaseOrder, error) {
	p := &purchasing.PurchaseOrder{}
	var s purchaseScan
	err := persistence.ScanRow(row,
		&p.ID, &p.CompanyID, &p.SupplierID, &s.branchID, &p.Number, &p.OrderDate,
		&s.expectedDate, &s.receivedDate, &s.status, &s.subtotal, &s.taxAmount, &s.total,
		&s.paidAmount, &s.discountAmount, &s.currencyCode, &s.exchangeRate, &s.notes,
		&s.cancelledAt, &s.cancelledReason, &p.CreatedAt, &p.UpdatedAt, &s.deletedAt,
		&s.createdBy, &s.updatedBy,
	)
	if err != nil {
		return nil, err
	}
	if err := decodePurchaseOrder(p, &s); err != nil {
		return nil, err
	}
	return p, nil
}

func scanPurchaseOrderFromRows(rows *sql.Rows) (*purchasing.PurchaseOrder, error) {
	p := &purchasing.PurchaseOrder{}
	var s purchaseScan
	if err := rows.Scan(
		&p.ID, &p.CompanyID, &p.SupplierID, &s.branchID, &p.Number, &p.OrderDate,
		&s.expectedDate, &s.receivedDate, &s.status, &s.subtotal, &s.taxAmount, &s.total,
		&s.paidAmount, &s.discountAmount, &s.currencyCode, &s.exchangeRate, &s.notes,
		&s.cancelledAt, &s.cancelledReason, &p.CreatedAt, &p.UpdatedAt, &s.deletedAt,
		&s.createdBy, &s.updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodePurchaseOrder(p, &s); err != nil {
		return nil, err
	}
	return p, nil
}

func decodePurchaseOrder(p *purchasing.PurchaseOrder, s *purchaseScan) error {
	if s.branchID.Valid {
		id := persistence.ParseUUID(s.branchID.String)
		p.BranchID = &id
	}
	if s.expectedDate.Valid {
		d := valueobjects.Date(s.expectedDate.Time)
		p.ExpectedDate = &d
	}
	if s.receivedDate.Valid {
		d := valueobjects.Date(s.receivedDate.Time)
		p.ReceivedDate = &d
	}
	if s.cancelledAt.Valid {
		t := s.cancelledAt.Time
		p.CancelledAt = &t
	}
	if s.notes.Valid {
		p.Notes = s.notes.String
	}
	if s.cancelledReason.Valid {
		p.CancelledReason = s.cancelledReason.String
	}
	if s.createdBy.Valid {
		id := persistence.ParseUUID(s.createdBy.String)
		p.CreatedBy = &id
	}
	if s.updatedBy.Valid {
		id := persistence.ParseUUID(s.updatedBy.String)
		p.UpdatedBy = &id
	}
	p.Status = persistence.ParsePurchaseStatus(s.status)
	var err error
	if p.CurrencyCode, err = valueobjects.NewCurrencyCode(s.currencyCode); err != nil {
		return err
	}
	if p.ExchangeRate, err = valueobjects.ExchangeRateFromString(s.exchangeRate); err != nil {
		return err
	}
	if p.Subtotal, err = persistence.ParseMoney(s.subtotal); err != nil {
		return err
	}
	if p.TaxAmount, err = persistence.ParseMoney(s.taxAmount); err != nil {
		return err
	}
	if p.Total, err = persistence.ParseMoney(s.total); err != nil {
		return err
	}
	if p.Paid, err = persistence.ParseMoney(s.paidAmount); err != nil {
		return err
	}
	if p.DiscountAmount, err = persistence.ParseMoney(s.discountAmount); err != nil {
		return err
	}
	p.Items = []*purchasing.PurchaseOrderItem{}
	return nil
}

func scanPurchaseOrderItem(rows *sql.Rows) (*purchasing.PurchaseOrderItem, error) {
	li := &purchasing.PurchaseOrderItem{}
	var (
		qtyOrdered, unitCost, discountAmount, taxAmount  string
		discountPercent, taxRate                         string
		ignoredReceived, ignoredLineTotal                string
		description                                      sql.NullString
	)
	if err := rows.Scan(
		&li.ID, &li.PurchaseOrderID, &li.ProductID, &li.LineNumber,
		&qtyOrdered, &ignoredReceived, &unitCost, &discountPercent, &discountAmount,
		&taxRate, &taxAmount, &ignoredLineTotal, &description, &li.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	var err error
	if li.Quantity, err = valueobjects.QuantityFromString(qtyOrdered); err != nil {
		return nil, err
	}
	if li.UnitPrice, err = persistence.ParseMoney(unitCost); err != nil {
		return nil, err
	}
	if li.DiscountPercent, err = valueobjects.PercentageFromString(discountPercent); err != nil {
		return nil, err
	}
	if li.DiscountAmount, err = persistence.ParseMoney(discountAmount); err != nil {
		return nil, err
	}
	if li.TaxRate, err = valueobjects.PercentageFromString(taxRate); err != nil {
		return nil, err
	}
	if li.TaxAmount, err = persistence.ParseMoney(taxAmount); err != nil {
		return nil, err
	}
	if description.Valid {
		li.Description = description.String
	}
	return li, nil
}

var _ purchasing.PurchaseRepository = (*purchaseRepository)(nil)
