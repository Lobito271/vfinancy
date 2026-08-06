package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/identity"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

type userRepository struct {
	q Querier
}

func newUserRepository(db *sql.DB) *userRepository {
	return &userRepository{q: &dbBox{db: db}}
}

func newUserRepositoryTx(tx *sql.Tx) *userRepository {
	return &userRepository{q: &txBox{tx: tx}}
}

const userColumns = `
	id, company_id, default_branch_id, username, email, full_name, password_hash,
	must_change_password, failed_login_attempts, locked_until, last_login_at, last_login_ip,
	is_active, created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *userRepository) Create(ctx context.Context, u *identity.User) error {
	const q = `INSERT INTO users (
		id, company_id, default_branch_id, username, email, full_name, password_hash,
		must_change_password, failed_login_attempts, locked_until, last_login_at, last_login_ip,
		is_active, created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`
	_, err := r.q.ExecContext(ctx, q,
		u.ID, u.CompanyID, nullIfEmptyUUID(u.DefaultBranchID), u.Username.String(),
		u.Email.String(), u.FullName.String(), u.PasswordHash,
		u.MustChangePassword, u.FailedLoginAttempts, nullIfZeroTime(u.LockedUntil),
		nullIfZeroTime(u.LastLoginAt), nullIfEmpty(u.LastLoginIP),
		u.IsActive, u.CreatedAt, u.UpdatedAt, nullIfZeroTime(u.DeletedAt),
		nullIfEmptyUUID(u.CreatedBy), nullIfEmptyUUID(u.UpdatedBy),
	)
	return Translate(err)
}

func (r *userRepository) Update(ctx context.Context, u *identity.User) error {
	const q = `UPDATE users SET
		default_branch_id = $1, username = $2, email = $3, full_name = $4,
		password_hash = $5, must_change_password = $6, failed_login_attempts = $7,
		locked_until = $8, last_login_at = $9, last_login_ip = $10,
		is_active = $11, updated_at = $12, updated_by = $13
	 WHERE id = $14 AND deleted_at IS NULL`
	res, err := r.q.ExecContext(ctx, q,
		nullIfEmptyUUID(u.DefaultBranchID), u.Username.String(),
		u.Email.String(), u.FullName.String(),
		u.PasswordHash, u.MustChangePassword, u.FailedLoginAttempts,
		nullIfZeroTime(u.LockedUntil), nullIfZeroTime(u.LastLoginAt),
		nullIfEmpty(u.LastLoginIP),
		u.IsActive, time.Now().UTC(), nullIfEmptyUUID(u.UpdatedBy),
		u.ID,
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

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET deleted_at = $1, updated_at = $2, is_active = FALSE WHERE id = $3 AND deleted_at IS NULL`
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

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE id = $1 AND deleted_at IS NULL`
	row := r.q.QueryRowContext(ctx, q, id)
	return scanUser(row)
}

func (r *userRepository) GetByUsername(ctx context.Context, companyID uuid.UUID, username string) (*identity.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE company_id = $1 AND username = $2 AND deleted_at IS NULL`
	row := r.q.QueryRowContext(ctx, q, companyID, username)
	return scanUser(row)
}

func (r *userRepository) GetByEmail(ctx context.Context, companyID uuid.UUID, email string) (*identity.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE company_id = $1 AND email = $2 AND deleted_at IS NULL`
	row := r.q.QueryRowContext(ctx, q, companyID, email)
	return scanUser(row)
}

func (r *userRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := r.q.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if isPgNoRows(err) {
			return false, nil
		}
		return false, Translate(err)
	}
	return n == 1, nil
}

func (r *userRepository) List(ctx context.Context, filter repositories.UserFilter) (repositories.Page[*identity.User], error) {
	var (
		clauses []string
		args    []any
	)
	if !filter.IncludeDeleted {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.Username != "" {
		clauses = append(clauses, fmt.Sprintf("username = $%d", len(args)+1))
		args = append(args, filter.Username)
	}
	if filter.Email != "" {
		clauses = append(clauses, fmt.Sprintf("email = $%d", len(args)+1))
		args = append(args, filter.Email)
	}
	if filter.Status != "" {
		switch filter.Status {
		case "active":
			clauses = append(clauses, "is_active = TRUE")
		case "inactive":
			clauses = append(clauses, "is_active = FALSE")
		}
	}
	if len(clauses) == 0 {
		clauses = append(clauses, "1=1")
	}

	limit, offset := limitOffset(filter.PageRequest, 25, 200)

	where := joinClauses(clauses)
	var total int
	if err := r.q.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*identity.User]{}, Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := r.q.QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM users WHERE %s ORDER BY username LIMIT $%d OFFSET $%d",
			userColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*identity.User]{}, Translate(err)
	}
	out := make([]*identity.User, 0, limit)
	if err := scanRows(rows, func(r *sql.Rows) error {
		u, err := scanUserFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, u)
		return nil
	}); err != nil {
		return repositories.Page[*identity.User]{}, err
	}
	return repositories.Page[*identity.User]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func scanUser(row *sql.Row) (*identity.User, error) {
	u := &identity.User{}
	var (
		username, email, fullName                     string
		branchID, lastLoginIP, createdBy, updatedBy   sql.NullString
		lockedUntil, lastLoginAt, deletedAt           sql.NullTime
	)
	err := scanRow(row,
		&u.ID, &u.CompanyID, &branchID,
		&username, &email, &fullName, &u.PasswordHash,
		&u.MustChangePassword, &u.FailedLoginAttempts, &lockedUntil,
		&lastLoginAt, &lastLoginIP,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &deletedAt,
		&createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	populateUserFields(u, username, email, fullName, branchID, lastLoginIP, createdBy, updatedBy, lockedUntil, lastLoginAt, deletedAt)
	return u, nil
}

func scanUserFromRows(rows *sql.Rows) (*identity.User, error) {
	u := &identity.User{}
	var (
		username, email, fullName                     string
		branchID, lastLoginIP, createdBy, updatedBy   sql.NullString
		lockedUntil, lastLoginAt, deletedAt           sql.NullTime
	)
	if err := rows.Scan(
		&u.ID, &u.CompanyID, &branchID,
		&username, &email, &fullName, &u.PasswordHash,
		&u.MustChangePassword, &u.FailedLoginAttempts, &lockedUntil,
		&lastLoginAt, &lastLoginIP,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &deletedAt,
		&createdBy, &updatedBy,
	); err != nil {
		return nil, Translate(err)
	}
	populateUserFields(u, username, email, fullName, branchID, lastLoginIP, createdBy, updatedBy, lockedUntil, lastLoginAt, deletedAt)
	return u, nil
}

func populateUserFields(u *identity.User, username, email, fullName string, branchID, lastLoginIP, createdBy, updatedBy sql.NullString, lockedUntil, lastLoginAt, deletedAt sql.NullTime) {
	u.Username = identityParseShortCode(username)
	u.Email = masterdataParseEmail(email)
	u.FullName = masterdataParseFullName(fullName)
	if branchID.Valid {
		id, _ := uuid.Parse(branchID.String)
		u.DefaultBranchID = &id
	}
	if lastLoginIP.Valid {
		u.LastLoginIP = lastLoginIP.String
	}
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		u.CreatedBy = &id
	}
	if updatedBy.Valid {
		id, _ := uuid.Parse(updatedBy.String)
		u.UpdatedBy = &id
	}
	if lockedUntil.Valid {
		t := lockedUntil.Time
		u.LockedUntil = &t
	}
	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		u.LastLoginAt = &t
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		u.DeletedAt = &t
	}
}

func nullIfEmptyUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func identityParseShortCode(s string) valueobjects.ShortCode {
	sc, _ := valueobjects.NewShortCode(s)
	return sc
}
