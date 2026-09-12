package bindings

import (
	"fmt"

	"vfinancy/backend/internal/features/workspace"
)

type LocalProfileDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	TaxID         string `json:"taxId"`
	Email         string `json:"email"`
	FiscalAddress string `json:"fiscalAddress"`
	PasswordEnabled bool `json:"passwordEnabled"`
}

type LocalAuthStateDTO struct {
	Configured      bool `json:"configured"`
	PasswordEnabled bool `json:"passwordEnabled"`
	Unlocked        bool `json:"unlocked"`
}

func profileDTO(p *workspace.LocalProfile) LocalProfileDTO {
	return LocalProfileDTO{ID: p.ID.String(), Name: p.Name, TaxID: p.TaxID, Email: p.Email,
		FiscalAddress: p.FiscalAddress, PasswordEnabled: p.PasswordEnabled}
}

// GetLocalAuthState reports the profile state so the UI can route
// between the setup wizard, the lock screen and the dashboard. It runs
// even while the profile is locked.
func (a *App) GetLocalAuthState() (LocalAuthStateDTO, error) {
	p, err := a.workspaceSvc.Profile()
	if err != nil {
		return LocalAuthStateDTO{Configured: false}, nil
	}
	return LocalAuthStateDTO{
		Configured:      true,
		PasswordEnabled: p.PasswordEnabled,
		Unlocked:        a.workspaceSvc.IsUnlocked(),
	}, nil
}

// GetLocalProfile returns the current local profile.
func (a *App) GetLocalProfile() (LocalProfileDTO, error) {
	p, err := a.workspaceSvc.Profile()
	if err != nil {
		return LocalProfileDTO{}, err
	}
	return profileDTO(p), nil
}

// UpdateLocalProfile replaces the corporate identity of the profile.
func (a *App) UpdateLocalProfile(req SetupWorkspaceRequest) (LocalProfileDTO, error) {
	p, err := a.workspaceSvc.SetCompany(a.Context(), workspace.CompanyInput{
		Name: req.Name, TaxID: req.TaxID, Email: req.Email, FiscalAddress: req.FiscalAddress,
	})
	if err != nil {
		return LocalProfileDTO{}, err
	}
	return profileDTO(p), nil
}

// SetupWorkspaceRequest is the reduced first-run wizard payload: the
// company identity and an optional password.
type SetupWorkspaceRequest struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	TaxID         string `json:"taxId"`
	Email         string `json:"email"`
	FiscalAddress string `json:"fiscalAddress"`
	Password      string `json:"password"`
}

// SetupWorkspace creates the single local profile. It runs before any
// lock can exist, so it uses the raw runtime context.
func (a *App) SetupWorkspace(req SetupWorkspaceRequest) (LocalProfileDTO, error) {
	p, err := a.workspaceSvc.Setup(a.rawContext(), workspace.CompanyInput{
		Name: req.Name, TaxID: req.TaxID, Email: req.Email, FiscalAddress: req.FiscalAddress,
	}, req.Password)
	if err != nil {
		return LocalProfileDTO{}, err
	}
	return profileDTO(p), nil
}

// UnlockLocalProfile verifies the password and lifts the lock.
func (a *App) UnlockLocalProfile(password string) (LocalProfileDTO, error) {
	if err := a.workspaceSvc.Unlock(a.rawContext(), password); err != nil {
		return LocalProfileDTO{}, err
	}
	p, err := a.workspaceSvc.Profile()
	if err != nil {
		return LocalProfileDTO{}, err
	}
	return profileDTO(p), nil
}

// RecoverWithTokenRequest unlocks the profile with the one-time
// recovery token and sets a new password.
type RecoverWithTokenRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// RecoverWithToken implements the lost-password flow (edge case: the
// offline single account has no reset channel).
func (a *App) RecoverWithToken(req RecoverWithTokenRequest) (LocalProfileDTO, error) {
	if err := a.workspaceSvc.UnlockWithRecoveryToken(a.rawContext(), req.Token, req.NewPassword); err != nil {
		return LocalProfileDTO{}, err
	}
	p, err := a.workspaceSvc.Profile()
	if err != nil {
		return LocalProfileDTO{}, err
	}
	return profileDTO(p), nil
}

type ChangePasswordRequest struct {
	Current string `json:"current"`
	Next    string `json:"next"`
}

// SetLocalPassword sets or replaces the local password.
func (a *App) SetLocalPassword(req ChangePasswordRequest) error {
	return a.workspaceSvc.SetPassword(a.Context(), req.Current, req.Next)
}

// RemoveLocalPassword disables the password gate.
func (a *App) RemoveLocalPassword(current string) error {
	return a.workspaceSvc.RemovePassword(a.Context(), current)
}

// GetRecoveryToken returns the one-time recovery token. It can only be
// issued once per profile.
func (a *App) GetRecoveryToken() (string, error) {
	token, err := a.workspaceSvc.GenerateRecoveryToken(a.Context())
	if err != nil {
		return "", fmt.Errorf("recovery token: %w", err)
	}
	return token, nil
}

// LockLocalProfile engages the password gate immediately.
func (a *App) LockLocalProfile() {
	a.workspaceSvc.Lock()
}
