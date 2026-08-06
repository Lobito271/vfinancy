package auth

import (
	"context"

	"github.com/google/uuid"

	adminservice "vfinancy/backend/internal/application/services/administration"
	authservice "vfinancy/backend/internal/application/services/auth"
	"vfinancy/backend/internal/application/usecases"
)

type ChangePasswordUseCase struct {
	Base     usecases.Base
	Auth     *authservice.AuthenticationService
	Sessions *authservice.SessionService
	Audit    *adminservice.AuditService
}

func NewChangePasswordUseCase(
	base usecases.Base,
	auth *authservice.AuthenticationService,
	sessions *authservice.SessionService,
	audit *adminservice.AuditService,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{Base: base, Auth: auth, Sessions: sessions, Audit: audit}
}

type ChangePasswordRequest struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
	SessionToken    string
	CompanyID       string
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, req ChangePasswordRequest) error {
	start := uc.Base.Now()
	uc.Base.LogStart("ChangePassword",
		"user_id", req.UserID,
		"company_id", req.CompanyID,
	)

	if req.SessionToken == "" {
		return fmtErr(usecases.ErrValidation, "session_token is required")
	}
	if req.UserID == "" {
		return fmtErr(usecases.ErrValidation, "user_id is required")
	}
	if req.CurrentPassword == "" {
		return fmtErr(usecases.ErrValidation, "current_password is required")
	}
	if req.NewPassword == "" {
		return fmtErr(usecases.ErrValidation, "new_password is required")
	}
	if req.CompanyID == "" {
		return fmtErr(usecases.ErrValidation, "company_id is required")
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return fmtErr(usecases.ErrValidation, "invalid company_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmtErr(usecases.ErrValidation, "invalid user_id")
	}

	if _, err := uc.Sessions.Validate(ctx, req.SessionToken); err != nil {
		uc.Base.LogFinish("ChangePassword", start, err, "user_id", req.UserID)
		return usecases.MapError(err)
	}

	if err := uc.Auth.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		uc.Base.LogFinish("ChangePassword", start, err, "user_id", req.UserID)
		return usecases.MapError(err)
	}

	if err := uc.Sessions.DestroyAll(ctx, userID); err != nil {
		uc.Base.LogFinish("ChangePassword", start, err, "user_id", req.UserID)
		return usecases.MapError(err)
	}

	_ = uc.Audit.RecordPasswordChange(ctx, companyID, userID)

	uc.Base.LogFinish("ChangePassword", start, nil,
		"user_id", userID,
	)
	return nil
}
