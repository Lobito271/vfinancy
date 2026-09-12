package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/sales"
)

type saleRepository struct {
	q persistence.Querier
}

// NewSaleRepository returns a SaleRepository backed by *sql.DB.
func NewSaleRepository(db *sql.DB) *saleRepository {
	return &saleRepository{q: persistence.FromDB(db)}
}

const saleColumns = `
	id, customer_id, number, sale_date, due_date, status, sale_type,
	total, paid_amount, cost_total, profit, notes, cancelled_at, cancelled_reason,
	created_at, updated_at, deleted_at, created_by, updated_by
`

const saleItemColumns = `
	id, sale_id, product_id, inventory_batch_id, line_number, quantity,
	unit_price, line_total, cost_snapshot, description, created_at
`

func (r *saleRepository) Create(ctx context.Context, s *sales.Sale, items []*sales.SaleItem) error {
	const q = `INSERT INTO sales (
		id, customer_id, number, sale_date, due_date, status, sale_type,
		total, paid_amount, cost_total, profit, notes, cancelled_at, cancelled_reason,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.ID, s.CustomerID, s.Number, s.SaleDate,
		persistence.NullIfZeroTime(s.DueDate), s.Status.String(), s.SaleType.String(),
		s.Total.String(), s.PaidAmount.String(), s.CostTotal.String(), s.Profit.String(),
		persistence.NullIfEmpty(s.Notes),
		persistence.NullIfZeroTime(s.CancelledAt), persistence.NullIfEmpty(s.CancelledReason),
		s.CreatedAt, s.UpdatedAt, persistence.NullIfZeroTime(s.DeletedAt),
		persistence.NullIfEmptyUUID(s.CreatedBy), persistence.NullIfEmptyUUID(s.UpdatedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	for _, li := range items {
		if err := r.insertItem(ctx, li); err != nil {
			return err
		}
	}
	return nil
}

func (r *saleRepository) insertItem(ctx context.Context, li *sales.SaleItem) error {
	const q = `INSERT INTO sale_items (
		id, sale_id, product_id, inventory_batch_id, line_number, quantity,
		unit_price, line_total, cost_snapshot, description, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		li.ID, li.SaleID, li.ProductID, persistence.NullIfEmptyUUID(li.InventoryBatchID),
		li.LineNumber, li.Quantity.String(), li.UnitPrice.String(),
		li.LineTotal.String(), li.CostSnapshot.String(),
		persistence.NullIfEmpty(li.Description), li.CreatedAt,
	)
	if err != nil {
		return persistence.Translate(err)
	}
	return nil
}

func (r *saleRepository) Update(ctx context.Context, s *sales.Sale) error {
	const q = `UPDATE sales SET
		due_date = $1, status = $2, total = $3, paid_amount = $4, cost_total = $5,
		profit = $6, notes = $7, cancelled_at = $8, cancelled_reason = $9,
		updated_at = $10, updated_by = $11
	 WHERE id = $12 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		persistence.NullIfZeroTime(s.DueDate), s.Status.String(),
		s.Total.String(), s.PaidAmount.String(), s.CostTotal.String(), s.Profit.String(),
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

func (r *saleRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
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
	items, err := r.ListItems(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.Items = items
	return s, nil
}

func (r *saleRepository) ListItems(ctx context.Context, saleID uuid.UUID) ([]*sales.SaleItem, error) {
	q := `SELECT ` + saleItemColumns + ` FROM sale_items WHERE sale_id = $1 ORDER BY line_number`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, saleID)
	if err != nil {
		return nil, persistence.Translate(err)
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
		return nil, err
	}
	return items, nil
}

func (r *saleRepository) NextNumber(ctx context.Context) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	like := fmt.Sprintf("V-%d-%%", year)
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM sales WHERE number LIKE $1`, like).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("V-%d-%05d", year, n+1), nil
}

func (r *saleRepository) List(ctx context.Context, filter sales.SaleFilter) (repositories.Page[*sales.Sale], error) {
	var clauses []string
	var args []any
	if !filter.IncludeDeleted {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("(number LIKE $%d OR notes LIKE $%d)", len(args)+1, len(args)+1))
		args = append(args, "%"+filter.Search+"%")
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.SaleType != "" {
		clauses = append(clauses, fmt.Sprintf("sale_type = $%d", len(args)+1))
		args = append(args, filter.SaleType)
	}
	if filter.CustomerID != nil {
		clauses = append(clauses, fmt.Sprintf("customer_id = $%d", len(args)+1))
		args = append(args, *filter.CustomerID)
	}
	if filter.From != nil {
		clauses = append(clauses, fmt.Sprintf("sale_date >= $%d", len(args)+1))
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		clauses = append(clauses, fmt.Sprintf("sale_date <= $%d", len(args)+1))
		args = append(args, *filter.To)
	}
	if filter.OnlyUnpaid {
		clauses = append(clauses, "status IN ('pending', 'partial')")
	}
	where := persistence.JoinClauses(clauses)
	if where == "" {
		where = "1=1"
	}

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM sales WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*sales.Sale]{}, persistence.Translate(err)
	}

	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)
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

func scanSale(row *sql.Row) (*sales.Sale, error) {
	s := &sales.Sale{}
	var (
		dueDate, cancelledAt, deletedAt     sql.NullTime
		notes, reason, createdBy, updatedBy sql.NullString
		saleDate                            time.Time
		status, saleType                    string
		total, paid, cost, profit           string
	)
	if err := persistence.ScanRow(row,
		&s.ID, &s.CustomerID, &s.Number, &saleDate, &dueDate, &status, &saleType,
		&total, &paid, &cost, &profit, &notes, &cancelledAt, &reason,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, err
	}
	s.SaleDate = saleDate
	return s, decodeSale(s, dueDate, cancelledAt, deletedAt, notes, reason, createdBy, updatedBy, status, saleType, total, paid, cost, profit)
}

func scanSaleFromRows(rows *sql.Rows) (*sales.Sale, error) {
	s := &sales.Sale{}
	var (
		dueDate, cancelledAt, deletedAt     sql.NullTime
		notes, reason, createdBy, updatedBy sql.NullString
		saleDate                            time.Time
		status, saleType                    string
		total, paid, cost, profit           string
	)
	if err := rows.Scan(
		&s.ID, &s.CustomerID, &s.Number, &saleDate, &dueDate, &status, &saleType,
		&total, &paid, &cost, &profit, &notes, &cancelledAt, &reason,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	s.SaleDate = saleDate
	return s, decodeSale(s, dueDate, cancelledAt, deletedAt, notes, reason, createdBy, updatedBy, status, saleType, total, paid, cost, profit)
}

func decodeSale(s *sales.Sale, dueDate, cancelledAt, deletedAt sql.NullTime, notes, reason, createdBy, updatedBy sql.NullString, status, saleType, total, paid, cost, profit string) error {
	if dueDate.Valid {
		t := dueDate.Time
		s.DueDate = &t
	}
	s.Status = persistence.ParseSaleStatus(status)
	s.SaleType = parseSaleType(saleType)
	var err error
	if s.Total, err = persistence.ParseMoney(total); err != nil {
		return err
	}
	if s.PaidAmount, err = persistence.ParseMoney(paid); err != nil {
		return err
	}
	if s.CostTotal, err = persistence.ParseMoney(cost); err != nil {
		return err
	}
	if s.Profit, err = persistence.ParseMoney(profit); err != nil {
		return err
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
	if deletedAt.Valid {
		t := deletedAt.Time
		s.DeletedAt = &t
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

func parseSaleType(s string) enums.SaleType {
	st := enums.SaleType(s)
	if !st.Valid() {
		return enums.SaleType("")
	}
	return st
}

func scanSaleItem(rows *sql.Rows) (*sales.SaleItem, error) {
	li := &sales.SaleItem{}
	var (
		batchID, description    sql.NullString
		quantity, unitPrice     string
		lineTotal, costSnapshot string
	)
	if err := rows.Scan(
		&li.ID, &li.SaleID, &li.ProductID, &batchID, &li.LineNumber,
		&quantity, &unitPrice, &lineTotal, &costSnapshot, &description, &li.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if batchID.Valid {
		id := persistence.ParseUUID(batchID.String)
		li.InventoryBatchID = &id
	}
	q, err := valueobjects.QuantityFromString(quantity)
	if err != nil {
		return nil, err
	}
	li.Quantity = q
	if li.UnitPrice, err = persistence.ParseMoney(unitPrice); err != nil {
		return nil, err
	}
	if li.LineTotal, err = persistence.ParseMoney(lineTotal); err != nil {
		return nil, err
	}
	if li.CostSnapshot, err = persistence.ParseMoney(costSnapshot); err != nil {
		return nil, err
	}
	if description.Valid {
		li.Description = description.String
	}
	return li, nil
}

var _ sales.SaleRepository = (*saleRepository)(nil)
