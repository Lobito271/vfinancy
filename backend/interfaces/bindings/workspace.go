package bindings

import "vfinancy/backend/internal/features/workspace"

type LocalProfileDTO struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	CommercialName     string `json:"commercialName"`
	TaxID              string `json:"taxId"`
	Email              string `json:"email"`
	FiscalAddress      string `json:"fiscalAddress"`
	Phone              string `json:"phone"`
	Website            string `json:"website"`
	SecurityQuestion   string `json:"securityQuestion"`
	PasswordEnabled    bool   `json:"passwordEnabled"`
}

type LocalAuthStateDTO struct {
	Configured      bool `json:"configured"`
	PasswordEnabled bool `json:"passwordEnabled"`
	Unlocked        bool `json:"unlocked"`
}

func profileDTO(p *workspace.LocalProfile) LocalProfileDTO {
	return LocalProfileDTO{ID: p.ID.String(), Name: p.Name, CommercialName: p.CommercialName, TaxID: p.TaxID,
		Email: p.Email, FiscalAddress: p.FiscalAddress, Phone: p.Phone, Website: p.Website,
		SecurityQuestion: p.SecurityQuestion, PasswordEnabled: p.PasswordEnabled}
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
		Name: req.Name, CommercialName: req.CommercialName, TaxID: req.TaxID, Email: req.Email,
		FiscalAddress: req.FiscalAddress, Phone: req.Phone, Website: req.Website,
	})
	if err != nil {
		return LocalProfileDTO{}, err
	}
	return profileDTO(p), nil
}

// SetupWorkspaceRequest is the reduced first-run wizard payload: the
// company identity and an optional password.
type SetupWorkspaceRequest struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CommercialName string `json:"commercialName"`
	TaxID          string `json:"taxId"`
	Email          string `json:"email"`
	FiscalAddress  string `json:"fiscalAddress"`
	Phone          string `json:"phone"`
	Website        string `json:"website"`
	Password       string `json:"password"`
}

// SetupWorkspace creates the single local profile. It runs before any
// lock can exist, so it uses the raw runtime context.
func (a *App) SetupWorkspace(req SetupWorkspaceRequest) (LocalProfileDTO, error) {
	p, err := a.workspaceSvc.Setup(a.rawContext(), workspace.CompanyInput{
		Name: req.Name, CommercialName: req.CommercialName, TaxID: req.TaxID, Email: req.Email,
		FiscalAddress: req.FiscalAddress, Phone: req.Phone, Website: req.Website,
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

// SetSecurityQuestionRequest stores a question text and its answer.
type SetSecurityQuestionRequest struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// RecoverWithAnswerRequest unlocks the profile with the stored security
// answer and sets a new password.
type RecoverWithAnswerRequest struct {
	Answer      string `json:"answer"`
	NewPassword string `json:"newPassword"`
}

// GetSecurityQuestion returns the stored security question so the lock
// screen can prompt for its answer. It runs while the profile is locked.
func (a *App) GetSecurityQuestion() (string, error) {
	return a.workspaceSvc.SecurityQuestion(), nil
}

// SetSecurityQuestion stores the question and hashed answer.
func (a *App) SetSecurityQuestion(req SetSecurityQuestionRequest) error {
	return a.workspaceSvc.SetSecurityQuestion(a.rawContext(), req.Question, req.Answer)
}

// ClearSecurityQuestion removes the stored question and answer.
func (a *App) ClearSecurityQuestion() error {
	return a.workspaceSvc.ClearSecurityQuestion(a.rawContext())
}

// RecoverWithAnswer implements the lost-password flow (edge case: the
// offline single account has no reset channel).
func (a *App) RecoverWithAnswer(req RecoverWithAnswerRequest) (LocalProfileDTO, error) {
	if err := a.workspaceSvc.UnlockWithAnswer(a.rawContext(), req.Answer, req.NewPassword); err != nil {
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

// LockLocalProfile engages the password gate immediately.
func (a *App) LockLocalProfile() {
	a.workspaceSvc.Lock()
}
