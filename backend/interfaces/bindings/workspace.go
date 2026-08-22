package bindings

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/workspace"
)

type LocalAuthStateDTO struct {
	Configured      bool `json:"configured"`
	PasswordEnabled bool `json:"passwordEnabled"`
	Unlocked        bool `json:"unlocked"`
}

type LocalProfileDTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	PasswordEnabled bool   `json:"passwordEnabled"`
	ActiveCompanyID string `json:"activeCompanyId"`
	Theme           string `json:"theme"`
	Language        string `json:"language"`
	DateFormat      string `json:"dateFormat"`
	NumberFormat    string `json:"numberFormat"`
	DecimalPlaces   int    `json:"decimalPlaces"`
	Timezone        string `json:"timezone"`
}

type CompanyDTO struct {
	ID                   string `json:"id"`
	Code                 string `json:"code"`
	LegalName            string `json:"legalName"`
	TradeName            string `json:"tradeName"`
	TaxID                string `json:"taxId"`
	Address              string `json:"address"`
	Phone                string `json:"phone"`
	Email                string `json:"email"`
	CountryCode          string `json:"countryCode"`
	FunctionalCurrency   string `json:"functionalCurrency"`
	Timezone             string `json:"timezone"`
	FiscalYearStartMonth int    `json:"fiscalYearStartMonth"`
	IsActive             bool   `json:"isActive"`
}

type CreateLocalProfileRequest struct {
	Name      string `json:"name"`
	CompanyID string `json:"companyId"`
}

type CompanyRequest struct {
	ID                   string `json:"id"`
	Code                 string `json:"code"`
	LegalName            string `json:"legalName"`
	TradeName            string `json:"tradeName"`
	TaxID                string `json:"taxId"`
	Address              string `json:"address"`
	Phone                string `json:"phone"`
	Email                string `json:"email"`
	CountryCode          string `json:"countryCode"`
	FunctionalCurrency   string `json:"functionalCurrency"`
	Timezone             string `json:"timezone"`
	FiscalYearStartMonth int    `json:"fiscalYearStartMonth"`
}

func localProfileDTO(p *workspace.LocalProfile) *LocalProfileDTO {
	return &LocalProfileDTO{ID: p.ID.String(), Name: p.Name, PasswordEnabled: p.PasswordEnabled,
		ActiveCompanyID: p.ActiveCompanyID.String(), Theme: p.Theme, Language: p.Language,
		DateFormat: p.DateFormat, NumberFormat: p.NumberFormat, DecimalPlaces: p.DecimalPlaces,
		Timezone: p.Timezone}
}

func companyDTO(c *workspace.Company) *CompanyDTO {
	return &CompanyDTO{ID: c.ID.String(), Code: c.Code, LegalName: c.LegalName, TradeName: c.TradeName,
		TaxID: c.TaxID, Address: c.Address, Phone: c.Phone, Email: c.Email, CountryCode: c.CountryCode,
		FunctionalCurrency: c.FunctionalCurrency, Timezone: c.Timezone,
		FiscalYearStartMonth: c.FiscalYearStartMonth, IsActive: c.IsActive}
}

func (a *App) GetLocalAuthState() (LocalAuthStateDTO, error) {
	p, err := a.workspaceSvc.Profile()
	if err != nil {
		return LocalAuthStateDTO{}, err
	}
	return LocalAuthStateDTO{Configured: true, PasswordEnabled: p.PasswordEnabled, Unlocked: a.workspaceSvc.IsUnlocked()}, nil
}

func (a *App) GetLocalProfile() (*LocalProfileDTO, error) {
	p, err := a.workspaceSvc.Profile()
	if err != nil {
		return nil, err
	}
	return localProfileDTO(p), nil
}

type UpdateLocalProfileRequest struct {
	Name          string `json:"name"`
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	DateFormat    string `json:"dateFormat"`
	NumberFormat  string `json:"numberFormat"`
	DecimalPlaces int    `json:"decimalPlaces"`
	Timezone      string `json:"timezone"`
}

func (a *App) UpdateLocalProfile(req UpdateLocalProfileRequest) (*LocalProfileDTO, error) {
	p, err := a.workspaceSvc.UpdateProfile(a.rawContext(), req.Name, req.Theme, req.Language, req.DateFormat, req.NumberFormat, req.Timezone, req.DecimalPlaces)
	if err != nil {
		return nil, err
	}
	return localProfileDTO(p), nil
}

func (a *App) InitializeLocalProfile(req CreateLocalProfileRequest) (*LocalProfileDTO, error) {
	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return nil, err
	}
	p, err := a.workspaceSvc.CreateProfile(a.rawContext(), req.Name, companyID)
	if err != nil {
		return nil, err
	}
	return localProfileDTO(p), nil
}

func (a *App) UnlockLocalProfile(password string) error {
	return a.workspaceSvc.Unlock(a.rawContext(), password)
}

func (a *App) SetLocalPassword(current, next string) error {
	return a.workspaceSvc.SetPassword(a.rawContext(), current, next)
}

func (a *App) RemoveLocalPassword(current string) error {
	return a.workspaceSvc.RemovePassword(a.rawContext(), current)
}

func (a *App) LockLocalProfile() {
	a.workspaceSvc.Lock()
}

func (a *App) ListCompanies() ([]*CompanyDTO, error) {
	companies, err := a.workspaceSvc.ListCompanies(a.Context())
	if err != nil {
		return nil, err
	}
	out := make([]*CompanyDTO, 0, len(companies))
	for _, c := range companies {
		out = append(out, companyDTO(c))
	}
	return out, nil
}

func (a *App) GetActiveCompany() (*CompanyDTO, error) {
	id, err := a.workspaceSvc.CurrentCompanyID()
	if err != nil {
		return nil, err
	}
	c, err := a.workspaceSvc.GetCompany(a.Context(), id)
	if err != nil {
		return nil, err
	}
	return companyDTO(c), nil
}

func (a *App) SetActiveCompany(id string) error {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return a.workspaceSvc.SetActiveCompany(a.rawContext(), companyID)
}

func (a *App) CreateCompany(req CompanyRequest) (*CompanyDTO, error) {
	c := &workspace.Company{ID: uuid.New(), Code: req.Code, LegalName: req.LegalName,
		TradeName: req.TradeName, TaxID: req.TaxID, Address: req.Address, Phone: req.Phone,
		Email: req.Email, CountryCode: req.CountryCode, FunctionalCurrency: req.FunctionalCurrency,
		Timezone: req.Timezone, FiscalYearStartMonth: req.FiscalYearStartMonth,
		CreatedAt: time.Now().UTC()}
	if err := a.workspaceSvc.CreateCompany(a.Context(), c); err != nil {
		return nil, err
	}
	return companyDTO(c), nil
}

func (a *App) UpdateCompany(req CompanyRequest) (*CompanyDTO, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	c, err := a.workspaceSvc.GetCompany(a.Context(), id)
	if err != nil {
		return nil, err
	}
	c.Code = req.Code
	c.LegalName = req.LegalName
	c.TradeName = req.TradeName
	c.TaxID = req.TaxID
	c.Address = req.Address
	c.Phone = req.Phone
	c.Email = req.Email
	c.CountryCode = req.CountryCode
	c.FunctionalCurrency = req.FunctionalCurrency
	c.Timezone = req.Timezone
	c.FiscalYearStartMonth = req.FiscalYearStartMonth
	if err := a.workspaceSvc.UpdateCompany(a.Context(), c); err != nil {
		return nil, err
	}
	return companyDTO(c), nil
}
