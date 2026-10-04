package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
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
	status, currency_code, payment_method, exchange_rate, notes,
	customer_id, supplier_id, credit_card_id, sale_id,
	cost_usd, extra_cost_usd, total_cost_usd, sale_price_pen, real_cost_pen, refund_amount,
	faulty, faulty_reason, cancelled_at, cancelled_reason,
	created_at, updated_at, deleted_at
`

// purchasesListSelect extends the base columns with the supplier, the
// customer and the originating sale of a client order; both single-order
// and list queries go through it so the "Proveedor" and "Cliente"
// columns resolve even for soft-deleted suppliers and customers.
const purchasesListSelect = `
	purchases.id, purchases.number, purchases.order_date,
	purchases.expected_date, purchases.received_date, purchases.arrival_date,
	purchases.status, purchases.currency_code, purchases.payment_method,
	purchases.exchange_rate, purchases.notes,
	purchases.customer_id, purchases.supplier_id, purchases.credit_card_id,
	purchases.sale_id,
	purchases.cost_usd, purchases.extra_cost_usd, purchases.total_cost_usd,
	purchases.sale_price_pen, purchases.real_cost_pen,
	purchases.refund_amount,
	purchases.faulty, purchases.faulty_reason, purchases.cancelled_at,
	purchases.cancelled_reason,
	purchases.created_at, purchases.updated_at, purchases.deleted_at,
	COALESCE(s.name, ''), COALESCE(c.business_name, ''), COALESCE(sa.number, '')
`

// purchasesListFrom is the join set matching purchasesListSelect.
const purchasesListFrom = `
	FROM purchases
	LEFT JOIN suppliers s ON s.id = purchases.supplier_id
	LEFT JOIN customers c ON c.id = purchases.customer_id
	LEFT JOIN sales sa ON sa.id = purchases.sale_id
`

const purchaseItemColumns = `
	id, purchase_id, product_id, line_number, description, unit_code,
	quantity_ordered, quantity_received, unit_cost_usd, line_total_usd,
	sale_price_pen, created_at
`

// quantitySoldExpr builds the net quantity sold out of the lots created
// from a purchase line. It is derived from the movement ledger (sale
// issues minus void refunds) so stock adjustments do not distort it, and
// returned as text so callers keep the decimal precision. It is a
// function rather than a var because the cast depends on the dialect,
// which is only known once the process has called SetDialect.
func quantitySoldExpr() string {
	return `CAST(COALESCE((SELECT SUM(CASE WHEN m.type = 'sale'
		THEN -` + persistence.Num("m.quantity_delta") + `
		ELSE ` + persistence.Num("m.quantity_delta") + ` END)
	FROM inventory_movements m
	JOIN inventory_batches b ON b.id = m.batch_id
	WHERE b.purchase_item_id = purchase_items.id
		AND m.type IN ('sale', 'void_sale')), 0) AS TEXT)`
}

// purchaseItemListSelect adds the sold quantity to the line columns.
func purchaseItemListSelect() string { return purchaseItemColumns + ", " + quantitySoldExpr() }

// Create inserts the order and all of its items.
func (r *purchaseRepository) Create(ctx context.Context, po *purchasing.Purchase, items []*purchasing.PurchaseItem) error {
	const q = `INSERT INTO purchases (
		id, number, order_date, expected_date, received_date, arrival_date,
		status, currency_code, payment_method, exchange_rate, notes,
		customer_id, supplier_id, credit_card_id, sale_id,
		cost_usd, extra_cost_usd, total_cost_usd, sale_price_pen, real_cost_pen, refund_amount,
		faulty, faulty_reason, cancelled_at, cancelled_reason,
		created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		po.ID, po.Number, po.OrderDate,
		persistence.NullIfZeroTime(po.ExpectedDate), persistence.NullIfZeroTime(po.ReceivedDate),
		persistence.NullIfZeroTime(po.ArrivalDate),
		po.Status.String(), po.CurrencyCode, string(po.PaymentMethod), po.ExchangeRate.String(),
		persistence.NullIfEmpty(po.Notes),
		persistence.NullIfEmptyUUID(po.CustomerID), persistence.NullIfEmptyUUID(po.SupplierID),
		persistence.NullIfEmptyUUID(po.CreditCardID), persistence.NullIfEmptyUUID(po.SaleID),
		po.CostUSD.String(), po.ExtraCostUSD.String(), po.TotalCostUSD.String(),
		po.SalePricePen.String(), po.RealCostPen.String(), po.RefundAmount.String(),
		po.Faulty, persistence.NullIfEmpty(po.FaultyReason),
		persistence.NullIfZeroTime(po.CancelledAt), persistence.NullIfEmpty(po.CancelledReason),
		po.CreatedAt, po.UpdatedAt,
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

func (r *purchaseRepository) insertItem(ctx context.Context, purchaseID uuid.UUID, li *purchasing.PurchaseItem) error {
	const q = `INSERT INTO purchase_items (
		id, purchase_id, product_id, line_number, description, unit_code,
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
func (r *purchaseRepository) Update(ctx context.Context, po *purchasing.Purchase) error {
	const q = `UPDATE purchases SET
		number = $1, expected_date = $2, received_date = $3, arrival_date = $4,
		status = $5, payment_method = $6, notes = $7,
		customer_id = $8, supplier_id = $9, credit_card_id = $10, sale_id = $11,
		cost_usd = $12, extra_cost_usd = $13, total_cost_usd = $14,
		sale_price_pen = $15, real_cost_pen = $16,
		refund_amount = $17,
		faulty = $18, faulty_reason = $19, cancelled_at = $20, cancelled_reason = $21,
		updated_at = $22
	 WHERE id = $23 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		po.Number,
		persistence.NullIfZeroTime(po.ExpectedDate), persistence.NullIfZeroTime(po.ReceivedDate),
		persistence.NullIfZeroTime(po.ArrivalDate), po.Status.String(),
		string(po.PaymentMethod), persistence.NullIfEmpty(po.Notes),
		persistence.NullIfEmptyUUID(po.CustomerID), persistence.NullIfEmptyUUID(po.SupplierID),
		persistence.NullIfEmptyUUID(po.CreditCardID), persistence.NullIfEmptyUUID(po.SaleID),
		po.CostUSD.String(), po.ExtraCostUSD.String(), po.TotalCostUSD.String(),
		po.SalePricePen.String(), po.RealCostPen.String(), po.RefundAmount.String(),
		po.Faulty, persistence.NullIfEmpty(po.FaultyReason),
		persistence.NullIfZeroTime(po.CancelledAt), persistence.NullIfEmpty(po.CancelledReason),
		time.Now().UTC(), po.ID,
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
		`UPDATE purchases SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`,
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
func (r *purchaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*purchasing.Purchase, error) {
	q := `SELECT ` + purchasesListSelect + purchasesListFrom + `
		WHERE purchases.id = $1 AND purchases.deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanPurchase(func(dest ...any) error { return persistence.ScanRow(row, dest...) })
}

// ListItems returns the lines of an order ordered by line number.
func (r *purchaseRepository) ListItems(ctx context.Context, purchaseOrderID uuid.UUID) ([]*purchasing.PurchaseItem, error) {
	q := `SELECT ` + purchaseItemListSelect() + `
		FROM purchase_items
		WHERE purchase_id = $1
		ORDER BY line_number`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, purchaseOrderID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.PurchaseItem, 0)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		li, err := scanPurchaseItem(row)
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
		`UPDATE purchase_items SET quantity_received = $1 WHERE id = $2`,
		received.String(), itemID)
	return persistence.Translate(err)
}

// NextNumber returns the next "PO-" zero-padded sequence number for
// the current year.
func (r *purchaseRepository) NextNumber(ctx context.Context) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM purchases WHERE number LIKE $1`,
		fmt.Sprintf("PO-%d-%%", year)).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("PO-%d-%05d", year, n+1), nil
}

// List returns the orders matching the filter.
func (r *purchaseRepository) List(ctx context.Context, filter purchasing.PurchaseFilter) (repositories.Page[*purchasing.Purchase], error) {
	var clauses []string
	var args []any
	if !filter.IncludeDeleted {
		clauses = append(clauses, "purchases.deleted_at IS NULL")
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("purchases.number LIKE $%d", len(args)+1))
		args = append(args, "%"+filter.Search+"%")
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("purchases.status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.CreditCardID != nil {
		clauses = append(clauses, fmt.Sprintf("purchases.credit_card_id = $%d", len(args)+1))
		args = append(args, *filter.CreditCardID)
	}
	if filter.From != nil {
		clauses = append(clauses, fmt.Sprintf("purchases.order_date >= $%d", len(args)+1))
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		clauses = append(clauses, fmt.Sprintf("purchases.order_date <= $%d", len(args)+1))
		args = append(args, *filter.To)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + persistence.JoinClauses(clauses)
	}

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		"SELECT count(*) FROM purchases"+where, args...).Scan(&total); err != nil {
		return repositories.Page[*purchasing.Purchase]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	query := `SELECT ` + purchasesListSelect + purchasesListFrom + where +
		fmt.Sprintf(` ORDER BY purchases.order_date DESC, purchases.number DESC LIMIT $%d OFFSET $%d`, limitPos, offsetPos)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, query, args...)
	if err != nil {
		return repositories.Page[*purchasing.Purchase]{}, persistence.Translate(err)
	}
	out := make([]*purchasing.Purchase, 0, limit)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		p, err := scanPurchase(row.Scan)
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return repositories.Page[*purchasing.Purchase]{}, err
	}
	return repositories.Page[*purchasing.Purchase]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

// ListLineSummaries returns the line descriptions of the given orders,
// keyed by order id and ordered by line number, for the list's
// products column and its sold-out flag.
func (r *purchaseRepository) ListLineSummaries(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]purchasing.PurchaseLineSummary, error) {
	out := make(map[uuid.UUID][]purchasing.PurchaseLineSummary, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT purchase_id, description, quantity_ordered, `+quantitySoldExpr()+`
		 FROM purchase_items
		 WHERE purchase_id IN (`+strings.Join(placeholders, ",")+`)
		 ORDER BY purchase_id, line_number`,
		args...)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		var orderID uuid.UUID
		var name, ordered, sold string
		if err := row.Scan(&orderID, &name, &ordered, &sold); err != nil {
			return persistence.Translate(err)
		}
		out[orderID] = append(out[orderID], purchasing.PurchaseLineSummary{
			Quantity: ordered, Name: name, Ordered: ordered, Sold: sold,
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateExtraCost inserts one extra cost against an order.
func (r *purchaseRepository) CreateExtraCost(ctx context.Context, ec *purchasing.ExtraCost) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`INSERT INTO purchase_extra_costs (
			id, purchase_id, concept, amount, currency_code, exchange_rate,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ec.ID, ec.PurchaseID, strings.TrimSpace(ec.Concept),
		ec.Amount.String(), ec.CurrencyCode.String(), ec.ExchangeRate.String(),
		ec.CreatedAt, ec.UpdatedAt,
	)
	return persistence.Translate(err)
}

// UpdateExtraCost persists the mutable fields of an extra cost scoped
// to its parent order.
func (r *purchaseRepository) UpdateExtraCost(ctx context.Context, ec *purchasing.ExtraCost) error {
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`UPDATE purchase_extra_costs SET
			concept = $1, amount = $2, currency_code = $3, exchange_rate = $4,
			updated_at = $5
		 WHERE id = $6 AND purchase_id = $7`,
		strings.TrimSpace(ec.Concept), ec.Amount.String(), ec.CurrencyCode.String(),
		ec.ExchangeRate.String(), time.Now().UTC(), ec.ID, ec.PurchaseID,
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

// DeleteExtraCost removes one extra cost scoped to its parent order.
func (r *purchaseRepository) DeleteExtraCost(ctx context.Context, id, purchaseOrderID uuid.UUID) error {
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`DELETE FROM purchase_extra_costs WHERE id = $1 AND purchase_id = $2`,
		id, purchaseOrderID)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

// ListExtraCosts returns the extra costs of an order, newest first.
func (r *purchaseRepository) ListExtraCosts(ctx context.Context, purchaseOrderID uuid.UUID) ([]*purchasing.ExtraCost, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT id, purchase_id, concept, amount, currency_code, exchange_rate,
		        created_at, updated_at
		 FROM purchase_extra_costs
		 WHERE purchase_id = $1
		 ORDER BY created_at DESC, id DESC`,
		purchaseOrderID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.ExtraCost, 0)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		ec := &purchasing.ExtraCost{}
		var amount, rate, currency string
		if err := row.Scan(
			&ec.ID, &ec.PurchaseID, &ec.Concept, &amount, &currency, &rate,
			&ec.CreatedAt, &ec.UpdatedAt,
		); err != nil {
			return persistence.Translate(err)
		}
		m, err := persistence.ParseMoney(amount)
		if err != nil {
			return err
		}
		ec.Amount = m
		er, err := valueobjects.ExchangeRateFromString(rate)
		if err != nil {
			return err
		}
		ec.ExchangeRate = er
		if cc, err := valueobjects.NewCurrencyCode(currency); err == nil {
			ec.CurrencyCode = cc
		}
		out = append(out, ec)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

type purchaseScan struct {
	notes, faultyReason, cancelledReason                             sql.NullString
	expectedDate, receivedDate, arrivalDate, cancelledAt, deletedAt  sql.NullTime
	customerID, supplierID, creditCardID, saleID                     sql.NullString
	status, currencyCode, paymentMethod, exchangeRate                string
	supplierName, customerName, saleNumber                           string
	costUSD, extraCostUSD, totalCostUSD                            string
	salePricePen, realCostPen, refundAmount                        string
	faulty                                                           bool
}

// scanPurchase decodes the purchasesListSelect columns of one order.
// Single-order and list queries share it so the destination list can
// never drift from the select.
func scanPurchase(scan func(dest ...any) error) (*purchasing.Purchase, error) {
	p := &purchasing.Purchase{}
	var s purchaseScan
	if err := scan(
		&p.ID, &p.Number, &p.OrderDate,
		&s.expectedDate, &s.receivedDate, &s.arrivalDate,
		&s.status, &s.currencyCode, &s.paymentMethod, &s.exchangeRate, &s.notes,
		&s.customerID, &s.supplierID, &s.creditCardID, &s.saleID,
		&s.costUSD, &s.extraCostUSD, &s.totalCostUSD,
		&s.salePricePen, &s.realCostPen, &s.refundAmount,
		&s.faulty, &s.faultyReason, &s.cancelledAt, &s.cancelledReason,
		&p.CreatedAt, &p.UpdatedAt, &s.deletedAt,
		&s.supplierName, &s.customerName, &s.saleNumber,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodePurchase(p, &s); err != nil {
		return nil, err
	}
	return p, nil
}

func decodePurchase(p *purchasing.Purchase, s *purchaseScan) error {
	if s.customerID.Valid {
		id := persistence.ParseUUID(s.customerID.String)
		p.CustomerID = &id
	}
	if s.supplierID.Valid {
		id := persistence.ParseUUID(s.supplierID.String)
		p.SupplierID = &id
	}
	if s.creditCardID.Valid {
		id := persistence.ParseUUID(s.creditCardID.String)
		p.CreditCardID = &id
	}
	if s.saleID.Valid {
		id := persistence.ParseUUID(s.saleID.String)
		p.SaleID = &id
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
	p.Status = persistence.ParsePurchaseStatus(s.status)
	p.PaymentMethod = purchasing.PurchasePaymentMethod(s.paymentMethod)
	p.SupplierName = s.supplierName
	p.CustomerName = s.customerName
	p.SaleNumber = s.saleNumber
	p.Faulty = s.faulty
	p.CurrencyCode = s.currencyCode
	var err error
	if p.ExchangeRate, err = valueobjects.ExchangeRateFromString(s.exchangeRate); err != nil {
		return err
	}
	if p.CostUSD, err = persistence.ParseMoney(s.costUSD); err != nil {
		return err
	}
	if p.ExtraCostUSD, err = persistence.ParseMoney(s.extraCostUSD); err != nil {
		return err
	}
	if p.TotalCostUSD, err = persistence.ParseMoney(s.totalCostUSD); err != nil {
		return err
	}
	if p.SalePricePen, err = persistence.ParseMoney(s.salePricePen); err != nil {
		return err
	}
	if p.RealCostPen, err = persistence.ParseMoney(s.realCostPen); err != nil {
		return err
	}
	if p.RefundAmount, err = persistence.ParseMoney(s.refundAmount); err != nil {
		return err
	}
	p.Items = []*purchasing.PurchaseItem{}
	return nil
}

func scanPurchaseItem(rows *sql.Rows) (*purchasing.PurchaseItem, error) {
	li := &purchasing.PurchaseItem{}
	var (
		productID, description                                  sql.NullString
		qtyOrdered, qtyReceived, unitCost, lineTotal, salePrice string
		qtySold                                                 sql.NullString
	)
	if err := rows.Scan(
		&li.ID, &li.PurchaseID, &productID, &li.LineNumber, &description, &li.UnitCode,
		&qtyOrdered, &qtyReceived, &unitCost, &lineTotal, &salePrice, &li.CreatedAt, &qtySold,
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
	li.QuantitySold = valueobjects.ZeroQuantity()
	if qtySold.Valid {
		if li.QuantitySold, err = valueobjects.QuantityFromString(qtySold.String); err != nil {
			return nil, err
		}
	}
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
// monthlyCostsQuery groups the non-cancelled purchase costs of
// [from, to) by calendar month of order_date. The PEN column sums the
// stored real_cost_pen rather than re-deriving it from the rate, so the
// dashboard can never disagree with the purchase list: real_cost_pen is
// already the landed value (total cost plus import factor, converted at
// the purchase's own rate). Amounts are cast to text so no precision is
// lost on the way in.
func monthlyCostsQuery() string {
	return `SELECT ` +
		persistence.Year("order_date") + ` AS y, ` +
		persistence.Month("order_date") + ` AS m, ` +
		`CAST(SUM(` + persistence.Num("cost_usd") + `) AS TEXT) AS base_usd, ` +
		`CAST(SUM(` + persistence.Num("extra_cost_usd") + `) AS TEXT) AS extra_usd, ` +
		`CAST(SUM(` + persistence.Num("real_cost_pen") + `) AS TEXT) AS total_pen ` +
		`FROM purchases ` +
		`WHERE deleted_at IS NULL AND status <> 'cancelled' ` +
		`AND order_date >= $1 AND order_date < $2 ` +
		`GROUP BY y, m ORDER BY y, m`
}

func (r *purchaseRepository) MonthlyCosts(ctx context.Context, from, to time.Time) ([]purchasing.MonthlyCost, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, monthlyCostsQuery(), from, to)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]purchasing.MonthlyCost, 0)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		var c purchasing.MonthlyCost
		if err := row.Scan(&c.Year, &c.Month, &c.BaseCostUSD, &c.ExtraCostUSD, &c.TotalCostPEN); err != nil {
			return persistence.Translate(err)
		}
		out = append(out, c)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}
