package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/masterdata"
	"vfinancy/backend/internal/domain/repositories"
)

// customerRepository is the PostgreSQL implementation of
// repositories.CustomerRepository. The same struct can run against
// either *sql.DB (auto-commit) or *sql.Tx (transaction-bound); the
// Querier field is set by the constructor.
type customerRepository struct {
	q Querier
}

// newCustomerRepository returns a transaction-bound (auto-commit)
// implementation.
func newCustomerRepository(db *sql.DB) *customerRepository {
	return &customerRepository{q: &dbBox{db: db}}
}

// newCustomerRepositoryTx returns a transaction-bound implementation.
func newCustomerRepositoryTx(tx *sql.Tx) *customerRepository {
	return &customerRepository{q: &txBox{tx: tx}}
}

const customerColumns = `
	id, company_id, default_branch_id, document_type, document_number,
	business_name, trade_name, tax_category, credit_limit, current_debt,
	payment_term_days, status, blocked_reason, email, phone, address,
	created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *customerRepository) Create(ctx context.Context, c *masterdata.Customer) error {
	const q = `INSERT INTO customers (
		id, company_id, default_branch_id, document_type, document_number,
		business_name, trade_name, tax_category, credit_limit, current_debt,
		payment_term_days, status, blocked_reason, email, phone, address,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`
	_, err := r.q.ExecContext(ctx, q,
		c.ID, c.CompanyID, c.BranchID, c.Document.Type(), c.Document.Number(),
		c.BusinessName.String(), nullIfEmptyFullName(c.TradeName), string(c.TaxCategory),
		c.CreditLimit.String(), c.CurrentDebt.String(),
		c.PaymentTermDays, string(c.Status), nullIfEmpty(c.BlockedReason),
		c.Email.String(), c.Phone.String(), c.Address.String(),
		c.CreatedAt, c.UpdatedAt, nullIfZeroTime(c.DeletedAt), c.CreatedBy, c.UpdatedBy,
	)
	return Translate(err)
}

func (r *customerRepository) Update(ctx context.Context, c *masterdata.Customer) error {
	const q = `UPDATE customers SET
		default_branch_id = $1, document_type = $2, document_number = $3,
		business_name = $4, trade_name = $5, tax_category = $6, credit_limit = $7,
		current_debt = $8, payment_term_days = $9, status = $10, blocked_reason = $11,
		email = $12, phone = $13, address = $14, updated_at = $15, updated_by = $16
	 WHERE id = $17 AND deleted_at IS NULL`
	res, err := r.q.ExecContext(ctx, q,
		c.BranchID, c.Document.Type(), c.Document.Number(),
		c.BusinessName.String(), nullIfEmptyFullName(c.TradeName), string(c.TaxCategory),
		c.CreditLimit.String(), c.CurrentDebt.String(),
		c.PaymentTermDays, string(c.Status), nullIfEmpty(c.BlockedReason),
		c.Email.String(), c.Phone.String(), c.Address.String(),
		time.Now().UTC(), c.UpdatedBy, c.ID,
	)
	if err != nil {
		return Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE customers SET deleted_at = $1, updated_at = $2, is_active = FALSE WHERE id = $3 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := r.q.ExecContext(ctx, q, now, now, id)
	if err != nil {
		return Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Customer, error) {
	q := `SELECT ` + customerColumns + ` FROM customers WHERE id = $1 AND deleted_at IS NULL`
	row := r.q.QueryRowContext(ctx, q, id)
	c, err := scanCustomer(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *customerRepository) GetByDocument(ctx context.Context, companyID uuid.UUID, documentType, documentNumber string) (*masterdata.Customer, error) {
	q := `SELECT ` + customerColumns + `
		FROM customers
		WHERE company_id = $1 AND document_type = $2 AND document_number = $3
		  AND deleted_at IS NULL`
	row := r.q.QueryRowContext(ctx, q, companyID, documentType, documentNumber)
	c, err := scanCustomer(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *customerRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := r.q.QueryRowContext(ctx, `SELECT 1 FROM customers WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if isPgNoRows(err) {
			return false, nil
		}
		return false, Translate(err)
	}
	return n == 1, nil
}

func (r *customerRepository) List(ctx context.Context, filter repositories.CustomerFilter) (repositories.Page[*masterdata.Customer], error) {
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
	if filter.BranchID != nil {
		clauses = append(clauses, fmt.Sprintf("default_branch_id = $%d", len(args)+1))
		args = append(args, *filter.BranchID)
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf(
			"(LOWER(business_name) LIKE LOWER($%d) OR document_number LIKE $%d OR LOWER(COALESCE(trade_name, '')) LIKE LOWER($%d))",
			len(args)+1, len(args)+2, len(args)+3,
		))
		like := "%" + filter.Search + "%"
		args = append(args, like, like, like)
	}
	limit, offset := limitOffset(filter.PageRequest, 25, 200)

	where := joinClauses(clauses)
	var total int
	if err := r.q.QueryRowContext(ctx, "SELECT count(*) FROM customers WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*masterdata.Customer]{}, Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := r.q.QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM customers WHERE %s ORDER BY business_name LIMIT $%d OFFSET $%d",
			customerColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*masterdata.Customer]{}, Translate(err)
	}
	out := make([]*masterdata.Customer, 0, limit)
	if err := scanRows(rows, func(r *sql.Rows) error {
		c, err := scanCustomerFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, c)
		return nil
	}); err != nil {
		return repositories.Page[*masterdata.Customer]{}, err
	}
	return repositories.Page[*masterdata.Customer]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *customerRepository) GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error) {
	var bal string
	err := r.q.QueryRowContext(ctx,
		`SELECT current_debt FROM customers WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&bal)
	if err != nil {
		if isPgNoRows(err) {
			return "", repositories.ErrNotFound
		}
		return "", Translate(err)
	}
	return bal, nil
}

// --- scanning helpers ---

// scanCustomer is the *sql.Row variant. It allocates one Customer
// and returns it.
func scanCustomer(row *sql.Row) (*masterdata.Customer, error) {
	c := &masterdata.Customer{}
	var (
		docType, docNum, businessName, taxCat, status, email, phone, addr          string
		tradeName                                                                            sql.NullString
		blocked                                                                              sql.NullString
		branchID, createdBy, updatedBy                                                      sql.NullString
		deletedAt                                                                             sql.NullTime
		creditLimit, currentDebt                                                             string
		paymentTerm                                                                           int
	)
	err := scanRow(row,
		&c.ID, &c.CompanyID, &branchID, &docType, &docNum,
		&businessName, &tradeName, &taxCat, &creditLimit, &currentDebt,
		&paymentTerm, &status, &blocked, &email, &phone, &addr,
		&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	dn, err := masterdataParseDocument(docType, docNum)
	if err != nil {
		return nil, err
	}
	c.Document = dn
	c.BusinessName = masterdataParseFullName(businessName)
	if tradeName.Valid {
		c.TradeName = masterdataParseFullName(tradeName.String)
	}
	if branchID.Valid {
		id := masterdataParseUUID(branchID.String)
		c.BranchID = &id
	}
	if createdBy.Valid {
		id := masterdataParseUUID(createdBy.String)
		c.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := masterdataParseUUID(updatedBy.String)
		c.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		c.DeletedAt = &t
	}
	c.TaxCategory = masterdataParseTaxCategory(taxCat)
	c.Status = masterdataParseCustomerStatus(status)
	if blocked.Valid {
		c.BlockedReason = blocked.String
	}
	c.Email = masterdataParseEmail(email)
	c.Email = masterdataParseEmail(email)
	c.Phone = masterdataParsePhone(phone)
	c.Address = masterdataParseAddress(addr)
	if v, err := masterdataParseMoney(creditLimit); err != nil {
		return nil, err
	} else {
		c.CreditLimit = v
	}
	if v, err := masterdataParseMoney(currentDebt); err != nil {
		return nil, err
	} else {
		c.CurrentDebt = v
	}
	c.PaymentTermDays = paymentTerm
	return c, nil
}

func scanCustomerFromRows(rows *sql.Rows) (*masterdata.Customer, error) {
	c := &masterdata.Customer{}
	var (
		docType, docNum, businessName, taxCat, status, email, phone, addr          string
		tradeName                                                                            sql.NullString
		blocked                                                                              sql.NullString
		branchID, createdBy, updatedBy                                                      sql.NullString
		deletedAt                                                                             sql.NullTime
		creditLimit, currentDebt                                                             string
		paymentTerm                                                                           int
	)
	if err := rows.Scan(
		&c.ID, &c.CompanyID, &branchID, &docType, &docNum,
		&businessName, &tradeName, &taxCat, &creditLimit, &currentDebt,
		&paymentTerm, &status, &blocked, &email, &phone, &addr,
		&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, Translate(err)
	}
	dn, err := masterdataParseDocument(docType, docNum)
	if err != nil {
		return nil, err
	}
	c.Document = dn
	c.BusinessName = masterdataParseFullName(businessName)
	if tradeName.Valid {
		c.TradeName = masterdataParseFullName(tradeName.String)
	}
	if branchID.Valid {
		id := masterdataParseUUID(branchID.String)
		c.BranchID = &id
	}
	if createdBy.Valid {
		id := masterdataParseUUID(createdBy.String)
		c.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := masterdataParseUUID(updatedBy.String)
		c.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		c.DeletedAt = &t
	}
	c.TaxCategory = masterdataParseTaxCategory(taxCat)
	c.Status = masterdataParseCustomerStatus(status)
	if blocked.Valid {
		c.BlockedReason = blocked.String
	}
	c.Email = masterdataParseEmail(email)
	c.Phone = masterdataParsePhone(phone)
	c.Address = masterdataParseAddress(addr)
	if v, err := masterdataParseMoney(creditLimit); err != nil {
		return nil, err
	} else {
		c.CreditLimit = v
	}
	if v, err := masterdataParseMoney(currentDebt); err != nil {
		return nil, err
	} else {
		c.CurrentDebt = v
	}
	c.PaymentTermDays = paymentTerm
	return c, nil
}
