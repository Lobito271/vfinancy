package bindings

import (
	"time"

	authuc "vfinancy/backend/internal/application/usecases/auth"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type LoginResponse struct {
	Token              string   `json:"token"`
	ExpiresAt          string   `json:"expiresAt"`
	UserID             string   `json:"userId"`
	FullName           string   `json:"fullName"`
	Email              string   `json:"email"`
	Username           string   `json:"username"`
	Roles              []string `json:"roles"`
	CompanyID          string   `json:"companyId"`
	MustChangePassword bool     `json:"mustChangePassword"`
}

func (a *App) Login(req LoginRequest) (*LoginResponse, error) {
	ctx := a.Context()

	ucReq := authuc.LoginRequest{
		CompanyID: demoCompanyID.String(),
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: "127.0.0.1",
		UserAgent: "desktop",
		Device:    "desktop",
		Remember:  req.Remember,
	}

	resp, err := a.loginUC.Execute(ctx, ucReq)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:              resp.SessionToken,
		ExpiresAt:          resp.ExpiresAt.Format(time.RFC3339),
		UserID:             resp.User.ID,
		FullName:           resp.User.FullName,
		Email:              resp.User.Email,
		Username:           resp.User.Username,
		Roles:              resp.User.Roles,
		CompanyID:          resp.User.CompanyID,
		MustChangePassword: resp.MustChangePassword,
	}, nil
}

func (a *App) Logout(sessionToken string) error {
	ctx := a.Context()

	req := authuc.LogoutRequest{
		SessionToken: sessionToken,
		CompanyID:    demoCompanyID.String(),
	}

	return a.logoutUC.Execute(ctx, req)
}

func (a *App) ChangePassword(currentPassword, newPassword, sessionToken string) error {
	ctx := a.Context()

	session, err := a.sessionSvc.Validate(ctx, sessionToken)
	if err != nil {
		return err
	}

	req := authuc.ChangePasswordRequest{
		UserID:          session.UserID.String(),
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
		SessionToken:    sessionToken,
		CompanyID:       demoCompanyID.String(),
	}

	return a.changePwUC.Execute(ctx, req)
}

func (a *App) ValidateSession(sessionToken string) (bool, error) {
	ctx := a.Context()

	_, err := a.sessionSvc.Validate(ctx, sessionToken)
	if err != nil {
		return false, nil
	}
	return true, nil
}
