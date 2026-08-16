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
	"vfinancy/backend/internal/features/treasury"
)

type bankAccountRepository struct {
	q persistence.Querier
}

func NewBankAccountRepository(db *sql.DB) *bankAccountRepository {
	return &bankAccountRepository{q: persistence.FromDB(db)}
}

const bankAccountColumns = `
	id, company_id, branch_id, bank_name, account_number,
	account_type, currency_code, gl_account_id, current_balance,
	is_default, is_active, created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *bankAccountRepository) Create(ctx context.Context, a *treasury.BankAccount) error {
	const q = `INSERT INTO bank_accounts (
		id, company_id, branch_id, bank_name, account_number,
		account_type, currency_code, gl_account_id, current_balance,
		is_default, is_active, created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		a.ID, a.CompanyID, persistence.NullIfEmptyUUID(a.BranchID),
		a.BankName, a.AccountNumber, a.AccountType, a.CurrencyCode.String(),
		a.GLAccountID, a.CurrentBalance.String(), a.IsDefault, a.IsActive,
		a.CreatedAt, a.UpdatedAt, persistence.NullIfZeroTime(a.DeletedAt),
		persistence.NullIfEmptyUUID(a.CreatedBy), persistence.NullIfEmptyUUID(a.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *bankAccountRepository) Update(ctx context.Context, a *treasury.BankAccount) error {
	const q = `UPDATE bank_accounts SET
		bank_name = $1, account_number = $2, account_type = $3, currency_code = $4,
		gl_account_id = $5, current_balance = $6, is_default = $7, is_active = $8,
		branch_id = $9, updated_at = $10, updated_by = $11
	 WHERE id = $12 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		a.BankName, a.AccountNumber, a.AccountType, a.CurrencyCode.String(),
		a.GLAccountID, a.CurrentBalance.String(), a.IsDefault, a.IsActive,
		persistence.NullIfEmptyUUID(a.BranchID), time.Now().UTC(), persistence.NullIfEmptyUUID(a.UpdatedBy),
		a.ID,
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

func (r *bankAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE bank_accounts SET deleted_at = $1, updated_at = $2, is_active = FALSE WHERE id = $3 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, now, now, id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *bankAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*treasury.BankAccount, error) {
	q := `SELECT ` + bankAccountColumns + ` FROM bank_accounts WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanBankAccount(row)
}

func (r *bankAccountRepository) List(ctx context.Context, filter treasury.BankAccountFilter) (repositories.Page[*treasury.BankAccount], error) {
	var (
		clauses = []string{"deleted_at IS NULL"}
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.BranchID != nil {
		clauses = append(clauses, fmt.Sprintf("branch_id = $%d", len(args)+1))
		args = append(args, *filter.BranchID)
	}
	if filter.IsActive != nil {
		clauses = append(clauses, fmt.Sprintf("is_active = $%d", len(args)+1))
		args = append(args, *filter.IsActive)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := persistence.JoinClauses(clauses)
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM bank_accounts WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*treasury.BankAccount]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM bank_accounts WHERE %s ORDER BY bank_name LIMIT $%d OFFSET $%d",
			bankAccountColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*treasury.BankAccount]{}, persistence.Translate(err)
	}
	out := make([]*treasury.BankAccount, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		a, err := scanBankAccountFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, a)
		return nil
	}); err != nil {
		return repositories.Page[*treasury.BankAccount]{}, err
	}
	return repositories.Page[*treasury.BankAccount]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func scanBankAccount(row *sql.Row) (*treasury.BankAccount, error) {
	a := &treasury.BankAccount{}
	var (
		branchID, createdBy, updatedBy sql.NullString
		deletedAt                      sql.NullTime
		currencyCode, currentBalance   string
	)
	err := persistence.ScanRow(row,
		&a.ID, &a.CompanyID, &branchID, &a.BankName, &a.AccountNumber,
		&a.AccountType, &currencyCode, &a.GLAccountID, &currentBalance,
		&a.IsDefault, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
		&deletedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	if err := decodeBankAccount(a, branchID, createdBy, updatedBy, deletedAt, currencyCode, currentBalance); err != nil {
		return nil, err
	}
	return a, nil
}

func scanBankAccountFromRows(rows *sql.Rows) (*treasury.BankAccount, error) {
	a := &treasury.BankAccount{}
	var (
		branchID, createdBy, updatedBy sql.NullString
		deletedAt                      sql.NullTime
		currencyCode, currentBalance   string
	)
	if err := rows.Scan(
		&a.ID, &a.CompanyID, &branchID, &a.BankName, &a.AccountNumber,
		&a.AccountType, &currencyCode, &a.GLAccountID, &currentBalance,
		&a.IsDefault, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
		&deletedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodeBankAccount(a, branchID, createdBy, updatedBy, deletedAt, currencyCode, currentBalance); err != nil {
		return nil, err
	}
	return a, nil
}

func decodeBankAccount(a *treasury.BankAccount, branchID, createdBy, updatedBy sql.NullString, deletedAt sql.NullTime, currencyCode, currentBalance string) error {
	if branchID.Valid {
		id := persistence.ParseUUID(branchID.String)
		a.BranchID = &id
	}
	cc, err := valueobjects.NewCurrencyCode(currencyCode)
	if err != nil {
		return err
	}
	a.CurrencyCode = cc
	bal, err := persistence.ParseMoney(currentBalance)
	if err != nil {
		return err
	}
	a.CurrentBalance = bal
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		a.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		a.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		a.DeletedAt = &t
	}
	return nil
}

var _ treasury.BankAccountRepository = (*bankAccountRepository)(nil)
