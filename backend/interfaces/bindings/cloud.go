package bindings

import (
	"errors"
	"time"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/internal/features/sync"
	syncpostgres "vfinancy/backend/internal/features/sync/postgres"
)

// SyncConfigDTO mirrors the cloud sync connection fields required by
// the spec: host, port, database, user, password (+ interval).
type SyncConfigDTO struct {
	Enabled        bool   `json:"enabled"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Database       string `json:"database"`
	User           string `json:"user"`
	Password       string `json:"password"`
	SSLMode        string `json:"sslMode"`
	PollIntervalSec int   `json:"pollIntervalSec"`
}

// effectiveSyncConfig merges the env config with the runtime settings
// file (UI takes precedence).
func (a *App) effectiveSyncConfig() (config.SyncConfig, error) {
	cfg := a.cfg.Sync
	persisted, err := loadPersistedSettings()
	if err != nil {
		a.log.Warn("load settings file failed; using env sync config", "error", err.Error())
		return cfg, nil
	}
	if persisted.Sync != nil {
		s := persisted.Sync
		if s.Host != "" {
			cfg.Host = s.Host
		}
		if s.Port > 0 {
			cfg.Port = s.Port
		}
		if s.Database != "" {
			cfg.Name = s.Database
		}
		if s.User != "" {
			cfg.User = s.User
		}
		if s.Password != "" {
			cfg.Password = s.Password
		}
		if s.SSLMode != "" {
			cfg.SSLMode = s.SSLMode
		}
		if s.PollIntervalSec > 0 {
			cfg.PollInterval = time.Duration(s.PollIntervalSec) * time.Second
		}
		cfg.Enabled = s.Enabled && cfg.Host != "" && cfg.Name != "" && cfg.User != ""
	}
	return cfg, nil
}

// GetSyncConfig returns the current cloud sync configuration.
func (a *App) GetSyncConfig() (SyncConfigDTO, error) {
	cfg, err := a.effectiveSyncConfig()
	if err != nil {
		return SyncConfigDTO{}, err
	}
	return SyncConfigDTO{
		Enabled:        cfg.Enabled,
		Host:           cfg.Host,
		Port:           cfg.Port,
		Database:       cfg.Name,
		User:           cfg.User,
		Password:       cfg.Password,
		SSLMode:        cfg.SSLMode,
		PollIntervalSec: int(cfg.PollInterval / time.Second),
	}, nil
}

// SaveSyncConfig persists the cloud sync connection and restarts the
// background worker (opt-in switch per spec).
func (a *App) SaveSyncConfig(req SyncConfigDTO) error {
	if err := savePersistedSettings(persistedSettings{Sync: &req}); err != nil {
		return err
	}
	a.startSyncWorker()
	return nil
}

// TestSyncConnection dials the cloud PostgreSQL mirror and verifies the
// remote schema.
func (a *App) TestSyncConnection(req SyncConfigDTO) error {
	dsn := (&config.SyncConfig{
		Host: req.Host, Port: req.Port, Name: req.Database,
		User: req.User, Password: req.Password, SSLMode: req.SSLMode,
	}).DSN()
	svc := sync.NewService(
		syncpostgres.NewLocal(a.db.DB),
		syncpostgres.NewRemote(dsn, a.log),
		sync.Config{DSN: dsn, MigrationsFS: a.pgMigrationsFS},
		a.log.Logger,
	)
	defer svc.Close()
	return svc.TestConnection(a.rawContext())
}

// SyncNow runs one replication pass on demand, even when the
// background worker interval is disabled.
func (a *App) SyncNow() error {
	if a.syncSvc != nil {
		return a.syncSvc.RunOnce(a.rawContext())
	}
	cfg, err := a.effectiveSyncConfig()
	if err != nil {
		return err
	}
	if !cfg.Enabled || cfg.DSN() == "" {
		return errors.New("sincronización desactivada: configura el servidor primero")
	}
	svc := sync.NewService(
		syncpostgres.NewLocal(a.db.DB),
		syncpostgres.NewRemote(cfg.DSN(), a.log),
		sync.Config{DSN: cfg.DSN(), MigrationsFS: a.pgMigrationsFS},
		a.log.Logger,
	)
	defer svc.Close()
	return svc.RunOnce(a.rawContext())
}
