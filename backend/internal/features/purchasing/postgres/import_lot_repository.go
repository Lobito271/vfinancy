package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/purchasing"
)

type importLotRepository struct {
	q persistence.Querier
}

// NewImportLotRepository returns the SQL import-lot repository.
func NewImportLotRepository(db *sql.DB) *importLotRepository {
	return &importLotRepository{q: persistence.FromDB(db)}
}

const importLotColumns = `
	id, code, description, status,
	created_at, updated_at, deleted_at, created_by, updated_by
`

// Create inserts a new lot.
func (r *importLotRepository) Create(ctx context.Context, lot *purchasing.ImportLot) error {
	const q = `INSERT INTO import_lots (
		id, code, description, status,
		created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		lot.ID, lot.Code, persistence.NullIfEmpty(lot.Description), lot.Status,
		lot.CreatedAt, lot.UpdatedAt,
		persistence.NullIfEmptyUUID(lot.CreatedBy), persistence.NullIfEmptyUUID(lot.UpdatedBy),
	)
	return persistence.Translate(err)
}

// Update persists the mutable lot fields.
func (r *importLotRepository) Update(ctx context.Context, lot *purchasing.ImportLot) error {
	const q = `UPDATE import_lots SET
		code = $1, description = $2, status = $3, updated_at = $4, updated_by = $5
		WHERE id = $6 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		lot.Code, persistence.NullIfEmpty(lot.Description), lot.Status,
		time.Now().UTC(), persistence.NullIfEmptyUUID(lot.UpdatedBy), lot.ID,
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

// SoftDelete marks the lot as deleted.
func (r *importLotRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`UPDATE import_lots SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`,
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

// GetByID loads a single lot without its members.
func (r *importLotRepository) GetByID(ctx context.Context, id uuid.UUID) (*purchasing.ImportLot, error) {
	const q = `SELECT ` + importLotColumns + ` FROM import_lots WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanImportLot(row)
}

// List returns the lots matching the filter.
func (r *importLotRepository) List(ctx context.Context, filter purchasing.ImportLotFilter) (repositories.Page[*purchasing.ImportLot], error) {
	var clauses = []string{"deleted_at IS NULL"}
	var args []any
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("(code LIKE $%d OR description LIKE $%d)", len(args)+1, len(args)+1))
		args = append(args, "%"+filter.Search+"%")
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)
	where := persistence.JoinClauses(clauses)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		"SELECT count(*) FROM import_lots WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*purchasing.ImportLot]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM import_lots WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
			importLotColumns, where, limitPos, offsetPos), args...)
	if err != nil {
		return repositories.Page[*purchasing.ImportLot]{}, persistence.Translate(err)
	}
	out := make([]*purchasing.ImportLot, 0, limit)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		lot, err := scanImportLotFromRows(row)
		if err != nil {
			return err
		}
		out = append(out, lot)
		return nil
	}); err != nil {
		return repositories.Page[*purchasing.ImportLot]{}, err
	}
	return repositories.Page[*purchasing.ImportLot]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

// NextCode returns the next "LOTE-" zero-padded sequence code for the
// current year.
func (r *importLotRepository) NextCode(ctx context.Context) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM import_lots WHERE code LIKE $1`,
		fmt.Sprintf("LOTE-%d-%%", year)).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("LOTE-%d-%05d", year, n+1), nil
}

// AddMembers links purchase orders to the lot (existing links are
// kept as-is).
func (r *importLotRepository) AddMembers(ctx context.Context, lotID uuid.UUID, purchaseIDs []uuid.UUID) error {
	for _, pid := range purchaseIDs {
		var exists int
		if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
			`SELECT COUNT(*) FROM import_lot_purchase_orders WHERE import_lot_id = $1 AND purchase_order_id = $2`,
			lotID, pid).Scan(&exists); err != nil {
			return persistence.Translate(err)
		}
		if exists > 0 {
			continue
		}
		_, err := persistence.Q(ctx, r.q).ExecContext(ctx,
			`INSERT INTO import_lot_purchase_orders (import_lot_id, purchase_order_id, added_at) VALUES ($1, $2, $3)`,
			lotID, pid, time.Now().UTC())
		if err != nil {
			return persistence.Translate(err)
		}
	}
	return nil
}

// RemoveMember unlinks a purchase order from the lot.
func (r *importLotRepository) RemoveMember(ctx context.Context, lotID, purchaseID uuid.UUID) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`DELETE FROM import_lot_purchase_orders WHERE import_lot_id = $1 AND purchase_order_id = $2`,
		lotID, purchaseID)
	return persistence.Translate(err)
}

// ListMembers returns the membership rows of a lot.
func (r *importLotRepository) ListMembers(ctx context.Context, lotID uuid.UUID) ([]*purchasing.ImportLotMember, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT import_lot_id, purchase_order_id, added_at
		 FROM import_lot_purchase_orders
		 WHERE import_lot_id = $1
		 ORDER BY added_at`, lotID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.ImportLotMember, 0, 8)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		m := &purchasing.ImportLotMember{}
		if err := row.Scan(&m.LotID, &m.PurchaseOrderID, &m.AddedAt); err != nil {
			return persistence.Translate(err)
		}
		out = append(out, m)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// ListPurchasesForLot returns the non-deleted purchase orders of a lot.
func (r *importLotRepository) ListPurchasesForLot(ctx context.Context, lotID uuid.UUID) ([]*purchasing.PurchaseOrder, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT po.`+purchaseColumns+`
		 FROM import_lot_purchase_orders m
		 JOIN purchase_orders po ON po.id = m.purchase_order_id
		 WHERE m.import_lot_id = $1 AND po.deleted_at IS NULL
		 ORDER BY po.order_date DESC, po.number DESC`, lotID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.PurchaseOrder, 0, 8)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		p, err := scanPurchaseOrderFromRows(row)
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

// LotForPurchase returns the lot that contains the order, if any.
func (r *importLotRepository) LotForPurchase(ctx context.Context, purchaseID uuid.UUID) (*purchasing.ImportLot, error) {
	const q = `SELECT l.` + importLotColumns + `
		FROM import_lots l
		JOIN import_lot_purchase_orders m ON m.import_lot_id = l.id
		WHERE m.purchase_order_id = $1 AND l.deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, purchaseID)
	return scanImportLot(row)
}

func scanImportLot(row *sql.Row) (*purchasing.ImportLot, error) {
	lot := &purchasing.ImportLot{}
	var (
		description sql.NullString
	deletedAt            sql.NullTime
		createdBy, updatedBy   sql.NullString
	)
	if err := persistence.ScanRow(row,
		&lot.ID, &lot.Code, &description, &lot.Status,
		&lot.CreatedAt, &lot.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, err
	}
	decodeImportLot(lot, description.String, deletedAt, createdBy.String, updatedBy.String)
	return lot, nil
}

func scanImportLotFromRows(rows *sql.Rows) (*purchasing.ImportLot, error) {
	lot := &purchasing.ImportLot{}
	var (
		description sql.NullString
	deletedAt            sql.NullTime
		createdBy, updatedBy   sql.NullString
	)
	if err := rows.Scan(
		&lot.ID, &lot.Code, &description, &lot.Status,
		&lot.CreatedAt, &lot.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	decodeImportLot(lot, description.String, deletedAt, createdBy.String, updatedBy.String)
	return lot, nil
}

func decodeImportLot(lot *purchasing.ImportLot, description string, deletedAt sql.NullTime, createdBy, updatedBy string) {
	lot.Description = description
	if deletedAt.Valid {
		t := deletedAt.Time
		lot.DeletedAt = &t
	}
	if createdBy != "" {
		id := persistence.ParseUUID(createdBy)
		lot.CreatedBy = &id
	}
	if updatedBy != "" {
		id := persistence.ParseUUID(updatedBy)
		lot.UpdatedBy = &id
	}
}

var _ purchasing.ImportLotRepository = (*importLotRepository)(nil)
