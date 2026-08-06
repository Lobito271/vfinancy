package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/administration"
	"vfinancy/backend/internal/shared/apperrors"
	"vfinancy/backend/internal/shared/logger"
)

type AuthenticationService struct {
	users       UserRepository
	userRoles   UserRoleRepository
	sessions    *SessionService
	audit       *administration.AuditService
	params      *Argon2Params
	log         *logger.Logger
	maxAttempts int
	lockoutTTL  time.Duration
}

func NewAuthenticationService(
	users UserRepository,
	userRoles UserRoleRepository,
	sessions *SessionService,
	audit *administration.AuditService,
	params *Argon2Params,
	log *logger.Logger,
	maxAttempts int,
	lockoutTTL time.Duration,
) *AuthenticationService {
	if users == nil {
		panic("auth: nil users repository")
	}
	if userRoles == nil {
		panic("auth: nil userRoles repository")
	}
	if sessions == nil {
		panic("auth: nil session service")
	}
	if audit == nil {
		panic("auth: nil audit service")
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
		sessions:    sessions,
		audit:       audit,
		params:      params,
		log:         log,
		maxAttempts: maxAttempts,
		lockoutTTL:  lockoutTTL,
	}
}

type AuthenticateResult struct {
	User  *User
	Roles []UserRoleAssignment
}

func (s *AuthenticationService) Authenticate(ctx context.Context, companyID uuid.UUID, username, password string) (*AuthenticateResult, error) {
	now := time.Now().UTC()

	user, err := s.users.GetByUsername(ctx, companyID, username)
	if err != nil {
		s.log.WithContext(ctx).Error("authentication failed: user lookup", "username", username, "error", err)
		return nil, ErrAuthFailed
	}

	if !user.IsActive {
		s.log.WithContext(ctx).Warn("authentication denied: inactive user", "user_id", user.ID, "username", username)
		return nil, ErrAuthInactive
	}

	if user.IsLocked(now) {
		s.log.WithContext(ctx).Warn("authentication denied: locked user", "user_id", user.ID, "locked_until", user.LockedUntil)
		return nil, ErrAuthLocked
	}

	match, err := VerifyPassword(password, user.PasswordHash)
	if err != nil {
		s.log.WithContext(ctx).Error("authentication failed: password verification", "user_id", user.ID, "error", err)
		return nil, ErrAuthFailed
	}

	if !match {
		attempts := user.RecordFailedLogin(s.maxAttempts, s.lockoutTTL, now)
		if updateErr := s.users.Update(ctx, user); updateErr != nil {
			s.log.WithContext(ctx).Error("failed to persist failed login", "user_id", user.ID, "error", updateErr)
		}
		s.log.WithContext(ctx).Warn("authentication failed: wrong password", "user_id", user.ID, "attempts", attempts)

		if user.IsLocked(now) {
			return nil, ErrAuthLocked
		}
		return nil, ErrAuthFailed
	}

	user.RecordSuccessfulLogin(now, "")
	if updateErr := s.users.Update(ctx, user); updateErr != nil {
		s.log.WithContext(ctx).Error("failed to persist successful login", "user_id", user.ID, "error", updateErr)
	}

	roles, err := s.userRoles.EffectiveRoles(ctx, user.ID, now)
	if err != nil {
		s.log.WithContext(ctx).Error("failed to load effective roles", "user_id", user.ID, "error", err)
		return nil, ErrAuthFailed
	}

	s.log.WithContext(ctx).Info("authentication successful", "user_id", user.ID, "username", username)

	return &AuthenticateResult{
		User:  user,
		Roles: roles,
	}, nil
}

// LoginRequest is the input for Login.
type LoginRequest struct {
	CompanyID string
	Username  string
	Password  string
	IPAddress string
	UserAgent string
	Device    string
	Remember  bool
}

// LoginResponse is what Login returns on success.
type LoginResponse struct {
	SessionToken       string
	ExpiresAt          time.Time
	User               LoginUserInfo
	MustChangePassword bool
}

// LoginUserInfo is the user projection returned by Login.
type LoginUserInfo struct {
	ID        string
	FullName  string
	Email     string
	Username  string
	Roles     []string
	CompanyID string
}

// Login authenticates a user, creates a session and records the audit
// event. It returns a session token on success.
func (s *AuthenticationService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return nil, apperrors.Errorf(apperrors.ErrValidation, "invalid company_id")
	}

	result, err := s.Authenticate(ctx, companyID, req.Username, req.Password)
	if err != nil {
		_ = s.audit.RecordLoginFailed(ctx, companyID, req.Username, req.IPAddress, req.Device, err.Error())
		s.log.WithContext(ctx).Warn("login failed", "username", req.Username, "error", err)
		return nil, apperrors.MapError(err)
	}

	session, err := s.sessions.Create(ctx, result.User.ID, req.IPAddress, req.UserAgent, req.Device)
	if err != nil {
		s.log.WithContext(ctx).Error("login failed: session creation", "user_id", result.User.ID, "error", err)
		return nil, apperrors.MapError(err)
	}

	_ = s.audit.RecordLogin(ctx, companyID, result.User.ID, session.ID, req.IPAddress, req.Device)

	roles := make([]string, 0, len(result.Roles))
	for _, r := range result.Roles {
		roles = append(roles, r.RoleID.String())
	}

	resp := &LoginResponse{
		SessionToken:       session.Token,
		ExpiresAt:          session.ExpiresAt,
		MustChangePassword: result.User.MustChangePassword,
		User: LoginUserInfo{
			ID:        result.User.ID.String(),
			FullName:  result.User.FullName.String(),
			Email:     result.User.Email.String(),
			Username:  result.User.Username.String(),
			Roles:     roles,
			CompanyID: result.User.CompanyID.String(),
		},
	}

	s.log.WithContext(ctx).Info("login successful",
		"user_id", result.User.ID,
		"session_id", session.ID,
	)
	return resp, nil
}

// LogoutRequest is the input for Logout.
type LogoutRequest struct {
	SessionToken string
	CompanyID    string
}

// Logout destroys the session and records the audit event.
func (s *AuthenticationService) Logout(ctx context.Context, req LogoutRequest) error {
	if req.SessionToken == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "session_token is required")
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return apperrors.Errorf(apperrors.ErrValidation, "invalid company_id")
	}

	session, err := s.sessions.Validate(ctx, req.SessionToken)
	if err != nil {
		return apperrors.MapError(err)
	}

	if err := s.sessions.Destroy(ctx, session.ID); err != nil {
		s.log.WithContext(ctx).Error("logout failed: session destroy", "session_id", session.ID, "error", err)
		return apperrors.MapError(err)
	}

	_ = s.audit.RecordLogout(ctx, companyID, session.UserID, session.ID)

	s.log.WithContext(ctx).Info("logout successful",
		"user_id", session.UserID,
		"session_id", session.ID,
	)
	return nil
}

// ChangePasswordRequest is the input for ChangePassword.
type ChangePasswordRequest struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
	SessionToken    string
	CompanyID       string
}

// ChangePassword validates the session, changes the password, destroys
// all other sessions and records the audit event.
func (s *AuthenticationService) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	if req.SessionToken == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "session_token is required")
	}
	if req.UserID == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "user_id is required")
	}
	if req.CurrentPassword == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "current_password is required")
	}
	if req.NewPassword == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "new_password is required")
	}
	if req.CompanyID == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "company_id is required")
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return apperrors.Errorf(apperrors.ErrValidation, "invalid company_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return apperrors.Errorf(apperrors.ErrValidation, "invalid user_id")
	}

	if _, err := s.sessions.Validate(ctx, req.SessionToken); err != nil {
		return apperrors.MapError(err)
	}

	if err := s.changePassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		return apperrors.MapError(err)
	}

	if err := s.sessions.DestroyAll(ctx, userID); err != nil {
		return apperrors.MapError(err)
	}

	_ = s.audit.RecordPasswordChange(ctx, companyID, userID)

	s.log.WithContext(ctx).Info("password changed", "user_id", userID)
	return nil
}

// changePassword verifies the current password, validates the strength
// of the new one, and persists the new hash.
func (s *AuthenticationService) changePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return ErrAuthNotFound
	}

	match, err := VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil {
		return ErrAuthFailed
	}
	if !match {
		return ErrAuthFailed
	}

	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	hash, err := HashPassword(newPassword, s.params)
	if err != nil {
		return ErrAuthFailed
	}

	user.ChangePassword(hash)
	if err := s.users.Update(ctx, user); err != nil {
		return ErrAuthFailed
	}

	s.log.WithContext(ctx).Info("password persisted", "user_id", userID)
	return nil
}

func validateLoginRequest(req LoginRequest) error {
	if req.CompanyID == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "company_id is required")
	}
	if req.Username == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "username is required")
	}
	if req.Password == "" {
		return apperrors.Errorf(apperrors.ErrValidation, "password is required")
	}
	return nil
}
