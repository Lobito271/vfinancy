package administration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/administration"
	"vfinancy/backend/internal/domain/entities/masterdata"
	"vfinancy/backend/internal/domain/repositories"
)

type SettingsService struct {
	settings   repositories.SettingRepository
	currencies repositories.CurrencyRepository
	taxes      repositories.TaxRepository
	countries  repositories.CountryRepository
	log        *common.Logger
}

func NewSettingsService(
	settings repositories.SettingRepository,
	currencies repositories.CurrencyRepository,
	taxes repositories.TaxRepository,
	countries repositories.CountryRepository,
	log *common.Logger,
) *SettingsService {
	if settings == nil {
		panic("administration: nil settings repository")
	}
	if currencies == nil {
		panic("administration: nil currencies repository")
	}
	if taxes == nil {
		panic("administration: nil taxes repository")
	}
	if countries == nil {
		panic("administration: nil countries repository")
	}
	if log == nil {
		panic("administration: nil logger")
	}
	return &SettingsService{
		settings:   settings,
		currencies: currencies,
		taxes:      taxes,
		countries:  countries,
		log:        log,
	}
}

type BusinessInfo struct {
	Name      string
	TradeName string
	TaxID     string
	Address   string
	Phone     string
	Email     string
	Logo      string
}

type SystemPreferences struct {
	DefaultCurrency string
	DefaultTaxCode  string
	ExpiryAlertDays int
	DefaultCountry  string
	DateFormat      string
	NumberFormat    string
	DecimalPlaces   int
	Language        string
	Theme           string
	Timezone        string
	FiscalYearStart int
	BackupFolder    string
	ExportFolder    string
	BackupFrequency string
}

func (s *SettingsService) GetBusinessInfo(ctx context.Context, companyID uuid.UUID) (*BusinessInfo, error) {
	if companyID == uuid.Nil {
		return nil, fmt.Errorf("REQUIRED: company id is required")
	}

	info := &BusinessInfo{}

	settings, err := s.settings.ListByCategory(ctx, companyID, "business")
	if err != nil {
		return nil, fmt.Errorf("failed to get business settings: %w", err)
	}

	for _, setting := range settings {
		switch setting.Key {
		case "business.name":
			info.Name = setting.StringValue()
		case "business.trade_name":
			info.TradeName = setting.StringValue()
		case "business.tax_id":
			info.TaxID = setting.StringValue()
		case "business.address":
			info.Address = setting.StringValue()
		case "business.phone":
			info.Phone = setting.StringValue()
		case "business.email":
			info.Email = setting.StringValue()
		case "business.logo":
			info.Logo = setting.StringValue()
		}
	}

	return info, nil
}

func (s *SettingsService) UpdateBusinessInfo(ctx context.Context, companyID uuid.UUID, info *BusinessInfo, updatedBy uuid.UUID) error {
	if companyID == uuid.Nil {
		return fmt.Errorf("REQUIRED: company id is required")
	}
	if info == nil {
		return fmt.Errorf("REQUIRED: business info is required")
	}

	updates := map[string]string{
		"business.name":       info.Name,
		"business.trade_name": info.TradeName,
		"business.tax_id":     info.TaxID,
		"business.address":    info.Address,
		"business.phone":      info.Phone,
		"business.email":      info.Email,
		"business.logo":       info.Logo,
	}

	for key, value := range updates {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}

		setting, err := s.settings.GetByKey(ctx, companyID, key)
		if err != nil {
			setting, err = administration.NewApplicationSetting(
				companyID,
				key,
				"business",
				key,
				"",
				jsonValue,
				false,
			)
			if err != nil {
				return fmt.Errorf("failed to create setting: %w", err)
			}
		} else {
			setting.Update(jsonValue, updatedBy)
		}

		if err := s.settings.Upsert(ctx, setting); err != nil {
			return fmt.Errorf("failed to upsert setting %s: %w", key, err)
		}
	}

	s.log.InfoContext(ctx, "business info updated", "company_id", companyID, "updated_by", updatedBy)
	return nil
}

func (s *SettingsService) GetPreferences(ctx context.Context, companyID uuid.UUID) (*SystemPreferences, error) {
	if companyID == uuid.Nil {
		return nil, fmt.Errorf("REQUIRED: company id is required")
	}

	prefs := &SystemPreferences{
		DefaultCurrency: "PEN",
		DefaultTaxCode:  "IGV",
		ExpiryAlertDays: 30,
		DefaultCountry:  "PE",
		DateFormat:      "DD/MM/YYYY",
		NumberFormat:    "es-PE",
		DecimalPlaces:   2,
		Language:        "es-PE",
		Theme:           "system",
		Timezone:        "America/Lima",
		FiscalYearStart: 1,
		BackupFrequency: "daily",
	}

	settings, err := s.settings.ListByCategory(ctx, companyID, "preferences")
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	for _, setting := range settings {
		switch setting.Key {
		case "preferences.default_currency":
			prefs.DefaultCurrency = setting.StringValue()
		case "preferences.default_tax_code":
			prefs.DefaultTaxCode = setting.StringValue()
		case "preferences.expiry_alert_days":
			prefs.ExpiryAlertDays = setting.IntValue()
		case "preferences.default_country":
			prefs.DefaultCountry = setting.StringValue()
		case "preferences.date_format":
			prefs.DateFormat = setting.StringValue()
		case "preferences.number_format":
			prefs.NumberFormat = setting.StringValue()
		case "preferences.decimal_places":
			prefs.DecimalPlaces = setting.IntValue()
		case "preferences.language":
			prefs.Language = setting.StringValue()
		case "preferences.theme":
			prefs.Theme = setting.StringValue()
		case "preferences.timezone":
			prefs.Timezone = setting.StringValue()
		case "preferences.fiscal_year_start":
			prefs.FiscalYearStart = setting.IntValue()
		case "preferences.backup_folder":
			prefs.BackupFolder = setting.StringValue()
		case "preferences.export_folder":
			prefs.ExportFolder = setting.StringValue()
		case "preferences.backup_frequency":
			prefs.BackupFrequency = setting.StringValue()
		}
	}

	return prefs, nil
}

func (s *SettingsService) UpdatePreference(ctx context.Context, companyID uuid.UUID, key string, value interface{}, updatedBy uuid.UUID) error {
	if companyID == uuid.Nil {
		return fmt.Errorf("REQUIRED: company id is required")
	}
	if key == "" {
		return fmt.Errorf("REQUIRED: key is required")
	}

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	fullKey := "preferences." + key

	setting, err := s.settings.GetByKey(ctx, companyID, fullKey)
	if err != nil {
		setting, err = administration.NewApplicationSetting(
			companyID,
			fullKey,
			"preferences",
			fullKey,
			"",
			jsonValue,
			false,
		)
		if err != nil {
			return fmt.Errorf("failed to create setting: %w", err)
		}
	} else {
		setting.Update(jsonValue, updatedBy)
	}

	if err := s.settings.Upsert(ctx, setting); err != nil {
		return fmt.Errorf("failed to upsert setting: %w", err)
	}

	s.log.InfoContext(ctx, "preference updated", "company_id", companyID, "key", key, "updated_by", updatedBy)
	return nil
}

func (s *SettingsService) GetDefaultCurrency(ctx context.Context, companyID uuid.UUID) (string, error) {
	if companyID == uuid.Nil {
		return "", fmt.Errorf("REQUIRED: company id is required")
	}

	setting, err := s.settings.GetByKey(ctx, companyID, "preferences.default_currency")
	if err != nil {
		return "PEN", nil
	}

	currency := setting.StringValue()
	if currency == "" {
		return "PEN", nil
	}

	return currency, nil
}

func (s *SettingsService) GetTaxConfiguration(ctx context.Context, companyID *uuid.UUID) ([]*masterdata.Tax, error) {
	taxes, err := s.taxes.List(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list taxes: %w", err)
	}
	return taxes, nil
}

func (s *SettingsService) GetCurrencies(ctx context.Context) ([]*masterdata.Currency, error) {
	currencies, err := s.currencies.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}
	return currencies, nil
}

func (s *SettingsService) GetCountries(ctx context.Context) ([]*masterdata.Country, error) {
	countries, err := s.countries.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list countries: %w", err)
	}
	return countries, nil
}

func (s *SettingsService) GetAllSettings(ctx context.Context, companyID uuid.UUID) (map[string]json.RawMessage, error) {
	if companyID == uuid.Nil {
		return nil, fmt.Errorf("REQUIRED: company id is required")
	}

	settings, err := s.settings.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}

	result := make(map[string]json.RawMessage, len(settings))
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}

	return result, nil
}
