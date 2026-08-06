package auth

import (
	"context"

	"github.com/google/uuid"

	adminservice "vfinancy/backend/internal/application/services/administration"
	authservice "vfinancy/backend/internal/application/services/auth"
	"vfinancy/backend/internal/application/usecases"
)

type LogoutUseCase struct {
	Base     usecases.Base
	Sessions *authservice.SessionService
	Audit    *adminservice.AuditService
}

func NewLogoutUseCase(
	base usecases.Base,
	sessions *authservice.SessionService,
	audit *adminservice.AuditService,
) *LogoutUseCase {
	return &LogoutUseCase{Base: base, Sessions: sessions, Audit: audit}
}

type LogoutRequest struct {
	SessionToken string
	CompanyID    string
}

func (uc *LogoutUseCase) Execute(ctx context.Context, req LogoutRequest) error {
	start := uc.Base.Now()
	uc.Base.LogStart("Logout",
		"company_id", req.CompanyID,
	)

	if req.SessionToken == "" {
		return fmtErr(usecases.ErrValidation, "session_token is required")
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return fmtErr(usecases.ErrValidation, "invalid company_id")
	}

	session, err := uc.Sessions.Validate(ctx, req.SessionToken)
	if err != nil {
		uc.Base.LogFinish("Logout", start, err)
		return usecases.MapError(err)
	}

	if err := uc.Sessions.Destroy(ctx, session.ID); err != nil {
		uc.Base.LogFinish("Logout", start, err, "session_id", session.ID)
		return usecases.MapError(err)
	}

	_ = uc.Audit.RecordLogout(ctx, companyID, session.UserID, session.ID)

	uc.Base.LogFinish("Logout", start, nil,
		"user_id", session.UserID,
		"session_id", session.ID,
	)
	return nil
}

