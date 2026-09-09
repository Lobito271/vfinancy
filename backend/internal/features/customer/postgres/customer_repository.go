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
	"vfinancy/backend/internal/features/customer"
)

// customerRepository is the SQL implementation of
// customer.CustomerRepository. The same struct can run against either
// *sql.DB (auto-commit) or *sql.Tx (transaction-bound); the Querier
// field is set by the constructor.
type customerRepository struct {
	q persistence.Querier
}

// NewCustomerRepository returns an auto-commit implementation.
func NewCustomerRepository(db *sql.DB) *customerRepository {
	return &customerRepository{q: persistence.FromDB(db)}
}

const customerColumns = `
	id, document_type, document_number, business_name,
	email, phone, address, current_debt, status,
	created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *customerRepository) Create(ctx context.Context, c *customer.Customer) error {
	const q = `INSERT INTO customers (
		id, document_type, document_number, business_name,
		email, phone, address, current_debt, status,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		c.ID,
		persistence.NullIfEmpty(c.DocumentType.String()),
		persistence.NullIfEmpty(c.DocumentNumber.Number()),
		c.BusinessName.String(),
		persistence.NullIfEmpty(c.Email.String()),
		persistence.NullIfEmpty(c.Phone.String()),
		persistence.NullIfEmpty(c.Address.String()),
		c.CurrentDebt.String(),
		string(c.Status),
		c.CreatedAt, c.UpdatedAt,
		persistence.NullIfZeroTime(c.DeletedAt),
		persistence.NullIfEmpty(c.CreatedBy),
		persistence.NullIfEmpty(c.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *customerRepository) Update(ctx context.Context, c *customer.Customer) error {
	const q = `UPDATE customers SET
		document_type = $1, document_number = $2, business_name = $3,
		email = $4, phone = $5, address = $6, current_debt = $7, status = $8,
		updated_at = $9, updated_by = $10
	 WHERE id = $11 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		persistence.NullIfEmpty(c.DocumentType.String()),
		persistence.NullIfEmpty(c.DocumentNumber.Number()),
		c.BusinessName.String(),
		persistence.NullIfEmpty(c.Email.String()),
		persistence.NullIfEmpty(c.Phone.String()),
		persistence.NullIfEmpty(c.Address.String()),
		c.CurrentDebt.String(),
		string(c.Status),
		time.Now().UTC(),
		persistence.NullIfEmpty(c.UpdatedBy),
		c.ID,
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

func (r *customerRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE customers SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, time.Now().UTC(), id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*customer.Customer, error) {
	q := `SELECT ` + customerColumns + ` FROM customers WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanCustomer(row)
}

func (r *customerRepository) GetByDocument(ctx context.Context, documentType, documentNumber string) (*customer.Customer, error) {
	q := `SELECT ` + customerColumns + ` FROM customers
		WHERE document_type = $1 AND document_number = $2 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, persistence.NullIfEmpty(documentType), persistence.NullIfEmpty(documentNumber))
	return scanCustomer(row)
}

func (r *customerRepository) GetByDescription(ctx context.Context, name string) (*customer.Customer, error) {
	q := `SELECT ` + customerColumns + ` FROM customers WHERE business_name = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, name)
	return scanCustomer(row)
}

func (r *customerRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, `SELECT 1 FROM customers WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if persistence.IsPgNoRows(err) {
			return false, nil
		}
		return false, persistence.Translate(err)
	}
	return n == 1, nil
}

func (r *customerRepository) List(ctx context.Context, filter customer.CustomerFilter) (repositories.Page[*customer.Customer], error) {
	var clauses []string
	if !filter.IncludeDeleted {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	var args []any
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf(
			"(LOWER(business_name) LIKE LOWER($%d) OR document_number LIKE $%d)",
			len(args)+1, len(args)+2,
		))
		like := "%" + filter.Search + "%"
		args = append(args, like, like)
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + persistence.JoinClauses(clauses)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 1000)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM customers"+where, args...).Scan(&total); err != nil {
		return repositories.Page[*customer.Customer]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM customers%s ORDER BY business_name LIMIT $%d OFFSET $%d",
			customerColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*customer.Customer]{}, persistence.Translate(err)
	}
	out := make([]*customer.Customer, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		c, err := scanCustomerFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, c)
		return nil
	}); err != nil {
		return repositories.Page[*customer.Customer]{}, err
	}
	return repositories.Page[*customer.Customer]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *customerRepository) GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error) {
	var bal string
	err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT current_debt FROM customers WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&bal)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return "", repositories.ErrNotFound
		}
		return "", persistence.Translate(err)
	}
	return bal, nil
}

// --- scanning helpers ---

// customerRow is the raw scan target shared by the *sql.Row and
// *sql.Rows variants.
type customerRow struct {
	docType, docNum      sql.NullString
	businessName         string
	email, phone, addr   sql.NullString
	currentDebt          string
	status               string
	deletedAt            sql.NullTime
	createdBy, updatedBy sql.NullString
}

// scanCustomer is the *sql.Row variant. NotFound rows are translated
// to repositories.ErrNotFound by persistence.ScanRow.
func scanCustomer(row *sql.Row) (*customer.Customer, error) {
	c := &customer.Customer{}
	cr := &customerRow{}
	if err := persistence.ScanRow(row,
		&c.ID, &cr.docType, &cr.docNum, &cr.businessName,
		&cr.email, &cr.phone, &cr.addr, &cr.currentDebt, &cr.status,
		&c.CreatedAt, &c.UpdatedAt, &cr.deletedAt, &cr.createdBy, &cr.updatedBy,
	); err != nil {
		return nil, err
	}
	return cr.fill(c)
}

// scanCustomerFromRows is the *sql.Rows variant.
func scanCustomerFromRows(rows *sql.Rows) (*customer.Customer, error) {
	c := &customer.Customer{}
	cr := &customerRow{}
	if err := rows.Scan(
		&c.ID, &cr.docType, &cr.docNum, &cr.businessName,
		&cr.email, &cr.phone, &cr.addr, &cr.currentDebt, &cr.status,
		&c.CreatedAt, &c.UpdatedAt, &cr.deletedAt, &cr.createdBy, &cr.updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	return cr.fill(c)
}

// fill converts the raw scanned columns into the entity's value
// objects.
func (cr *customerRow) fill(c *customer.Customer) (*customer.Customer, error) {
	if cr.docType.Valid {
		c.DocumentType = enums.DocumentType(cr.docType.String)
	}
	if cr.docNum.Valid {
		dn, err := persistence.ParseDocument(c.DocumentType.String(), cr.docNum.String)
		if err != nil {
			return nil, err
		}
		c.DocumentNumber = dn
	}
	c.BusinessName = persistence.ParseFullName(cr.businessName)
	c.Email = persistence.ParseEmail(cr.email.String)
	c.Phone = persistence.ParsePhone(cr.phone.String)
	c.Address = persistence.ParseAddress(cr.addr.String)
	debt, err := persistence.ParseMoney(cr.currentDebt)
	if err != nil {
		return nil, err
	}
	c.CurrentDebt = debt
	c.Status = persistence.ParseCustomerStatus(cr.status)
	if cr.deletedAt.Valid {
		t := cr.deletedAt.Time
		c.DeletedAt = &t
	}
	if cr.createdBy.Valid {
		c.CreatedBy = cr.createdBy.String
	}
	if cr.updatedBy.Valid {
		c.UpdatedBy = cr.updatedBy.String
	}
	return c, nil
}
