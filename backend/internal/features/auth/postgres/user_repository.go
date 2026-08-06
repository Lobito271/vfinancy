package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/auth"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

type userRepository struct {
	q persistence.Querier
}

func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{q: persistence.FromDB(db)}
}


const userColumns = `
	id, company_id, default_branch_id, username, email, full_name, password_hash,
	must_change_password, failed_login_attempts, locked_until, last_login_at, last_login_ip,
	is_active, created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *userRepository) Create(ctx context.Context, u *auth.User) error {
	const q = `INSERT INTO users (
		id, company_id, default_branch_id, username, email, full_name, password_hash,
		must_change_password, failed_login_attempts, locked_until, last_login_at, last_login_ip,
		is_active, created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		u.ID, u.CompanyID, persistence.NullIfEmptyUUID(u.DefaultBranchID), u.Username.String(),
		u.Email.String(), u.FullName.String(), u.PasswordHash,
		u.MustChangePassword, u.FailedLoginAttempts, persistence.NullIfZeroTime(u.LockedUntil),
		persistence.NullIfZeroTime(u.LastLoginAt), persistence.NullIfEmpty(u.LastLoginIP),
		u.IsActive, u.CreatedAt, u.UpdatedAt, persistence.NullIfZeroTime(u.DeletedAt),
		persistence.NullIfEmptyUUID(u.CreatedBy), persistence.NullIfEmptyUUID(u.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *userRepository) Update(ctx context.Context, u *auth.User) error {
	const q = `UPDATE users SET
		default_branch_id = $1, username = $2, email = $3, full_name = $4,
		password_hash = $5, must_change_password = $6, failed_login_attempts = $7,
		locked_until = $8, last_login_at = $9, last_login_ip = $10,
		is_active = $11, updated_at = $12, updated_by = $13
	 WHERE id = $14 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		persistence.NullIfEmptyUUID(u.DefaultBranchID), u.Username.String(),
		u.Email.String(), u.FullName.String(),
		u.PasswordHash, u.MustChangePassword, u.FailedLoginAttempts,
		persistence.NullIfZeroTime(u.LockedUntil), persistence.NullIfZeroTime(u.LastLoginAt),
		persistence.NullIfEmpty(u.LastLoginIP),
		u.IsActive, time.Now().UTC(), persistence.NullIfEmptyUUID(u.UpdatedBy),
		u.ID,
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

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET deleted_at = $1, updated_at = $2, is_active = FALSE WHERE id = $3 AND deleted_at IS NULL`
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

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanUser(row)
}

func (r *userRepository) GetByUsername(ctx context.Context, companyID uuid.UUID, username string) (*auth.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE company_id = $1 AND username = $2 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, username)
	return scanUser(row)
}

func (r *userRepository) GetByEmail(ctx context.Context, companyID uuid.UUID, email string) (*auth.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE company_id = $1 AND email = $2 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, email)
	return scanUser(row)
}

func (r *userRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if persistence.IsPgNoRows(err) {
			return false, nil
		}
		return false, persistence.Translate(err)
	}
	return n == 1, nil
}

func (r *userRepository) List(ctx context.Context, filter auth.UserFilter) (repositories.Page[*auth.User], error) {
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

	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := persistence.JoinClauses(clauses)
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM users WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*auth.User]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM users WHERE %s ORDER BY username LIMIT $%d OFFSET $%d",
			userColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*auth.User]{}, persistence.Translate(err)
	}
	out := make([]*auth.User, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		u, err := scanUserFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, u)
		return nil
	}); err != nil {
		return repositories.Page[*auth.User]{}, err
	}
	return repositories.Page[*auth.User]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func scanUser(row *sql.Row) (*auth.User, error) {
	u := &auth.User{}
	var (
		username, email, fullName                     string
		branchID, lastLoginIP, createdBy, updatedBy   sql.NullString
		lockedUntil, lastLoginAt, deletedAt           sql.NullTime
	)
	err := persistence.ScanRow(row,
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

func scanUserFromRows(rows *sql.Rows) (*auth.User, error) {
	u := &auth.User{}
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
		return nil, persistence.Translate(err)
	}
	populateUserFields(u, username, email, fullName, branchID, lastLoginIP, createdBy, updatedBy, lockedUntil, lastLoginAt, deletedAt)
	return u, nil
}

func populateUserFields(u *auth.User, username, email, fullName string, branchID, lastLoginIP, createdBy, updatedBy sql.NullString, lockedUntil, lastLoginAt, deletedAt sql.NullTime) {
	u.Username = identityParseShortCode(username)
	u.Email = persistence.ParseEmail(email)
	u.FullName = persistence.ParseFullName(fullName)
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

func identityParseShortCode(s string) valueobjects.ShortCode {
	sc, _ := valueobjects.NewShortCode(s)
	return sc
}
