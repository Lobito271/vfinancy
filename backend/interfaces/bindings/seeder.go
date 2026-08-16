package bindings

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/auth"
)

// seedAuthData ensures the demo admin user exists so the application is
// usable right after a fresh install. It is idempotent and self-healing:
// if the user already exists it is left untouched (including its
// password hash), but its system "admin" role assignment is enforced so
// pre-existing databases still get full CRUD access.
//
// Credentials:
//
//	username: victorfinancy
//	password: adminvic123
//
// The user is granted the system "admin" role, which carries every
// permission in the catalog (plus the "*.*" superadmin wildcard seeded
// by migration 0033).
func (a *App) seedAuthData(ctx context.Context, users auth.UserRepository, userRoles auth.UserRoleRepository) error {
	const (
		username = "victorfinancy"
		password = "adminvic123"
		email    = "victorfinancy@vfinancy.local"
	)

	var userID uuid.UUID
	existing, err := users.GetByUsername(ctx, demoCompanyID, username)
	switch {
	case err == nil:
		userID = existing.ID
		a.log.Info("auth seeder: admin user already exists; ensuring admin role", "username", username)
	case errors.Is(err, repositories.ErrNotFound):
		userID, err = a.createDemoAdminUser(ctx, users, username, password, email)
		if err != nil {
			return err
		}
		a.log.Info("auth seeder: admin user created", "username", username, "user_id", userID)
	default:
		a.log.Warn("auth seeder: could not look up admin user", "username", username, "error", err.Error())
		return err
	}

	var adminRoleID uuid.UUID
	if err := a.db.QueryRowContext(ctx,
		`SELECT id FROM roles WHERE company_id = $1 AND code = 'admin' AND deleted_at IS NULL`,
		demoCompanyID,
	).Scan(&adminRoleID); err != nil {
		return err
	}

	if err := userRoles.Assign(ctx, userID, adminRoleID, nil, nil); err != nil {
		return err
	}

	return nil
}

func (a *App) createDemoAdminUser(ctx context.Context, users auth.UserRepository, username, password, email string) (uuid.UUID, error) {
	hash, err := auth.HashPassword(password, nil)
	if err != nil {
		return uuid.Nil, err
	}

	branchID := demoCompanyID
	user, err := auth.NewUser(time.Now().UTC(), auth.NewUserOptions{
		CompanyID:       demoCompanyID,
		DefaultBranchID: &branchID,
		Username:        valueobjects.ShortCode(username),
		Email:           valueobjects.MustEmail(email),
		FullName:        valueobjects.MustFullName("Victor Financy"),
		PasswordHash:    hash,
	})
	if err != nil {
		return uuid.Nil, err
	}
	user.MustChangePassword = false

	if err := users.Create(ctx, user); err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}
