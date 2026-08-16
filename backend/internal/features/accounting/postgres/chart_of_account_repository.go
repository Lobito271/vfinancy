package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/accounting"
)

type chartOfAccountRepository struct {
	q persistence.Querier
}

func NewChartOfAccountRepository(db *sql.DB) *chartOfAccountRepository {
	return &chartOfAccountRepository{q: persistence.FromDB(db)}
}

const chartOfAccountColumns = `
	id, company_id, code, name, type, parent_id, path, depth,
	is_active, allows_movement, description, created_at, updated_at, created_by, updated_by
`

func (r *chartOfAccountRepository) Create(ctx context.Context, a *accounting.ChartOfAccount) error {
	const q = `INSERT INTO chart_of_accounts (
		id, company_id, code, name, type, parent_id, path, depth,
		is_active, allows_movement, description, created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		a.ID, a.CompanyID, a.Code.String(), a.Name, a.Type.String(),
		persistence.NullIfEmptyUUID(a.ParentID), a.Path, a.Depth,
		a.IsActive, a.AllowsMovement, persistence.NullIfEmpty(a.Description),
		a.CreatedAt, a.UpdatedAt,
		persistence.NullIfEmptyUUID(a.CreatedBy), persistence.NullIfEmptyUUID(a.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *chartOfAccountRepository) Update(ctx context.Context, a *accounting.ChartOfAccount) error {
	const q = `UPDATE chart_of_accounts SET
		code = $1, name = $2, type = $3, parent_id = $4, path = $5, depth = $6,
		is_active = $7, allows_movement = $8, description = $9, updated_at = $10, updated_by = $11
	 WHERE id = $12`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		a.Code.String(), a.Name, a.Type.String(), persistence.NullIfEmptyUUID(a.ParentID),
		a.Path, a.Depth, a.IsActive, a.AllowsMovement, persistence.NullIfEmpty(a.Description),
		a.UpdatedAt, persistence.NullIfEmptyUUID(a.UpdatedBy), a.ID,
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

func (r *chartOfAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE chart_of_accounts SET is_active = FALSE, updated_at = $2 WHERE id = $1`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, id, time.Now().UTC())
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *chartOfAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*accounting.ChartOfAccount, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+chartOfAccountColumns+` FROM chart_of_accounts WHERE id = $1`, id)
	a, err := scanChartOfAccount(row)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *chartOfAccountRepository) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*accounting.ChartOfAccount, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+chartOfAccountColumns+` FROM chart_of_accounts WHERE company_id = $1 AND code = $2`, companyID, code)
	a, err := scanChartOfAccount(row)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *chartOfAccountRepository) List(ctx context.Context, filter accounting.ChartOfAccountsFilter) (repositories.Page[*accounting.ChartOfAccount], error) {
	var (
		clauses []string
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.ActiveOnly {
		clauses = append(clauses, "is_active = TRUE")
	}
	if filter.AccountType != "" {
		clauses = append(clauses, fmt.Sprintf("type = $%d", len(args)+1))
		args = append(args, filter.AccountType)
	}
	if filter.ParentID != nil {
		clauses = append(clauses, fmt.Sprintf("parent_id = $%d", len(args)+1))
		args = append(args, *filter.ParentID)
	}
	where := "1=1"
	if len(clauses) > 0 {
		where = persistence.JoinClauses(clauses)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM chart_of_accounts WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*accounting.ChartOfAccount]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM chart_of_accounts WHERE %s ORDER BY code LIMIT $%d OFFSET $%d",
			chartOfAccountColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*accounting.ChartOfAccount]{}, persistence.Translate(err)
	}
	out := make([]*accounting.ChartOfAccount, 0, limit)
	if err := persistence.ScanRows(rows, func(rs *sql.Rows) error {
		a, err := scanChartOfAccountFromRows(rs)
		if err != nil {
			return err
		}
		out = append(out, a)
		return nil
	}); err != nil {
		return repositories.Page[*accounting.ChartOfAccount]{}, err
	}
	return repositories.Page[*accounting.ChartOfAccount]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *chartOfAccountRepository) ListChildren(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID) ([]*accounting.ChartOfAccount, error) {
	query := `SELECT ` + chartOfAccountColumns + ` FROM chart_of_accounts WHERE company_id = $1 AND is_active = TRUE`
	args := []any{companyID}
	if parentID != nil {
		query += fmt.Sprintf(" AND parent_id = $%d", len(args)+1)
		args = append(args, *parentID)
	} else {
		query += " AND parent_id IS NULL"
	}
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, query+" ORDER BY code", args...)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*accounting.ChartOfAccount, 0)
	if err := persistence.ScanRows(rows, func(rs *sql.Rows) error {
		a, err := scanChartOfAccountFromRows(rs)
		if err != nil {
			return err
		}
		out = append(out, a)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func scanChartOfAccount(row *sql.Row) (*accounting.ChartOfAccount, error) {
	a := &accounting.ChartOfAccount{}
	var (
		code, acctType                   string
		description, parentID            sql.NullString
		createdBy, updatedBy             sql.NullString
	)
	err := persistence.ScanRow(row,
		&a.ID, &a.CompanyID, &code, &a.Name, &acctType, &parentID, &a.Path, &a.Depth,
		&a.IsActive, &a.AllowsMovement, &description, &a.CreatedAt, &a.UpdatedAt,
		&createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	decodeChartOfAccount(a, code, acctType, description, parentID, createdBy, updatedBy)
	return a, nil
}

func scanChartOfAccountFromRows(rows *sql.Rows) (*accounting.ChartOfAccount, error) {
	a := &accounting.ChartOfAccount{}
	var (
		code, acctType                   string
		description, parentID            sql.NullString
		createdBy, updatedBy             sql.NullString
	)
	if err := rows.Scan(
		&a.ID, &a.CompanyID, &code, &a.Name, &acctType, &parentID, &a.Path, &a.Depth,
		&a.IsActive, &a.AllowsMovement, &description, &a.CreatedAt, &a.UpdatedAt,
		&createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	decodeChartOfAccount(a, code, acctType, description, parentID, createdBy, updatedBy)
	return a, nil
}

func decodeChartOfAccount(a *accounting.ChartOfAccount, code, acctType string, description, parentID, createdBy, updatedBy sql.NullString) {
	a.Code = valueobjects.ChartOfAccountsCode(code)
	a.Type = persistence.ParseAccountType(acctType)
	if description.Valid {
		a.Description = description.String
	}
	if parentID.Valid {
		id := persistence.ParseUUID(parentID.String)
		a.ParentID = &id
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		a.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		a.UpdatedBy = &id
	}
}

var _ accounting.ChartOfAccountsRepository = (*chartOfAccountRepository)(nil)
