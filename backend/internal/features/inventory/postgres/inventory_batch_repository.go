package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/infrastructure/persistence"
)

var _ inventory.InventoryBatchRepository = (*inventoryBatchRepository)(nil)

type inventoryBatchRepository struct {
	q persistence.Querier
}

func NewInventoryBatchRepository(db *sql.DB) *inventoryBatchRepository {
	return &inventoryBatchRepository{q: persistence.FromDB(db)}
}

const batchColumns = `
	id, company_id, product_id, warehouse_id, supplier_id, purchase_order_item_id,
	lot, batch_code, arrival_date, expiry_date, quantity, original_quantity,
	unit_cost, currency_code, status, clearance_date, is_clearance,
	created_at, updated_at, created_by, updated_by
`

func batchLotValue(l valueobjects.LotNumber) any {
	if l.IsEmpty() {
		return nil
	}
	return l.String()
}

func (r *inventoryBatchRepository) Create(ctx context.Context, b *inventory.InventoryBatch) error {
	const q = `INSERT INTO inventory_batches (
		id, company_id, product_id, warehouse_id, supplier_id, purchase_order_item_id,
		lot, batch_code, arrival_date, expiry_date, quantity, original_quantity,
		unit_cost, currency_code, exchange_rate, status, clearance_date, is_clearance,
		created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		b.ID, b.CompanyID, b.ProductID, b.WarehouseID,
		persistence.NullIfEmptyUUID(b.SupplierID),
		persistence.NullIfEmptyUUID(b.PurchaseLineID),
		batchLotValue(b.LotNumber),
		persistence.NullIfEmpty(b.SerialNumber),
		b.ArrivalDate, persistence.NullIfZeroTime(b.ExpiryDate),
		b.CurrentQuantity.String(), b.InitialQuantity.String(),
		b.UnitCost.String(), b.CurrencyCode.String(),
		"1.000000", b.Status,
		b.MaximumSaleDate(), b.IsClearance(valueobjects.Date(time.Now().UTC())),
		b.CreatedAt, b.UpdatedAt, b.CreatedBy, b.UpdatedBy,
	)
	return persistence.Translate(err)
}

func (r *inventoryBatchRepository) Update(ctx context.Context, b *inventory.InventoryBatch) error {
	const q = `UPDATE inventory_batches SET
		lot = $1, batch_code = $2, arrival_date = $3, expiry_date = $4,
		quantity = $5, original_quantity = $6, unit_cost = $7, currency_code = $8,
		status = $9, clearance_date = $10, is_clearance = $11,
		supplier_id = $12, purchase_order_item_id = $13,
		updated_at = $14, updated_by = $15
	 WHERE id = $16`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		batchLotValue(b.LotNumber),
		persistence.NullIfEmpty(b.SerialNumber),
		b.ArrivalDate, persistence.NullIfZeroTime(b.ExpiryDate),
		b.CurrentQuantity.String(), b.InitialQuantity.String(),
		b.UnitCost.String(), b.CurrencyCode.String(),
		b.Status, b.MaximumSaleDate(), b.IsClearance(valueobjects.Date(time.Now().UTC())),
		persistence.NullIfEmptyUUID(b.SupplierID),
		persistence.NullIfEmptyUUID(b.PurchaseLineID),
		time.Now().UTC(), b.UpdatedBy, b.ID,
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

func (r *inventoryBatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*inventory.InventoryBatch, error) {
	q := `SELECT ` + batchColumns + ` FROM inventory_batches WHERE id = $1`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanBatch(row)
}

func (r *inventoryBatchRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*inventory.InventoryBatch, error) {
	// SQLite has no FOR UPDATE; its single-writer model plus the
	// BEGIN IMMEDIATE transaction gives the needed lock. Postgres
	// locks the row for the duration of the write transaction.
	lock := " FOR UPDATE"
	if persistence.IsSQLite() {
		lock = ""
	}
	q := `SELECT ` + batchColumns + ` FROM inventory_batches WHERE id = $1` + lock
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanBatch(row)
}

func (r *inventoryBatchRepository) List(ctx context.Context, filter inventory.InventoryBatchFilter) (repositories.Page[*inventory.InventoryBatch], error) {
	var (
		clauses = []string{"TRUE"}
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.ProductID != nil {
		clauses = append(clauses, fmt.Sprintf("product_id = $%d", len(args)+1))
		args = append(args, *filter.ProductID)
	}
	if filter.WarehouseID != nil {
		clauses = append(clauses, fmt.Sprintf("warehouse_id = $%d", len(args)+1))
		args = append(args, *filter.WarehouseID)
	}
	if filter.PurchaseLineID != nil {
		clauses = append(clauses, fmt.Sprintf("purchase_order_item_id = $%d", len(args)+1))
		args = append(args, *filter.PurchaseLineID)
	}
	if filter.OnlyActive {
		clauses = append(clauses, "status <> 'depleted' AND status <> 'written_off' AND status <> 'voided'")
	}
	if filter.OnlyClearance {
		clauses = append(clauses, "is_clearance = TRUE")
	}
	if !filter.ArrivalRange.IsZero() {
		if !filter.ArrivalRange.From.IsZero() {
			clauses = append(clauses, fmt.Sprintf("arrival_date >= $%d", len(args)+1))
			args = append(args, filter.ArrivalRange.From)
		}
		if !filter.ArrivalRange.To.IsZero() {
			clauses = append(clauses, fmt.Sprintf("arrival_date <= $%d", len(args)+1))
			args = append(args, filter.ArrivalRange.To)
		}
	}
	where := persistence.JoinClauses(clauses)
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM inventory_batches WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*inventory.InventoryBatch]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM inventory_batches WHERE %s ORDER BY arrival_date DESC LIMIT $%d OFFSET $%d",
			batchColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*inventory.InventoryBatch]{}, persistence.Translate(err)
	}
	out := make([]*inventory.InventoryBatch, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		b, err := scanBatchFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, b)
		return nil
	}); err != nil {
		return repositories.Page[*inventory.InventoryBatch]{}, err
	}
	return repositories.Page[*inventory.InventoryBatch]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *inventoryBatchRepository) ExistsByPurchaseLineID(ctx context.Context, purchaseLineID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM inventory_batches WHERE purchase_order_item_id = $1)`
	var exists bool
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, purchaseLineID).Scan(&exists); err != nil {
		return false, persistence.Translate(err)
	}
	return exists, nil
}

func (r *inventoryBatchRepository) GetStockSummary(ctx context.Context, productID, warehouseID uuid.UUID) (float64, string, error) {
	const q = `SELECT
		COALESCE(SUM(CAST(quantity AS REAL)), 0),
		COALESCE(SUM(CAST(unit_cost AS REAL) * CAST(quantity AS REAL)) / NULLIF(SUM(CAST(quantity AS REAL)), 0), 0)
	FROM inventory_batches
	WHERE product_id = $1 AND warehouse_id = $2 AND status = 'active'`
	var qty, avg float64
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, productID, warehouseID).Scan(&qty, &avg); err != nil {
		return 0, "", persistence.Translate(err)
	}
	return qty, decimal.NewFromFloat(avg).StringFixed(2), nil
}

func (r *inventoryBatchRepository) ListClearance(ctx context.Context, companyID uuid.UUID, at time.Time) ([]*inventory.InventoryBatch, error) {
	q := `SELECT ` + batchColumns + ` FROM inventory_batches
		WHERE company_id = $1 AND is_clearance = TRUE AND status = 'active' AND CAST(quantity AS REAL) > 0
		ORDER BY arrival_date`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, companyID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := []*inventory.InventoryBatch{}
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		b, err := scanBatchFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, b)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *inventoryBatchRepository) SetClearanceDate(ctx context.Context, id uuid.UUID, clearanceDate valueobjects.Date) (int, error) {
	const q = `UPDATE inventory_batches
		SET is_clearance = TRUE, clearance_date = $1, updated_at = $2
		WHERE id = $3 AND is_clearance = FALSE`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, clearanceDate, time.Now().UTC(), id)
	if err != nil {
		return 0, persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func scanBatch(row *sql.Row) (*inventory.InventoryBatch, error) {
	b := &inventory.InventoryBatch{}
	var (
		supplierID, purchaseLineID, lot, batchCode      sql.NullString
		expiryDate, clearanceDate                       sql.NullTime
		createdBy, updatedBy                            sql.NullString
		quantity, originalQuantity, unitCost, currency  string
		status                                          string
		isClearance                                     bool
	)
	err := persistence.ScanRow(row,
		&b.ID, &b.CompanyID, &b.ProductID, &b.WarehouseID,
		&supplierID, &purchaseLineID, &lot, &batchCode,
		&b.ArrivalDate, &expiryDate,
		&quantity, &originalQuantity, &unitCost, &currency,
		&status, &clearanceDate, &isClearance,
		&b.CreatedAt, &b.UpdatedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	if lot.Valid {
		ln, err := valueobjects.NewLotNumber(lot.String)
		if err != nil {
			return nil, err
		}
		b.LotNumber = ln
	}
	if batchCode.Valid {
		b.SerialNumber = batchCode.String
	}
	if supplierID.Valid {
		id := persistence.ParseUUID(supplierID.String)
		b.SupplierID = &id
	}
	if purchaseLineID.Valid {
		id := persistence.ParseUUID(purchaseLineID.String)
		b.PurchaseLineID = &id
	}
	if expiryDate.Valid {
		t := expiryDate.Time
		b.ExpiryDate = &t
	}
	q, err := valueobjects.QuantityFromString(quantity)
	if err != nil {
		return nil, err
	}
	b.CurrentQuantity = q
	oq, err := valueobjects.QuantityFromString(originalQuantity)
	if err != nil {
		return nil, err
	}
	b.InitialQuantity = oq
	if v, err := persistence.ParseMoney(unitCost); err != nil {
		return nil, err
	} else {
		b.UnitCost = v
	}
	if c, err := valueobjects.NewCurrencyCode(currency); err != nil {
		return nil, err
	} else {
		b.CurrencyCode = c
	}
	b.Status = status
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		b.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		b.UpdatedBy = &id
	}
	return b, nil
}

func scanBatchFromRows(rows *sql.Rows) (*inventory.InventoryBatch, error) {
	b := &inventory.InventoryBatch{}
	var (
		supplierID, purchaseLineID, lot, batchCode      sql.NullString
		expiryDate, clearanceDate                       sql.NullTime
		createdBy, updatedBy                            sql.NullString
		quantity, originalQuantity, unitCost, currency  string
		status                                          string
		isClearance                                     bool
	)
	if err := rows.Scan(
		&b.ID, &b.CompanyID, &b.ProductID, &b.WarehouseID,
		&supplierID, &purchaseLineID, &lot, &batchCode,
		&b.ArrivalDate, &expiryDate,
		&quantity, &originalQuantity, &unitCost, &currency,
		&status, &clearanceDate, &isClearance,
		&b.CreatedAt, &b.UpdatedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if lot.Valid {
		ln, err := valueobjects.NewLotNumber(lot.String)
		if err != nil {
			return nil, err
		}
		b.LotNumber = ln
	}
	if batchCode.Valid {
		b.SerialNumber = batchCode.String
	}
	if supplierID.Valid {
		id := persistence.ParseUUID(supplierID.String)
		b.SupplierID = &id
	}
	if purchaseLineID.Valid {
		id := persistence.ParseUUID(purchaseLineID.String)
		b.PurchaseLineID = &id
	}
	if expiryDate.Valid {
		t := expiryDate.Time
		b.ExpiryDate = &t
	}
	q, err := valueobjects.QuantityFromString(quantity)
	if err != nil {
		return nil, err
	}
	b.CurrentQuantity = q
	oq, err := valueobjects.QuantityFromString(originalQuantity)
	if err != nil {
		return nil, err
	}
	b.InitialQuantity = oq
	if v, err := persistence.ParseMoney(unitCost); err != nil {
		return nil, err
	} else {
		b.UnitCost = v
	}
	if c, err := valueobjects.NewCurrencyCode(currency); err != nil {
		return nil, err
	} else {
		b.CurrencyCode = c
	}
	b.Status = status
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		b.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		b.UpdatedBy = &id
	}
	return b, nil
}
