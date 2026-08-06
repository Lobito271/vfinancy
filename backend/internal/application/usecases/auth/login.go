package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	adminservice "vfinancy/backend/internal/application/services/administration"
	authservice "vfinancy/backend/internal/application/services/auth"
	"vfinancy/backend/internal/application/usecases"
)

type LoginUseCase struct {
	Base     usecases.Base
	Auth     *authservice.AuthenticationService
	Sessions *authservice.SessionService
	Audit    *adminservice.AuditService
}

func NewLoginUseCase(
	base usecases.Base,
	auth *authservice.AuthenticationService,
	sessions *authservice.SessionService,
	audit *adminservice.AuditService,
) *LoginUseCase {
	return &LoginUseCase{Base: base, Auth: auth, Sessions: sessions, Audit: audit}
}

type LoginRequest struct {
	CompanyID string
	Username  string
	Password  string
	IPAddress string
	UserAgent string
	Device    string
	Remember  bool
}

type LoginResponse struct {
	SessionToken       string
	ExpiresAt          time.Time
	User               LoginUserInfo
	MustChangePassword bool
}

type LoginUserInfo struct {
	ID        string
	FullName  string
	Email     string
	Username  string
	Roles     []string
	CompanyID string
}

func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	start := uc.Base.Now()
	uc.Base.LogStart("Login",
		"company_id", req.CompanyID,
		"username", req.Username,
	)

	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return nil, fmtErr(usecases.ErrValidation, "invalid company_id")
	}

	result, err := uc.Auth.Authenticate(ctx, companyID, req.Username, req.Password)
	if err != nil {
		_ = uc.Audit.RecordLoginFailed(ctx, companyID, req.Username, req.IPAddress, req.Device, err.Error())
		uc.Base.LogFinish("Login", start, err, "username", req.Username)
		return nil, usecases.MapError(err)
	}

	session, err := uc.Sessions.Create(ctx, result.User.ID, req.IPAddress, req.UserAgent, req.Device)
	if err != nil {
		uc.Base.LogFinish("Login", start, err, "user_id", result.User.ID)
		return nil, usecases.MapError(err)
	}

	_ = uc.Audit.RecordLogin(ctx, companyID, result.User.ID, session.ID, req.IPAddress, req.Device)

	roles := make([]string, 0, len(result.Roles))
	for _, r := range result.Roles {
		roles = append(roles, r.RoleID.String())
	}

	resp := &LoginResponse{
		SessionToken: session.Token,
		ExpiresAt:    session.ExpiresAt,
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

	uc.Base.LogFinish("Login", start, nil,
		"user_id", result.User.ID,
		"session_id", session.ID,
	)
	return resp, nil
}

func validateLoginRequest(req LoginRequest) error {
	if req.CompanyID == "" {
		return fmtErr(usecases.ErrValidation, "company_id is required")
	}
	if req.Username == "" {
		return fmtErr(usecases.ErrValidation, "username is required")
	}
	if req.Password == "" {
		return fmtErr(usecases.ErrValidation, "password is required")
	}
	return nil
}

func fmtErr(base error, msg string) error {
	return fmt.Errorf("%w: %s", base, msg)
}
