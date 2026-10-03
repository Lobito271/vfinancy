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

// shipmentListSelect is the single projection every read goes through:
// the shipment plus its originating sale number, joined so the "Venta"
// column still resolves for a soft-deleted sale.
const shipmentListSelect = `
	SELECT s.id, s.code, s.security_code, s.sale_id, s.customer_id, s.location,
	s.shipment_date, s.delivered_at, s.description, s.notes, s.status,
	s.created_at, s.updated_at, s.deleted_at,
	COALESCE(sa.number, '')
	FROM shipments s
	LEFT JOIN sales sa ON sa.id = s.sale_id
`

const shipmentInsert = `INSERT INTO shipments (
	id, code, security_code, sale_id, customer_id, location,
	shipment_date, delivered_at, description, notes, status,
	created_at, updated_at, deleted_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

func (r *shipmentRepository) Create(ctx context.Context, s *shipment.Shipment) error {
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, shipmentInsert,
		s.ID, s.Code, s.SecurityCode, s.SaleID,
		persistence.NullIfEmptyUUID(s.CustomerID), persistence.NullIfEmpty(s.Location),
		persistence.NullIfZeroTime(s.ShipmentDate), persistence.NullIfZeroTime(s.DeliveredAt),
		persistence.NullIfEmpty(s.Description), persistence.NullIfEmpty(s.Notes), s.Status.String(),
		s.CreatedAt, s.UpdatedAt, persistence.NullIfZeroTime(s.DeletedAt),
	)
	return persistence.Translate(err)
}

// Update persists the mutable fields. Code, SecurityCode and SaleID are
// assigned at creation and deliberately absent from the SET clause.
func (r *shipmentRepository) Update(ctx context.Context, s *shipment.Shipment) error {
	const q = `UPDATE shipments SET
	customer_id = $1, location = $2, shipment_date = $3, delivered_at = $4,
	description = $5, notes = $6, status = $7, updated_at = $8
	WHERE id = $9 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		persistence.NullIfEmptyUUID(s.CustomerID), persistence.NullIfEmpty(s.Location),
		persistence.NullIfZeroTime(s.ShipmentDate), persistence.NullIfZeroTime(s.DeliveredAt),
		persistence.NullIfEmpty(s.Description), persistence.NullIfEmpty(s.Notes),
		s.Status.String(), s.UpdatedAt,
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
	q := shipmentListSelect + ` WHERE s.id = $1 AND s.deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanShipment(row)
}

// NextCode returns the next year-scoped shipment code, ENV-<year>-<seq>.
func (r *shipmentRepository) NextCode(ctx context.Context) (string, error) {
	year := time.Now().UTC().Year()
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM shipments WHERE code LIKE $1`,
		fmt.Sprintf("ENV-%d-%%", year)).Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("ENV-%d-%05d", year, n+1), nil
}

func (r *shipmentRepository) List(ctx context.Context, filter shipment.ShipmentFilter) (repositories.Page[*shipment.Shipment], error) {
	var clauses []string
	clauses = append(clauses, "s.deleted_at IS NULL")
	var args []any
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("s.status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf(
			"(s.code LIKE $%d OR s.security_code LIKE $%d OR LOWER(s.description) LIKE LOWER($%d))",
			len(args)+1, len(args)+2, len(args)+3))
		like := "%" + filter.Search + "%"
		args = append(args, like, like, like)
	}
	where := " WHERE " + persistence.JoinClauses(clauses)
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 1000)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT count(*) FROM shipments s`+where, args...).Scan(&total); err != nil {
		return repositories.Page[*shipment.Shipment]{}, persistence.Translate(err)
	}

	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		shipmentListSelect+fmt.Sprintf("%s ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d",
			where, len(args)+1, len(args)+2),
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

type shipmentScan struct {
	customerID, location, description, notes, saleNumber sql.NullString
	shipmentDate, deliveredAt, deletedAt                sql.NullTime
	status                                              string
}

func (s *shipmentScan) dests(shp *shipment.Shipment) []any {
	return []any{
		&shp.ID, &shp.Code, &shp.SecurityCode, &shp.SaleID, &s.customerID, &s.location,
		&s.shipmentDate, &s.deliveredAt, &s.description, &s.notes, &s.status,
		&shp.CreatedAt, &shp.UpdatedAt, &s.deletedAt, &s.saleNumber,
	}
}

func (s *shipmentScan) decode(shp *shipment.Shipment) {
	if s.customerID.Valid {
		id := persistence.ParseUUID(s.customerID.String)
		shp.CustomerID = &id
	}
	shp.Location = s.location.String
	shp.Description = s.description.String
	shp.Notes = s.notes.String
	shp.SaleNumber = s.saleNumber.String
	shp.Status = enums.ShipmentStatus(s.status)
	if s.shipmentDate.Valid {
		t := s.shipmentDate.Time
		shp.ShipmentDate = &t
	}
	if s.deliveredAt.Valid {
		t := s.deliveredAt.Time
		shp.DeliveredAt = &t
	}
	if s.deletedAt.Valid {
		t := s.deletedAt.Time
		shp.DeletedAt = &t
	}
}

func scanShipment(row *sql.Row) (*shipment.Shipment, error) {
	shp := &shipment.Shipment{}
	var sc shipmentScan
	if err := persistence.ScanRow(row, sc.dests(shp)...); err != nil {
		return nil, err
	}
	sc.decode(shp)
	return shp, nil
}

func scanShipmentFromRows(rows *sql.Rows) (*shipment.Shipment, error) {
	shp := &shipment.Shipment{}
	var sc shipmentScan
	if err := rows.Scan(sc.dests(shp)...); err != nil {
		return nil, persistence.Translate(err)
	}
	sc.decode(shp)
	return shp, nil
}

var _ shipment.ShipmentRepository = (*shipmentRepository)(nil)
