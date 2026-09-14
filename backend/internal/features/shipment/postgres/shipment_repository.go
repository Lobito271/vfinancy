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
	"vfinancy/backend/internal/features/shipment"
)

type shipmentRepository struct {
	q persistence.Querier
}

// NewShipmentRepository returns a ShipmentRepository backed by the
// given database handle.
func NewShipmentRepository(db *sql.DB) *shipmentRepository {
	return &shipmentRepository{q: persistence.FromDB(db)}
}

const shipmentColumns = `
	id, code, sale_id, customer_id, description, notes, status,
	created_at, updated_at, deleted_at, created_by, updated_by
`

const shipmentInsert = `INSERT INTO shipments (
	id, code, sale_id, customer_id, description, notes, status,
	created_at, updated_at, deleted_at, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

func (r *shipmentRepository) Create(ctx context.Context, s *shipment.Shipment) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, shipmentInsert,
		s.ID, s.Code,
		persistence.NullIfEmptyUUID(s.SaleID), persistence.NullIfEmptyUUID(s.CustomerID),
		persistence.NullIfEmpty(s.Description), persistence.NullIfEmpty(s.Notes), s.Status.String(),
		s.CreatedAt, s.UpdatedAt, persistence.NullIfZeroTime(s.DeletedAt),
		persistence.NullIfEmptyUUID(s.CreatedBy), persistence.NullIfEmptyUUID(s.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *shipmentRepository) Update(ctx context.Context, s *shipment.Shipment) error {
	const q = `UPDATE shipments SET
		sale_id = $1, customer_id = $2, description = $3, notes = $4,
		status = $5, updated_at = $6, updated_by = $7
	 WHERE id = $8 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		persistence.NullIfEmptyUUID(s.SaleID), persistence.NullIfEmptyUUID(s.CustomerID),
		persistence.NullIfEmpty(s.Description), persistence.NullIfEmpty(s.Notes),
		s.Status.String(), s.UpdatedAt, persistence.NullIfEmptyUUID(s.UpdatedBy),
		s.ID,
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

func (r *shipmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE shipments SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, now, id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *shipmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*shipment.Shipment, error) {
	q := `SELECT ` + shipmentColumns + ` FROM shipments WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanShipment(row)
}

// NextCode returns the next four-digit sequence code. The count-based
// assignment never reuses a code while rows stay soft-deleted.
// ponytail: exhausts at 9999; switch to random-with-retry if ever needed.
func (r *shipmentRepository) NextCode(ctx context.Context) (string, error) {
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM shipments`).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("%04d", n+1), nil
}

func (r *shipmentRepository) List(ctx context.Context, filter shipment.ShipmentFilter) (repositories.Page[*shipment.Shipment], error) {
	var clauses []string
	clauses = append(clauses, "deleted_at IS NULL")
	var args []any
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("(code LIKE $%d OR LOWER(description) LIKE LOWER($%d))", len(args)+1, len(args)+2))
		like := "%" + filter.Search + "%"
		args = append(args, like, like)
	}
	where := " WHERE " + persistence.JoinClauses(clauses)
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 1000)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM shipments`+where, args...).Scan(&total); err != nil {
		return repositories.Page[*shipment.Shipment]{}, persistence.Translate(err)
	}

	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM shipments%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
			shipmentColumns, where, len(args)+1, len(args)+2),
		append(args, limit, offset)...)
	if err != nil {
		return repositories.Page[*shipment.Shipment]{}, persistence.Translate(err)
	}
	out := make([]*shipment.Shipment, 0)
	if err := persistence.ScanRows(rows, func(rows *sql.Rows) error {
		s, err := scanShipmentFromRows(rows)
		if err != nil {
			return err
		}
		out = append(out, s)
		return nil
	}); err != nil {
		return repositories.Page[*shipment.Shipment]{}, err
	}
	return repositories.Page[*shipment.Shipment]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func scanShipment(row *sql.Row) (*shipment.Shipment, error) {
	s := &shipment.Shipment{}
	var (
		saleID, customerID, createdBy, updatedBy sql.NullString
		description, notes                       sql.NullString
		deletedAt                                sql.NullTime
		status                                   string
	)
	err := persistence.ScanRow(row,
		&s.ID, &s.Code, &saleID, &customerID, &description, &notes, &status,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	decodeShipment(s, saleID, customerID, description, notes, status, createdBy, updatedBy, deletedAt)
	return s, nil
}

func scanShipmentFromRows(rows *sql.Rows) (*shipment.Shipment, error) {
	s := &shipment.Shipment{}
	var (
		saleID, customerID, createdBy, updatedBy sql.NullString
		description, notes                       sql.NullString
		deletedAt                                sql.NullTime
		status                                   string
	)
	if err := rows.Scan(
		&s.ID, &s.Code, &saleID, &customerID, &description, &notes, &status,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	decodeShipment(s, saleID, customerID, description, notes, status, createdBy, updatedBy, deletedAt)
	return s, nil
}

func decodeShipment(s *shipment.Shipment, saleID, customerID, description, notes sql.NullString, status string, createdBy, updatedBy sql.NullString, deletedAt sql.NullTime) {
	if saleID.Valid {
		id := persistence.ParseUUID(saleID.String)
		s.SaleID = &id
	}
	if customerID.Valid {
		id := persistence.ParseUUID(customerID.String)
		s.CustomerID = &id
	}
	s.Description = description.String
	s.Notes = notes.String
	s.Status = enums.ShipmentStatus(status)
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		s.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		s.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		s.DeletedAt = &t
	}
}

var _ shipment.ShipmentRepository = (*shipmentRepository)(nil)
