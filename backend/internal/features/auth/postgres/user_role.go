package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/auth"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
)

type userRoleRepository struct {
	q persistence.Querier
}

func NewUserRoleRepository(db *sql.DB) *userRoleRepository {
	return &userRoleRepository{q: persistence.FromDB(db)}
}

func (r *userRoleRepository) Assign(ctx context.Context, userID, roleID uuid.UUID, branchID *uuid.UUID, expiresAt *time.Time) error {
	const del = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2 AND branch_id IS NOT DISTINCT FROM $3`
	if _, err := persistence.Q(ctx, r.q).ExecContext(ctx, del, userID, roleID, persistence.NullIfEmptyUUID(branchID)); err != nil {
		return persistence.Translate(err)
	}
	const ins = `INSERT INTO user_roles (user_id, role_id, branch_id, assigned_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, ins,
		userID, roleID, persistence.NullIfEmptyUUID(branchID), time.Now().UTC(), persistence.NullIfZeroTime(expiresAt),
	)
	return persistence.Translate(err)
}

func (r *userRoleRepository) Revoke(ctx context.Context, userID, roleID uuid.UUID, branchID *uuid.UUID) error {
	const q = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2 AND branch_id IS NOT DISTINCT FROM $3`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, userID, roleID, persistence.NullIfEmptyUUID(branchID))
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *userRoleRepository) EffectiveRoles(ctx context.Context, userID uuid.UUID, at time.Time) ([]auth.UserRoleAssignment, error) {
	const q = `SELECT ur.user_id, ur.role_id, ro.code, ur.branch_id, ur.expires_at
		FROM user_roles ur
		JOIN roles ro ON ro.id = ur.role_id
		WHERE ur.user_id = $1 AND (ur.expires_at IS NULL OR ur.expires_at > $2)
		ORDER BY ur.role_id`
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx, q, userID, at)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]auth.UserRoleAssignment, 0)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		a := auth.UserRoleAssignment{}
		var (
			branchID  sql.NullString
			expiresAt sql.NullTime
		)
		if err := r.Scan(&a.UserID, &a.RoleID, &a.RoleCode, &branchID, &expiresAt); err != nil {
			return persistence.Translate(err)
		}
		if branchID.Valid {
			id, _ := uuid.Parse(branchID.String)
			a.BranchID = &id
		}
		if expiresAt.Valid {
			t := expiresAt.Time
			a.ExpiresAt = &t
		}
		out = append(out, a)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}
