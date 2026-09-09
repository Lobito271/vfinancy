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
	"vfinancy/backend/internal/features/purchasing"
)

type purchaseRepository struct {
	q persistence.Querier
}

// NewPurchaseRepository returns the SQL purchase repository.
func NewPurchaseRepository(db *sql.DB) *purchaseRepository {
	return &purchaseRepository{q: persistence.FromDB(db)}
}

const purchaseColumns = `
	id, number, order_date, expected_date, received_date, arrival_date,
	status, currency_code, exchange_rate, notes, order_type,
	customer_id, credit_card_id,
	cost_usd, sale_price_pen, real_cost_pen, projected_profit_pen, refund_amount,
	faulty, faulty_reason, cancelled_at, cancelled_reason,
	created_at, updated_at, deleted_at, created_by, updated_by
`

const purchaseItemColumns = `
	id, purchase_order_id, product_id, line_number, description, unit_code,
	quantity_ordered, quantity_received, unit_cost_usd, line_total_usd,
	sale_price_pen, created_at
`

// Create inserts the order and all of its items.
func (r *purchaseRepository) Create(ctx context.Context, po *purchasing.PurchaseOrder, items []*purchasing.PurchaseOrderItem) error {
	const q = `INSERT INTO purchase_orders (
		id, number, order_date, expected_date, received_date, arrival_date,
		status, currency_code, exchange_rate, notes, order_type,
		customer_id, credit_card_id,
		cost_usd, sale_price_pen, real_cost_pen, projected_profit_pen, refund_amount,
		faulty, faulty_reason, cancelled_at, cancelled_reason,
		created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		po.ID, po.Number, po.OrderDate,
		persistence.NullIfZeroTime(po.ExpectedDate), persistence.NullIfZeroTime(po.ReceivedDate),
		persistence.NullIfZeroTime(po.ArrivalDate),
		po.Status.String(), po.CurrencyCode, po.ExchangeRate.String(),
		persistence.NullIfEmpty(po.Notes), po.OrderType.String(),
		persistence.NullIfEmptyUUID(po.CustomerID), persistence.NullIfEmptyUUID(po.CreditCardID),
		po.CostUSD.String(), po.SalePricePen.String(), po.RealCostPen.String(),
		po.ProjectedProfitPen.String(), po.RefundAmount.String(),
		po.Faulty, persistence.NullIfEmpty(po.FaultyReason),
		persistence.NullIfZeroTime(po.CancelledAt), persistence.NullIfEmpty(po.CancelledReason),
		po.CreatedAt, po.UpdatedAt,
		persistence.NullIfEmptyUUID(po.CreatedBy), persistence.NullIfEmptyUUID(po.UpdatedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	for _, li := range items {
		if err := r.insertItem(ctx, po.ID, li); err != nil {
			return err
		}
	}
	return nil
}

func (r *purchaseRepository) insertItem(ctx context.Context, purchaseID uuid.UUID, li *purchasing.PurchaseOrderItem) error {
	const q = `INSERT INTO purchase_order_items (
		id, purchase_order_id, product_id, line_number, description, unit_code,
		quantity_ordered, quantity_received, unit_cost_usd, line_total_usd,
		sale_price_pen, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	createdAt := li.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		li.ID, purchaseID, persistence.NullIfEmptyUUID(li.ProductID), li.LineNumber,
		li.Description, li.UnitCode,
		li.QuantityOrdered.String(), li.QuantityReceived.String(),
		li.UnitCostUSD.String(), li.LineTotalUSD.String(), li.SalePricePen.String(),
		createdAt,
	)
	return persistence.Translate(err)
}

// Update persists the mutable order fields.
func (r *purchaseRepository) Update(ctx context.Context, po *purchasing.PurchaseOrder) error {
	const q = `UPDATE purchase_orders SET
		expected_date = $1, received_date = $2, arrival_date = $3, status = $4,
		notes = $5, customer_id = $6, credit_card_id = $7,
		cost_usd = $8, sale_price_pen = $9, real_cost_pen = $10,
		projected_profit_pen = $11, refund_amount = $12,
		faulty = $13, faulty_reason = $14, cancelled_at = $15, cancelled_reason = $16,
		updated_at = $17, updated_by = $18
	 WHERE id = $19 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		persistence.NullIfZeroTime(po.ExpectedDate), persistence.NullIfZeroTime(po.ReceivedDate),
		persistence.NullIfZeroTime(po.ArrivalDate), po.Status.String(),
		persistence.NullIfEmpty(po.Notes),
		persistence.NullIfEmptyUUID(po.CustomerID), persistence.NullIfEmptyUUID(po.CreditCardID),
		po.CostUSD.String(), po.SalePricePen.String(), po.RealCostPen.String(),
		po.ProjectedProfitPen.String(), po.RefundAmount.String(),
		po.Faulty, persistence.NullIfEmpty(po.FaultyReason),
		persistence.NullIfZeroTime(po.CancelledAt), persistence.NullIfEmpty(po.CancelledReason),
		time.Now().UTC(), persistence.NullIfEmptyUUID(po.UpdatedBy), po.ID,
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

// SoftDelete marks the order as deleted.
func (r *purchaseRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
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

// GetByID loads a single order without its items.
func (r *purchaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*purchasing.PurchaseOrder, error) {
	q := `SELECT ` + purchaseColumns + ` FROM purchase_orders WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanPurchaseOrder(row)
}

// ListItems returns the lines of an order ordered by line number.
func (r *purchaseRepository) ListItems(ctx context.Context, purchaseOrderID uuid.UUID) ([]*purchasing.PurchaseOrderItem, error) {
	q := `SELECT ` + purchaseItemColumns + `
		FROM purchase_order_items
		WHERE purchase_order_id = $1
		ORDER BY line_number`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, purchaseOrderID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.PurchaseOrderItem, 0)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		li, err := scanPurchaseOrderItem(row)
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

// UpdateItemReceipt records the received quantity of a line.
func (r *purchaseRepository) UpdateItemReceipt(ctx context.Context, itemID uuid.UUID, received valueobjects.Quantity) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`UPDATE purchase_order_items SET quantity_received = $1 WHERE id = $2`,
		received.String(), itemID)
	return persistence.Translate(err)
}

// NextNumber returns the next "PO-" zero-padded sequence number for
// the current year.
func (r *purchaseRepository) NextNumber(ctx context.Context) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM purchase_orders WHERE number LIKE $1`,
		fmt.Sprintf("PO-%d-%%", year)).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("PO-%d-%05d", year, n+1), nil
}

// List returns the orders matching the filter.
func (r *purchaseRepository) List(ctx context.Context, filter purchasing.PurchaseFilter) (repositories.Page[*purchasing.PurchaseOrder], error) {
	var clauses []string
	var args []any
	if !filter.IncludeDeleted {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("number LIKE $%d", len(args)+1))
		args = append(args, "%"+filter.Search+"%")
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.OrderType != "" {
		clauses = append(clauses, fmt.Sprintf("order_type = $%d", len(args)+1))
		args = append(args, filter.OrderType)
	}
	if filter.CreditCardID != nil {
		clauses = append(clauses, fmt.Sprintf("credit_card_id = $%d", len(args)+1))
		args = append(args, *filter.CreditCardID)
	}
	if filter.ImportLotID != nil {
		clauses = append(clauses, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM import_lot_purchase_orders m WHERE m.purchase_order_id = purchase_orders.id AND m.import_lot_id = $%d)",
			len(args)+1))
		args = append(args, *filter.ImportLotID)
	}
	if filter.From != nil {
		clauses = append(clauses, fmt.Sprintf("order_date >= $%d", len(args)+1))
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		clauses = append(clauses, fmt.Sprintf("order_date <= $%d", len(args)+1))
		args = append(args, *filter.To)
	}
	if len(clauses) == 0 {
		clauses = append(clauses, "1 = 1")
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)
	where := persistence.JoinClauses(clauses)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		"SELECT count(*) FROM purchase_orders WHERE "+where, args...).Scan(&total); err != nil {
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
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		p, err := scanPurchaseOrderFromRows(row)
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

type purchaseScan struct {
	notes, faultyReason, cancelledReason, createdBy, updatedBy           sql.NullString
	expectedDate, receivedDate, arrivalDate, cancelledAt, deletedAt      sql.NullTime
	customerID, creditCardID                                             sql.NullString
	status, orderType, currencyCode, exchangeRate                        string
	costUSD, salePricePen, realCostPen, projectedProfitPen, refundAmount string
	faulty                                                               bool
}

func scanPurchaseOrder(row *sql.Row) (*purchasing.PurchaseOrder, error) {
	p := &purchasing.PurchaseOrder{}
	var s purchaseScan
	if err := persistence.ScanRow(row,
		&p.ID, &p.Number, &p.OrderDate,
		&s.expectedDate, &s.receivedDate, &s.arrivalDate,
		&s.status, &s.currencyCode, &s.exchangeRate, &s.notes, &s.orderType,
		&s.customerID, &s.creditCardID,
		&s.costUSD, &s.salePricePen, &s.realCostPen, &s.projectedProfitPen, &s.refundAmount,
		&s.faulty, &s.faultyReason, &s.cancelledAt, &s.cancelledReason,
		&p.CreatedAt, &p.UpdatedAt, &s.deletedAt, &s.createdBy, &s.updatedBy,
	); err != nil {
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
		&p.ID, &p.Number, &p.OrderDate,
		&s.expectedDate, &s.receivedDate, &s.arrivalDate,
		&s.status, &s.currencyCode, &s.exchangeRate, &s.notes, &s.orderType,
		&s.customerID, &s.creditCardID,
		&s.costUSD, &s.salePricePen, &s.realCostPen, &s.projectedProfitPen, &s.refundAmount,
		&s.faulty, &s.faultyReason, &s.cancelledAt, &s.cancelledReason,
		&p.CreatedAt, &p.UpdatedAt, &s.deletedAt, &s.createdBy, &s.updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodePurchaseOrder(p, &s); err != nil {
		return nil, err
	}
	return p, nil
}

func decodePurchaseOrder(p *purchasing.PurchaseOrder, s *purchaseScan) error {
	if s.customerID.Valid {
		id := persistence.ParseUUID(s.customerID.String)
		p.CustomerID = &id
	}
	if s.creditCardID.Valid {
		id := persistence.ParseUUID(s.creditCardID.String)
		p.CreditCardID = &id
	}
	if s.expectedDate.Valid {
		t := s.expectedDate.Time
		p.ExpectedDate = &t
	}
	if s.receivedDate.Valid {
		t := s.receivedDate.Time
		p.ReceivedDate = &t
	}
	if s.arrivalDate.Valid {
		t := s.arrivalDate.Time
		p.ArrivalDate = &t
	}
	if s.cancelledAt.Valid {
		t := s.cancelledAt.Time
		p.CancelledAt = &t
	}
	if s.deletedAt.Valid {
		t := s.deletedAt.Time
		p.DeletedAt = &t
	}
	if s.notes.Valid {
		p.Notes = s.notes.String
	}
	if s.faultyReason.Valid {
		p.FaultyReason = s.faultyReason.String
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
	p.OrderType = enums.OrderType(s.orderType)
	p.Faulty = s.faulty
	p.CurrencyCode = s.currencyCode
	var err error
	if p.ExchangeRate, err = valueobjects.ExchangeRateFromString(s.exchangeRate); err != nil {
		return err
	}
	if p.CostUSD, err = persistence.ParseMoney(s.costUSD); err != nil {
		return err
	}
	if p.SalePricePen, err = persistence.ParseMoney(s.salePricePen); err != nil {
		return err
	}
	if p.RealCostPen, err = persistence.ParseMoney(s.realCostPen); err != nil {
		return err
	}
	if p.ProjectedProfitPen, err = persistence.ParseMoney(s.projectedProfitPen); err != nil {
		return err
	}
	if p.RefundAmount, err = persistence.ParseMoney(s.refundAmount); err != nil {
		return err
	}
	p.Items = []*purchasing.PurchaseOrderItem{}
	return nil
}

func scanPurchaseOrderItem(rows *sql.Rows) (*purchasing.PurchaseOrderItem, error) {
	li := &purchasing.PurchaseOrderItem{}
	var (
		productID, description                                  sql.NullString
		qtyOrdered, qtyReceived, unitCost, lineTotal, salePrice string
	)
	if err := rows.Scan(
		&li.ID, &li.PurchaseOrderID, &productID, &li.LineNumber, &description, &li.UnitCode,
		&qtyOrdered, &qtyReceived, &unitCost, &lineTotal, &salePrice, &li.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if productID.Valid {
		id := persistence.ParseUUID(productID.String)
		li.ProductID = &id
	}
	if description.Valid {
		li.Description = description.String
	}
	var err error
	if li.QuantityOrdered, err = valueobjects.QuantityFromString(qtyOrdered); err != nil {
		return nil, err
	}
	if li.QuantityReceived, err = valueobjects.QuantityFromString(qtyReceived); err != nil {
		return nil, err
	}
	if li.UnitCostUSD, err = persistence.ParseMoney(unitCost); err != nil {
		return nil, err
	}
	if li.LineTotalUSD, err = persistence.ParseMoney(lineTotal); err != nil {
		return nil, err
	}
	if li.SalePricePen, err = persistence.ParseMoney(salePrice); err != nil {
		return nil, err
	}
	return li, nil
}

var _ purchasing.PurchaseRepository = (*purchaseRepository)(nil)
