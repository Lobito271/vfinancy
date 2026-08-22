package bindings

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/postgres"
)

// ConnectionConfigDTO is the database connection settings surface to
// the frontend. The password is included so the settings screen can
// round-trip the full configuration.
type ConnectionConfigDTO struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslMode"`
}

// AppSettingsDTO is the application-level (non-business) settings
// surface to the frontend.
type AppSettingsDTO struct {
	WindowTitle string `json:"windowTitle"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	LogLevel    string `json:"logLevel"`
	LogFormat   string `json:"logFormat"`
}

// ModuleDTO describes one module in the module enable/disable list.
type ModuleDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type persistedSettings struct {
	Connection *ConnectionConfigDTO `json:"connection,omitempty"`
	App        *AppSettingsDTO      `json:"app,omitempty"`
	Modules    []moduleSetting      `json:"modules,omitempty"`
}

type moduleSetting struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type moduleCatalog struct {
	id, name, description string
}

var defaultModules = []moduleCatalog{
	{"dashboard", "Dashboard", "Financial and operational overview"},
	{"customers", "Customers", "Customer master data"},
	{"suppliers", "Suppliers", "Supplier master data"},
	{"products", "Products", "Product catalog"},
	{"inventory", "Inventory", "Stock and warehouse management"},
	{"purchasing", "Purchasing", "Purchase orders and supplier payments"},
	{"sales", "Sales", "Sales orders, payments and advances"},
	{"treasury", "Treasury", "Bank accounts, cards and exchange rates"},
	{"accounting", "Accounting", "Chart of accounts and journal entries"},
	{"reports", "Reports", "Financial and operational reports"},
	{"administration", "Administration", "Companies, local profile and audit"},
	{"settings", "Settings", "Application configuration"},
}

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("bindings: user config dir: %w", err)
	}
	return filepath.Join(dir, "vfinancy", "settings.json"), nil
}

func loadPersistedSettings() (persistedSettings, error) {
	var p persistedSettings
	path, err := settingsPath()
	if err != nil {
		return p, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return p, fmt.Errorf("bindings: read settings: %w", err)
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("bindings: parse settings: %w", err)
	}
	return p, nil
}

func savePersistedSettings(p persistedSettings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("bindings: create settings dir: %w", err)
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("bindings: encode settings: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("bindings: write settings: %w", err)
	}
	return nil
}

func (a *App) GetConnectionConfig() (ConnectionConfigDTO, error) {
	p, err := loadPersistedSettings()
	if err != nil {
		return ConnectionConfigDTO{}, err
	}
	if p.Connection != nil {
		return *p.Connection, nil
	}
	return ConnectionConfigDTO{
		Host:     a.cfg.Database.Host,
		Port:     a.cfg.Database.Port,
		User:     a.cfg.Database.User,
		Password: a.cfg.Database.Password,
		Database: a.cfg.Database.Name,
		SSLMode:  a.cfg.Database.SSLMode,
	}, nil
}

func (a *App) SaveConnectionConfig(cfg ConnectionConfigDTO) error {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}
	if cfg.Database == "" {
		return fmt.Errorf("bindings: database name is required")
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	a.cfg.Database.Host = cfg.Host
	a.cfg.Database.Port = cfg.Port
	a.cfg.Database.User = cfg.User
	a.cfg.Database.Password = cfg.Password
	a.cfg.Database.Name = cfg.Database
	a.cfg.Database.SSLMode = cfg.SSLMode

	p, err := loadPersistedSettings()
	if err != nil {
		return err
	}
	p.Connection = &cfg
	return savePersistedSettings(p)
}

func (a *App) TestDatabaseConnection(cfg ConnectionConfigDTO) (string, error) {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}
	if cfg.Database == "" {
		return "", fmt.Errorf("bindings: database name is required")
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	dbCfg := config.DatabaseConfig{
		Host:        cfg.Host,
		Port:        cfg.Port,
		User:        cfg.User,
		Password:    cfg.Password,
		Name:        cfg.Database,
		SSLMode:     cfg.SSLMode,
		MaxOpen:     a.cfg.Database.MaxOpen,
		MaxIdle:     a.cfg.Database.MaxIdle,
		MaxLifetime: a.cfg.Database.MaxLifetime,
	}

	ctx, cancel := context.WithTimeout(a.Context(), 10*time.Second)
	defer cancel()

	db, err := postgres.Connect(ctx, &dbCfg, a.log)
	if err != nil {
		return "", err
	}
	_ = db.Close()
	return "OK", nil
}

func (a *App) GetAppSettings() (AppSettingsDTO, error) {
	p, err := loadPersistedSettings()
	if err != nil {
		return AppSettingsDTO{}, err
	}
	if p.App != nil {
		return *p.App, nil
	}
	return AppSettingsDTO{
		WindowTitle: a.cfg.App.WindowTitle,
		Width:       a.cfg.App.Width,
		Height:      a.cfg.App.Height,
		LogLevel:    a.cfg.Logger.Level,
		LogFormat:   a.cfg.Logger.Format,
	}, nil
}

func (a *App) SaveAppSettings(settings AppSettingsDTO) error {
	if settings.WindowTitle != "" {
		a.cfg.App.WindowTitle = settings.WindowTitle
	}
	if settings.Width > 0 {
		a.cfg.App.Width = settings.Width
	}
	if settings.Height > 0 {
		a.cfg.App.Height = settings.Height
	}
	if settings.LogLevel != "" {
		a.cfg.Logger.Level = settings.LogLevel
	}
	if settings.LogFormat != "" {
		a.cfg.Logger.Format = settings.LogFormat
	}

	p, err := loadPersistedSettings()
	if err != nil {
		return err
	}
	p.App = &settings
	return savePersistedSettings(p)
}

func (a *App) GetModules() ([]ModuleDTO, error) {
	p, err := loadPersistedSettings()
	if err != nil {
		return nil, err
	}

	enabled := make(map[string]bool, len(p.Modules))
	for _, m := range p.Modules {
		enabled[m.ID] = m.Enabled
	}

	result := make([]ModuleDTO, len(defaultModules))
	for i, m := range defaultModules {
		e, ok := enabled[m.id]
		if !ok {
			e = true
		}
		result[i] = ModuleDTO{
			ID:          m.id,
			Name:        m.name,
			Description: m.description,
			Enabled:     e,
		}
	}
	return result, nil
}

func (a *App) SetModuleEnabled(id string, enabled bool) error {
	found := false
	for _, m := range defaultModules {
		if m.id == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("bindings: unknown module %q", id)
	}

	p, err := loadPersistedSettings()
	if err != nil {
		return err
	}
	replaced := false
	for i := range p.Modules {
		if p.Modules[i].ID == id {
			p.Modules[i].Enabled = enabled
			replaced = true
			break
		}
	}
	if !replaced {
		p.Modules = append(p.Modules, moduleSetting{ID: id, Enabled: enabled})
	}
	return savePersistedSettings(p)
}
