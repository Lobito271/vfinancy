package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
)

var _ inventory.InventoryMovementRepository = (*inventoryMovementRepository)(nil)

type inventoryMovementRepository struct {
	q persistence.Querier
}

// NewInventoryMovementRepository returns an InventoryMovementRepository
// backed by *sql.DB.
func NewInventoryMovementRepository(db *sql.DB) *inventoryMovementRepository {
	return &inventoryMovementRepository{q: persistence.FromDB(db)}
}

const movementColumns = `
	id, batch_id, product_id, movement_date, type,
	reference_type, reference_id, quantity_delta, balance_after,
	unit_cost, notes, created_at
`

func (r *inventoryMovementRepository) Create(ctx context.Context, m *inventory.InventoryMovement) error {
	var refType, refID any
	if m.Reference != nil {
		refType = m.Reference.Type.String()
		refID = m.Reference.ID
	}
	const q = `INSERT INTO inventory_movements (
		id, batch_id, product_id, movement_date, type,
		reference_type, reference_id, quantity_delta, balance_after,
		unit_cost, notes, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		m.ID, m.BatchID, m.ProductID, m.MovementDate, m.Type.String(),
		refType, refID, m.QuantityDelta.String(), m.BalanceAfter.String(),
		m.UnitCost.String(), persistence.NullIfEmpty(m.Notes), m.CreatedAt,
	)
	return persistence.Translate(err)
}

func (r *inventoryMovementRepository) List(ctx context.Context, filter inventory.InventoryMovementFilter) (repositories.Page[*inventory.InventoryMovement], error) {
	var (
		clauses = []string{"TRUE"}
		args    []any
	)
	if filter.ProductID != nil {
		clauses = append(clauses, fmt.Sprintf("product_id = $%d", len(args)+1))
		args = append(args, *filter.ProductID)
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

func scanMovementFromRows(rows *sql.Rows) (*inventory.InventoryMovement, error) {
	return scanMovementInto(rows.Scan)
}

func scanMovementInto(scan func(dest ...any) error) (*inventory.InventoryMovement, error) {
	m := &inventory.InventoryMovement{}
	var (
		mvtType, quantity, balanceAfter, unitCost string
		refType, refID, notes                     sql.NullString
	)
	if err := scan(
		&m.ID, &m.BatchID, &m.ProductID, &m.MovementDate, &mvtType,
		&refType, &refID,
		&quantity, &balanceAfter, &unitCost,
		&notes, &m.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	m.Type = persistence.ParseMovementType(mvtType)
	q, err := valueobjects.QuantityFromString(quantity)
	if err != nil {
		return nil, err
	}
	m.QuantityDelta = q
	if b, err := valueobjects.QuantityFromString(balanceAfter); err != nil {
		return nil, err
	} else {
		m.BalanceAfter = b
	}
	if v, err := persistence.ParseMoney(unitCost); err != nil {
		return nil, err
	} else {
		m.UnitCost = v
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
	return m, nil
}
