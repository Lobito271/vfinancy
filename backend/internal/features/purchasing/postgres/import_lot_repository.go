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

type importLotRepository struct {
	q persistence.Querier
}

// NewImportLotRepository returns the SQL import-lot repository.
func NewImportLotRepository(db *sql.DB) *importLotRepository {
	return &importLotRepository{q: persistence.FromDB(db)}
}

const importLotColumns = `
	id, company_id, code, description, status,
	created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *importLotRepository) Create(ctx context.Context, lot *purchasing.ImportLot) error {
	const q = `INSERT INTO import_lots (
		id, company_id, code, description, status,
		created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		lot.ID, lot.CompanyID, lot.Code, persistence.NullIfEmpty(lot.Description), lot.Status,
		lot.CreatedAt, lot.UpdatedAt,
		persistence.NullIfEmptyUUID(lot.CreatedBy), persistence.NullIfEmptyUUID(lot.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *importLotRepository) Update(ctx context.Context, lot *purchasing.ImportLot) error {
	const q = `UPDATE import_lots SET
		code = $1, description = $2, status = $3, updated_at = $4, updated_by = $5
		WHERE id = $6 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		lot.Code, persistence.NullIfEmpty(lot.Description), lot.Status, time.Now().UTC(),
		persistence.NullIfEmptyUUID(lot.UpdatedBy), lot.ID,
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

func (r *importLotRepository) GetByID(ctx context.Context, id uuid.UUID, loadMembers bool) (*purchasing.ImportLot, error) {
	const q = `SELECT ` + importLotColumns + ` FROM import_lots WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	lot, err := scanImportLot(row)
	if err != nil {
		return nil, err
	}
	if loadMembers {
		members, err := r.listMembers(ctx, id)
		if err != nil {
			return nil, err
		}
		lot.Members = members
	}
	return lot, nil
}

func (r *importLotRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM import_lots WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n)
	return n > 0, persistence.Translate(err)
}

func (r *importLotRepository) List(ctx context.Context, filter purchasing.ImportLotFilter) (repositories.Page[*purchasing.ImportLot], error) {
	var (
		clauses = []string{"deleted_at IS NULL"}
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
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

func (r *importLotRepository) GetNextCode(ctx context.Context, companyID uuid.UUID) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	like := fmt.Sprintf("LOT-%d-%%", year)
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM import_lots WHERE company_id = $1 AND code LIKE $2`,
		companyID, like).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("LOT-%d-%05d", year, n+1), nil
}

func (r *importLotRepository) AddMember(ctx context.Context, lotID, purchaseID uuid.UUID) error {
	var exists int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM import_lot_purchase_orders WHERE import_lot_id = $1 AND purchase_order_id = $2`,
		lotID, purchaseID).Scan(&exists); err != nil {
		return persistence.Translate(err)
	}
	if exists > 0 {
		return nil
	}
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`INSERT INTO import_lot_purchase_orders (import_lot_id, purchase_order_id, added_at) VALUES ($1, $2, $3)`,
		lotID, purchaseID, time.Now().UTC())
	return persistence.Translate(err)
}

func (r *importLotRepository) RemoveMember(ctx context.Context, lotID, purchaseID uuid.UUID) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx,
		`DELETE FROM import_lot_purchase_orders WHERE import_lot_id = $1 AND purchase_order_id = $2`,
		lotID, purchaseID)
	return persistence.Translate(err)
}

func (r *importLotRepository) MemberCount(ctx context.Context, lotID uuid.UUID) (int, error) {
	var n int
	err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM import_lot_purchase_orders WHERE import_lot_id = $1`, lotID).Scan(&n)
	return n, persistence.Translate(err)
}

func (r *importLotRepository) TotalUSD(ctx context.Context, lotID uuid.UUID) (valueobjects.Money, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT po.cost_usd FROM purchase_orders po
		 JOIN import_lot_purchase_orders m ON m.purchase_order_id = po.id
		 WHERE m.import_lot_id = $1 AND po.deleted_at IS NULL AND po.status <> 'cancelled'`, lotID)
	if err != nil {
		return valueobjects.Money{}, persistence.Translate(err)
	}
	total := valueobjects.Zero()
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		var s string
		if err := row.Scan(&s); err != nil {
			return err
		}
		m, err := valueobjects.MoneyFromString(s)
		if err != nil {
			return err
		}
		total = total.Add(m)
		return nil
	}); err != nil {
		return valueobjects.Money{}, err
	}
	return total, nil
}

func (r *importLotRepository) listMembers(ctx context.Context, lotID uuid.UUID) ([]*purchasing.ImportLotMember, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT import_lot_id, purchase_order_id, added_at FROM import_lot_purchase_orders WHERE import_lot_id = $1 ORDER BY added_at`, lotID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*purchasing.ImportLotMember, 0, 8)
	if err := persistence.ScanRows(rows, func(row *sql.Rows) error {
		m := &purchasing.ImportLotMember{}
		if err := row.Scan(&m.ImportLotID, &m.PurchaseID, &m.AddedAt); err != nil {
			return err
		}
		out = append(out, m)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func scanImportLot(row *sql.Row) (*purchasing.ImportLot, error) {
	lot := &purchasing.ImportLot{}
	var (
		description, deletedAt sql.NullString
		createdBy, updatedBy   sql.NullString
	)
	if err := persistence.ScanRow(row,
		&lot.ID, &lot.CompanyID, &lot.Code, &description, &lot.Status,
		&lot.CreatedAt, &lot.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, err
	}
	if description.Valid {
		lot.Description = description.String
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		lot.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		lot.UpdatedBy = &id
	}
	return lot, nil
}

func scanImportLotFromRows(rows *sql.Rows) (*purchasing.ImportLot, error) {
	lot := &purchasing.ImportLot{}
	var (
		description, deletedAt sql.NullString
		createdBy, updatedBy   sql.NullString
	)
	if err := rows.Scan(
		&lot.ID, &lot.CompanyID, &lot.Code, &description, &lot.Status,
		&lot.CreatedAt, &lot.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if description.Valid {
		lot.Description = description.String
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		lot.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		lot.UpdatedBy = &id
	}
	lot.Members = []*purchasing.ImportLotMember{}
	return lot, nil
}

var _ purchasing.ImportLotRepository = (*importLotRepository)(nil)
