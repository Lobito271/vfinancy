package bindings

import (
	"encoding/json"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/administration"
)

type BusinessInfoDTO struct {
	Name      string `json:"name"`
	TradeName string `json:"tradeName"`
	TaxID     string `json:"taxId"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Logo      string `json:"logo"`
}

type PreferencesDTO struct {
	DefaultCurrency string `json:"defaultCurrency"`
	DefaultTaxCode  string `json:"defaultTaxCode"`
	ExpiryAlertDays int    `json:"expiryAlertDays"`
	DefaultCountry  string `json:"defaultCountry"`
	DateFormat      string `json:"dateFormat"`
	NumberFormat    string `json:"numberFormat"`
	DecimalPlaces   int    `json:"decimalPlaces"`
	Language        string `json:"language"`
	Theme           string `json:"theme"`
	Timezone        string `json:"timezone"`
	FiscalYearStart int    `json:"fiscalYearStart"`
	BackupFolder    string `json:"backupFolder"`
	ExportFolder    string `json:"exportFolder"`
	BackupFrequency string `json:"backupFrequency"`
}

type CurrencyDTO struct {
	Code          string `json:"code"`
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	DecimalPlaces int    `json:"decimalPlaces"`
	Type          string `json:"type"`
	IsActive      bool   `json:"isActive"`
}

type TaxDTO struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	ShortName    string  `json:"shortName"`
	CountryCode  string  `json:"countryCode"`
	DefaultRate  float64 `json:"defaultRate"`
	IsInclusive  bool    `json:"isInclusive"`
	IsPercentage bool    `json:"isPercentage"`
	Category     string  `json:"category"`
	IsActive     bool    `json:"isActive"`
}

func (a *App) GetBusinessInfo() (*BusinessInfoDTO, error) {
	ctx := a.Context()

	info, err := a.settingsSvc.GetBusinessInfo(ctx, demoCompanyID)
	if err != nil {
		return nil, err
	}

	return &BusinessInfoDTO{
		Name:      info.Name,
		TradeName: info.TradeName,
		TaxID:     info.TaxID,
		Address:   info.Address,
		Phone:     info.Phone,
		Email:     info.Email,
		Logo:      info.Logo,
	}, nil
}

func (a *App) UpdateBusinessInfo(info BusinessInfoDTO) error {
	ctx := a.Context()

	domainInfo := &administration.BusinessInfo{
		Name:      info.Name,
		TradeName: info.TradeName,
		TaxID:     info.TaxID,
		Address:   info.Address,
		Phone:     info.Phone,
		Email:     info.Email,
		Logo:      info.Logo,
	}

	return a.settingsSvc.UpdateBusinessInfo(ctx, demoCompanyID, domainInfo, uuid.Nil)
}

func (a *App) GetPreferences() (*PreferencesDTO, error) {
	ctx := a.Context()

	prefs, err := a.settingsSvc.GetPreferences(ctx, demoCompanyID)
	if err != nil {
		return nil, err
	}

	return &PreferencesDTO{
		DefaultCurrency: prefs.DefaultCurrency,
		DefaultTaxCode:  prefs.DefaultTaxCode,
		ExpiryAlertDays: prefs.ExpiryAlertDays,
		DefaultCountry:  prefs.DefaultCountry,
		DateFormat:      prefs.DateFormat,
		NumberFormat:    prefs.NumberFormat,
		DecimalPlaces:   prefs.DecimalPlaces,
		Language:        prefs.Language,
		Theme:           prefs.Theme,
		Timezone:        prefs.Timezone,
		FiscalYearStart: prefs.FiscalYearStart,
		BackupFolder:    prefs.BackupFolder,
		ExportFolder:    prefs.ExportFolder,
		BackupFrequency: prefs.BackupFrequency,
	}, nil
}

func (a *App) UpdatePreference(key string, value string) error {
	ctx := a.Context()
	return a.settingsSvc.UpdatePreference(ctx, demoCompanyID, key, value, uuid.Nil)
}

func (a *App) GetCurrencies() ([]CurrencyDTO, error) {
	ctx := a.Context()

	currencies, err := a.settingsSvc.GetCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]CurrencyDTO, len(currencies))
	for i, c := range currencies {
		result[i] = CurrencyDTO{
			Code:          c.Code.String(),
			Symbol:        c.Symbol,
			Name:          c.Name,
			DecimalPlaces: c.DecimalPlaces,
			Type:          c.Type.String(),
			IsActive:      c.IsActive,
		}
	}
	return result, nil
}

func (a *App) GetTaxes() ([]TaxDTO, error) {
	ctx := a.Context()

	taxes, err := a.settingsSvc.GetTaxConfiguration(ctx, nil)
	if err != nil {
		return nil, err
	}

	result := make([]TaxDTO, len(taxes))
	for i, t := range taxes {
		result[i] = TaxDTO{
			ID:           t.ID.String(),
			Code:         t.Code,
			Name:         t.Name,
			ShortName:    t.ShortName,
			CountryCode:  t.CountryCode,
			DefaultRate:  t.DefaultRate.Decimal().InexactFloat64(),
			IsInclusive:  t.IsInclusive,
			IsPercentage: t.IsPercentage,
			Category:     t.Category.String(),
			IsActive:     t.IsActive,
		}
	}
	return result, nil
}

func (a *App) GetAllSettings() (map[string]interface{}, error) {
	ctx := a.Context()

	raw, err := a.settingsSvc.GetAllSettings(ctx, demoCompanyID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		var parsed interface{}
		if err := json.Unmarshal(v, &parsed); err != nil {
			result[k] = string(v)
		} else {
			result[k] = parsed
		}
	}
	return result, nil
}
