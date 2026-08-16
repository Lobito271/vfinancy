package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/infrastructure/persistence"
)

var _ inventory.InventoryMovementRepository = (*inventoryMovementRepository)(nil)

type inventoryMovementRepository struct {
	q persistence.Querier
}

func NewInventoryMovementRepository(db *sql.DB) *inventoryMovementRepository {
	return &inventoryMovementRepository{q: persistence.FromDB(db)}
}

const movementColumns = `
	id, company_id, batch_id, product_id, warehouse_id, movement_date, type,
	reference_type, reference_id, quantity_delta, balance_after, unit_cost,
	currency_code, notes, created_by, created_at
`

func (r *inventoryMovementRepository) batchBalance(ctx context.Context, batchID uuid.UUID) (valueobjects.Quantity, error) {
	var s string
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, `SELECT quantity FROM inventory_batches WHERE id = $1`, batchID).Scan(&s); err != nil {
		if persistence.IsPgNoRows(err) {
			return valueobjects.ZeroQuantity(), repositories.ErrNotFound
		}
		return valueobjects.ZeroQuantity(), persistence.Translate(err)
	}
	return valueobjects.QuantityFromString(s)
}

func (r *inventoryMovementRepository) Create(ctx context.Context, m *inventory.InventoryMovement) error {
	if m.BatchID == nil {
		return repositories.ErrCheckConstraint
	}
	balance, err := r.batchBalance(ctx, *m.BatchID)
	if err != nil {
		return err
	}
	var refType, refID any
	if m.Reference != nil {
		refType = m.Reference.Type.String()
		refID = m.Reference.ID
	}
	const q = `INSERT INTO inventory_movements (
		id, company_id, batch_id, product_id, warehouse_id, movement_date, type,
		reference_type, reference_id, quantity_delta, balance_after, unit_cost,
		currency_code, notes, created_by, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	_, err = persistence.Q(ctx, r.q).ExecContext(ctx, q,
		m.ID, m.CompanyID, *m.BatchID, m.ProductID, m.WarehouseID,
		m.OccurredAt, m.Type.String(), refType, refID,
		m.Quantity.String(), balance.String(), m.UnitCost.String(),
		m.CurrencyCode.String(), persistence.NullIfEmpty(m.Notes), m.CreatedBy, m.CreatedAt,
	)
	return persistence.Translate(err)
}

func (r *inventoryMovementRepository) GetByID(ctx context.Context, id uuid.UUID) (*inventory.InventoryMovement, error) {
	q := `SELECT ` + movementColumns + ` FROM inventory_movements WHERE id = $1`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanMovement(row)
}

func (r *inventoryMovementRepository) List(ctx context.Context, filter inventory.InventoryMovementFilter) (repositories.Page[*inventory.InventoryMovement], error) {
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
	if filter.BatchID != nil {
		clauses = append(clauses, fmt.Sprintf("batch_id = $%d", len(args)+1))
		args = append(args, *filter.BatchID)
	}
	if filter.ReferenceType != "" {
		clauses = append(clauses, fmt.Sprintf("reference_type = $%d", len(args)+1))
		args = append(args, filter.ReferenceType)
	}
	if filter.ReferenceID != nil {
		clauses = append(clauses, fmt.Sprintf("reference_id = $%d", len(args)+1))
		args = append(args, *filter.ReferenceID)
	}
	if !filter.OccurredRange.IsZero() {
		if !filter.OccurredRange.From.IsZero() {
			clauses = append(clauses, fmt.Sprintf("movement_date >= $%d", len(args)+1))
			args = append(args, filter.OccurredRange.From)
		}
		if !filter.OccurredRange.To.IsZero() {
			clauses = append(clauses, fmt.Sprintf("movement_date <= $%d", len(args)+1))
			args = append(args, filter.OccurredRange.To)
		}
	}
	where := persistence.JoinClauses(clauses)
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM inventory_movements WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*inventory.InventoryMovement]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM inventory_movements WHERE %s ORDER BY movement_date DESC LIMIT $%d OFFSET $%d",
			movementColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*inventory.InventoryMovement]{}, persistence.Translate(err)
	}
	out := make([]*inventory.InventoryMovement, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		m, err := scanMovementFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, m)
		return nil
	}); err != nil {
		return repositories.Page[*inventory.InventoryMovement]{}, err
	}
	return repositories.Page[*inventory.InventoryMovement]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *inventoryMovementRepository) StockAt(ctx context.Context, productID, warehouseID uuid.UUID, at time.Time) (float64, error) {
	var qty float64
	err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COALESCE(SUM(CAST(quantity_delta AS REAL)), 0) FROM inventory_movements
		 WHERE product_id = $1 AND warehouse_id = $2 AND movement_date <= $3`,
		productID, warehouseID, at).Scan(&qty)
	if err != nil {
		return 0, persistence.Translate(err)
	}
	return qty, nil
}

func scanMovement(row *sql.Row) (*inventory.InventoryMovement, error) {
	m := &inventory.InventoryMovement{}
	var (
		batchID                                    uuid.UUID
		mvtType, quantity, balanceAfter, unitCost  string
		currency                                   string
		refType, refID, notes, createdBy           sql.NullString
	)
	err := persistence.ScanRow(row,
		&m.ID, &m.CompanyID, &batchID, &m.ProductID, &m.WarehouseID,
		&m.OccurredAt, &mvtType,
		&refType, &refID,
		&quantity, &balanceAfter, &unitCost,
		&currency, &notes, &createdBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.BatchID = &batchID
	m.Type = persistence.ParseMovementType(mvtType)
	q, err := valueobjects.QuantityFromString(quantity)
	if err != nil {
		return nil, err
	}
	m.Quantity = q
	if v, err := persistence.ParseMoney(unitCost); err != nil {
		return nil, err
	} else {
		m.UnitCost = v
	}
	if c, err := valueobjects.NewCurrencyCode(currency); err != nil {
		return nil, err
	} else {
		m.CurrencyCode = c
	}
	if notes.Valid {
		m.Notes = notes.String
	}
	if refType.Valid && refID.Valid {
		ref, err := valueobjects.NewReference(enums.ReferenceType(refType.String), persistence.ParseUUID(refID.String))
		if err != nil {
			return nil, err
		}
		m.Reference = &ref
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		m.CreatedBy = &id
	}
	return m, nil
}

func scanMovementFromRows(rows *sql.Rows) (*inventory.InventoryMovement, error) {
	m := &inventory.InventoryMovement{}
	var (
		batchID                                    uuid.UUID
		mvtType, quantity, balanceAfter, unitCost  string
		currency                                   string
		refType, refID, notes, createdBy           sql.NullString
	)
	if err := rows.Scan(
		&m.ID, &m.CompanyID, &batchID, &m.ProductID, &m.WarehouseID,
		&m.OccurredAt, &mvtType,
		&refType, &refID,
		&quantity, &balanceAfter, &unitCost,
		&currency, &notes, &createdBy, &m.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	m.BatchID = &batchID
	m.Type = persistence.ParseMovementType(mvtType)
	q, err := valueobjects.QuantityFromString(quantity)
	if err != nil {
		return nil, err
	}
	m.Quantity = q
	if v, err := persistence.ParseMoney(unitCost); err != nil {
		return nil, err
	} else {
		m.UnitCost = v
	}
	if c, err := valueobjects.NewCurrencyCode(currency); err != nil {
		return nil, err
	} else {
		m.CurrencyCode = c
	}
	if notes.Valid {
		m.Notes = notes.String
	}
	if refType.Valid && refID.Valid {
		ref, err := valueobjects.NewReference(enums.ReferenceType(refType.String), persistence.ParseUUID(refID.String))
		if err != nil {
			return nil, err
		}
		m.Reference = &ref
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		m.CreatedBy = &id
	}
	return m, nil
}
