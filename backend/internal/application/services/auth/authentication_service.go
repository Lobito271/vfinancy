package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/identity"
	"vfinancy/backend/internal/domain/repositories"
)

type AuthenticationService struct {
	users       repositories.UserRepository
	userRoles   repositories.UserRoleRepository
	params      *Argon2Params
	log         *common.Logger
	maxAttempts int
	lockoutTTL  time.Duration
}

func NewAuthenticationService(
	users repositories.UserRepository,
	userRoles repositories.UserRoleRepository,
	params *Argon2Params,
	log *common.Logger,
	maxAttempts int,
	lockoutTTL time.Duration,
) *AuthenticationService {
	if users == nil {
		panic("auth: nil users repository")
	}
	if userRoles == nil {
		panic("auth: nil userRoles repository")
	}
	if params == nil {
		params = DefaultParams()
	}
	if log == nil {
		panic("auth: nil logger")
	}
	return &AuthenticationService{
		users:       users,
		userRoles:   userRoles,
		params:      params,
		log:         log,
		maxAttempts: maxAttempts,
		lockoutTTL:  lockoutTTL,
	}
}

type AuthenticateResult struct {
	User  *identity.User
	Roles []repositories.UserRoleAssignment
}

func (s *AuthenticationService) Authenticate(ctx context.Context, companyID uuid.UUID, username, password string) (*AuthenticateResult, error) {
	now := time.Now().UTC()

	user, err := s.users.GetByUsername(ctx, companyID, username)
	if err != nil {
		s.log.WithContext(ctx).Error("authentication failed: user lookup", "username", username, "error", err)
		return nil, fmt.Errorf("AUTH_FAILED: credenciales inválidas")
	}

	if !user.IsActive {
		s.log.WithContext(ctx).Warn("authentication denied: inactive user", "user_id", user.ID, "username", username)
		return nil, fmt.Errorf("AUTH_INACTIVE: el usuario está desactivado")
	}

	if user.IsLocked(now) {
		s.log.WithContext(ctx).Warn("authentication denied: locked user", "user_id", user.ID, "locked_until", user.LockedUntil)
		return nil, fmt.Errorf("AUTH_LOCKED: el usuario está bloqueado hasta %s", user.LockedUntil.Format(time.RFC3339))
	}

	match, err := VerifyPassword(password, user.PasswordHash)
	if err != nil {
		s.log.WithContext(ctx).Error("authentication failed: password verification", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("AUTH_FAILED: error interno de verificación")
	}

	if !match {
		attempts := user.RecordFailedLogin(s.maxAttempts, s.lockoutTTL, now)
		if updateErr := s.users.Update(ctx, user); updateErr != nil {
			s.log.WithContext(ctx).Error("failed to persist failed login", "user_id", user.ID, "error", updateErr)
		}
		s.log.WithContext(ctx).Warn("authentication failed: wrong password", "user_id", user.ID, "attempts", attempts)

		if user.IsLocked(now) {
			return nil, fmt.Errorf("AUTH_LOCKED: demasiados intentos fallidos, usuario bloqueado")
		}
		return nil, fmt.Errorf("AUTH_FAILED: credenciales inválidas")
	}

	user.RecordSuccessfulLogin(now, "")
	if updateErr := s.users.Update(ctx, user); updateErr != nil {
		s.log.WithContext(ctx).Error("failed to persist successful login", "user_id", user.ID, "error", updateErr)
	}

	roles, err := s.userRoles.EffectiveRoles(ctx, user.ID, now)
	if err != nil {
		s.log.WithContext(ctx).Error("failed to load effective roles", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("AUTH_FAILED: error cargando roles del usuario")
	}

	s.log.WithContext(ctx).Info("authentication successful", "user_id", user.ID, "username", username)

	return &AuthenticateResult{
		User:  user,
		Roles: roles,
	}, nil
}

func (s *AuthenticationService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("AUTH_NOT_FOUND: usuario no encontrado")
	}

	match, err := VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("AUTH_FAILED: error interno de verificación")
	}
	if !match {
		return fmt.Errorf("AUTH_FAILED: contraseña actual incorrecta")
	}

	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	hash, err := HashPassword(newPassword, s.params)
	if err != nil {
		return fmt.Errorf("AUTH_FAILED: error generando hash de contraseña")
	}

	user.ChangePassword(hash)
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("AUTH_FAILED: error actualizando contraseña")
	}

	s.log.WithContext(ctx).Info("password changed", "user_id", userID)
	return nil
}
