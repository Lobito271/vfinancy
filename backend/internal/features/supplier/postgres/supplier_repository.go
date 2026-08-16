package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/supplier"
	"vfinancy/backend/infrastructure/persistence"
)

// supplierRepository is the SQL implementation of supplier.SupplierRepository.
type supplierRepository struct {
	q persistence.Querier
}

// NewSupplierRepository returns an auto-commit implementation.
func NewSupplierRepository(db *sql.DB) *supplierRepository {
	return &supplierRepository{q: persistence.FromDB(db)}
}

const supplierColumns = `
	id, company_id, document_type, document_number, business_name,
	trade_name, tax_id, is_international, default_currency, current_debt,
	payment_term_days, status, email, phone, address, is_active,
	created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *supplierRepository) Create(ctx context.Context, s *supplier.Supplier) error {
	const q = `INSERT INTO suppliers (
		id, company_id, document_type, document_number, business_name,
		trade_name, tax_id, is_international, default_currency, current_debt,
		payment_term_days, status, email, phone, address, is_active,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.ID, s.CompanyID, s.Document.Type(), s.Document.Number(),
		s.BusinessName.String(), persistence.NullIfEmptyFullName(s.TradeName),
		persistence.NullIfEmpty(s.TaxID), s.IsInternational, s.DefaultCurrency.String(),
		s.CurrentDebt.String(), s.PaymentTermDays, string(s.Status),
		s.Email.String(), s.Phone.String(), s.Address.String(), true,
		s.CreatedAt, s.UpdatedAt, persistence.NullIfZeroTime(s.DeletedAt), s.CreatedBy, s.UpdatedBy,
	)
	return persistence.Translate(err)
}

func (r *supplierRepository) Update(ctx context.Context, s *supplier.Supplier) error {
	const q = `UPDATE suppliers SET
		business_name = $1, trade_name = $2, tax_id = $3, is_international = $4,
		default_currency = $5, current_debt = $6, payment_term_days = $7,
		status = $8, email = $9, phone = $10, address = $11,
		updated_at = $12, updated_by = $13
	 WHERE id = $14 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		s.BusinessName.String(), persistence.NullIfEmptyFullName(s.TradeName),
		persistence.NullIfEmpty(s.TaxID), s.IsInternational, s.DefaultCurrency.String(),
		s.CurrentDebt.String(), s.PaymentTermDays, string(s.Status),
		s.Email.String(), s.Phone.String(), s.Address.String(),
		time.Now().UTC(), s.UpdatedBy, s.ID,
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

func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM suppliers WHERE id = $1`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *supplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*supplier.Supplier, error) {
	q := `SELECT ` + supplierColumns + ` FROM suppliers WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	s, err := scanSupplier(row)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *supplierRepository) GetByDocument(ctx context.Context, companyID uuid.UUID, documentNumber string) (*supplier.Supplier, error) {
	q := `SELECT ` + supplierColumns + `
		FROM suppliers
		WHERE company_id = $1 AND document_number = $2 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, documentNumber)
	s, err := scanSupplier(row)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *supplierRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, `SELECT 1 FROM suppliers WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if persistence.IsPgNoRows(err) {
			return false, nil
		}
		return false, persistence.Translate(err)
	}
	return n == 1, nil
}

func (r *supplierRepository) List(ctx context.Context, filter supplier.SupplierFilter) (repositories.Page[*supplier.Supplier], error) {
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
		clauses = append(clauses, fmt.Sprintf(
			"(LOWER(business_name) LIKE LOWER($%d) OR document_number LIKE $%d OR LOWER(COALESCE(trade_name, '')) LIKE LOWER($%d))",
			len(args)+1, len(args)+2, len(args)+3,
		))
		like := "%" + filter.Search + "%"
		args = append(args, like, like, like)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := persistence.JoinClauses(clauses)
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM suppliers WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*supplier.Supplier]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM suppliers WHERE %s ORDER BY business_name LIMIT $%d OFFSET $%d",
			supplierColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*supplier.Supplier]{}, persistence.Translate(err)
	}
	out := make([]*supplier.Supplier, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		s, err := scanSupplierFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, s)
		return nil
	}); err != nil {
		return repositories.Page[*supplier.Supplier]{}, err
	}
	return repositories.Page[*supplier.Supplier]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *supplierRepository) GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error) {
	var bal string
	err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT current_debt FROM suppliers WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&bal)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return "", repositories.ErrNotFound
		}
		return "", persistence.Translate(err)
	}
	return bal, nil
}

// --- scanning helpers ---

// scanSupplier is the *sql.Row variant.
func scanSupplier(row *sql.Row) (*supplier.Supplier, error) {
	s := &supplier.Supplier{}
	var (
		docType, docNum, businessName, currency, status                  string
		tradeName, taxID, email, phone, addr, createdBy, updatedBy       sql.NullString
		deletedAt                                                        sql.NullTime
		currentDebt                                                      string
		paymentTerm                                                      int
		isInternational, isActive                                        bool
	)
	err := persistence.ScanRow(row,
		&s.ID, &s.CompanyID, &docType, &docNum, &businessName,
		&tradeName, &taxID, &isInternational, &currency, &currentDebt,
		&paymentTerm, &status, &email, &phone, &addr, &isActive,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	fillSupplier(s, docType, docNum, businessName, tradeName, taxID, isInternational, currency, currentDebt, paymentTerm, status, email, phone, addr, deletedAt, createdBy, updatedBy)
	return s, nil
}

// scanSupplierFromRows is the *sql.Rows variant.
func scanSupplierFromRows(rows *sql.Rows) (*supplier.Supplier, error) {
	s := &supplier.Supplier{}
	var (
		docType, docNum, businessName, currency, status                  string
		tradeName, taxID, email, phone, addr, createdBy, updatedBy       sql.NullString
		deletedAt                                                        sql.NullTime
		currentDebt                                                      string
		paymentTerm                                                      int
		isInternational, isActive                                        bool
	)
	if err := rows.Scan(
		&s.ID, &s.CompanyID, &docType, &docNum, &businessName,
		&tradeName, &taxID, &isInternational, &currency, &currentDebt,
		&paymentTerm, &status, &email, &phone, &addr, &isActive,
		&s.CreatedAt, &s.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	fillSupplier(s, docType, docNum, businessName, tradeName, taxID, isInternational, currency, currentDebt, paymentTerm, status, email, phone, addr, deletedAt, createdBy, updatedBy)
	return s, nil
}

// fillSupplier converts the raw scanned columns into value objects. It is
// shared by both scan variants.
func fillSupplier(s *supplier.Supplier,
	docType, docNum, businessName string, tradeName, taxID sql.NullString, isInternational bool,
	currency, currentDebt string, paymentTerm int, status string, email, phone, addr sql.NullString,
	deletedAt sql.NullTime, createdBy, updatedBy sql.NullString,
) {
	doc, err := persistence.ParseDocument(docType, docNum)
	if err == nil {
		s.Document = doc
	}
	s.BusinessName = persistence.ParseFullName(businessName)
	if tradeName.Valid {
		s.TradeName = persistence.ParseFullName(tradeName.String)
	}
	if taxID.Valid {
		s.TaxID = taxID.String
	}
	s.IsInternational = isInternational
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
	s.Status = persistence.ParseSupplierStatus(status)
	if email.Valid {
		s.Email = persistence.ParseEmail(email.String)
	}
	if phone.Valid {
		s.Phone = persistence.ParsePhone(phone.String)
	}
	if addr.Valid {
		s.Address = persistence.ParseAddress(addr.String)
	}
	if v, err := persistence.ParseMoney(currentDebt); err == nil {
		s.CurrentDebt = v
	}
	s.PaymentTermDays = paymentTerm
	if cc, err := valueobjects.NewCurrencyCode(currency); err == nil {
		s.DefaultCurrency = cc
	}
}
