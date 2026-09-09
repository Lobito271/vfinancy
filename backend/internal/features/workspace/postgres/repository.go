package postgres

import (
	"context"
	"database/sql"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/workspace"
)

type repository struct{ q persistence.Querier }

func NewRepository(db *sql.DB) *repository { return &repository{q: persistence.FromDB(db)} }

func (r *repository) GetProfile(ctx context.Context) (*workspace.LocalProfile, error) {
	const q = `SELECT id, name, password_hash, recovery_token_hash, password_enabled,
	 failed_attempts, locked_until, created_at, updated_at FROM local_profiles LIMIT 1`
	p := &workspace.LocalProfile{}
	var hash, tokenHash sql.NullString
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q).Scan(
		&p.ID, &p.Name, &hash, &tokenHash, &p.PasswordEnabled, &p.FailedAttempts,
		&p.LockedUntil, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, workspace.ErrProfileNotFound
		}
		return nil, persistence.Translate(err)
	}
	if hash.Valid {
		p.PasswordHash = hash.String
	}
	if tokenHash.Valid {
		p.RecoveryTokenHash = tokenHash.String
	}
	return p, nil
}

func (r *repository) CreateProfile(ctx context.Context, p *workspace.LocalProfile) error {
	const q = `INSERT INTO local_profiles
	 (id, name, password_hash, recovery_token_hash, password_enabled, failed_attempts,
	  locked_until, created_at, updated_at)
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, p.ID, p.Name,
		p.PasswordHash, p.RecoveryTokenHash, p.PasswordEnabled, p.FailedAttempts,
		p.LockedUntil, p.CreatedAt, p.UpdatedAt)
	return persistence.Translate(err)
}

func (r *repository) UpdateProfile(ctx context.Context, p *workspace.LocalProfile) error {
	const q = `UPDATE local_profiles SET name = $2, password_hash = $3,
	 recovery_token_hash = $4, password_enabled = $5, failed_attempts = $6,
	 locked_until = $7, updated_at = $8 WHERE id = $1`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, p.ID, p.Name,
		p.PasswordHash, p.RecoveryTokenHash, p.PasswordEnabled, p.FailedAttempts,
		p.LockedUntil, p.UpdatedAt)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

var _ workspace.Repository = (*repository)(nil)
