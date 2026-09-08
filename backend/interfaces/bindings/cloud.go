package bindings

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"vfinancy/backend/internal/utils"
)

// SyncConfigDTO is the sync-server configuration surface to the
// settings screen.
type SyncConfigDTO struct {
	ServerURL       string `json:"serverUrl"`
	APIKey          string `json:"apiKey"`
	Enabled         bool   `json:"enabled"`
	PollIntervalSec int    `json:"pollIntervalSec"`
}

// GetSyncConfig returns the persisted sync configuration, falling back
// to the values loaded from the environment.
func (a *App) GetSyncConfig() (SyncConfigDTO, error) {
	p, err := loadPersistedSettings()
	if err != nil {
		return SyncConfigDTO{}, utils.ProcessError(err)
	}
	if p.Sync != nil {
		return *p.Sync, nil
	}
	return SyncConfigDTO{
		ServerURL:       a.cfg.Sync.ServerURL,
		APIKey:          a.cfg.Sync.APIKey,
		Enabled:         a.cfg.Sync.Enabled,
		PollIntervalSec: int(a.cfg.Sync.PollInterval / time.Second),
	}, nil
}

// SaveSyncConfig persists the sync configuration, applies it to the
// running app and restarts the background worker.
func (a *App) SaveSyncConfig(cfg SyncConfigDTO) error {
	cfg.ServerURL = strings.TrimRight(cfg.ServerURL, "/")
	if cfg.PollIntervalSec <= 0 {
		cfg.PollIntervalSec = 30
	}

	p, err := loadPersistedSettings()
	if err != nil {
		return utils.ProcessError(err)
	}
	p.Sync = &cfg
	if err := savePersistedSettings(p); err != nil {
		return utils.ProcessError(err)
	}

	a.cfg.Sync.ServerURL = cfg.ServerURL
	a.cfg.Sync.APIKey = cfg.APIKey
	a.cfg.Sync.Enabled = cfg.Enabled
	a.cfg.Sync.PollInterval = time.Duration(cfg.PollIntervalSec) * time.Second

	if a.syncCancel != nil {
		a.syncCancel()
		a.syncCancel = nil
	}
	if a.db != nil {
		a.startSyncWorker(a.rawContext())
	}
	return nil
}

// TestSyncConnection verifies a sync server is reachable and the API
// key is accepted. A 200 from /api/v1/health proves both.
func (a *App) TestSyncConnection(serverURL, apiKey string) error {
	serverURL = strings.TrimRight(serverURL, "/")
	if serverURL == "" {
		return fmt.Errorf("bindings: sync server URL is required")
	}
	ctx, cancel := context.WithTimeout(a.rawContext(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL+"/api/v1/health", nil)
	if err != nil {
		return utils.ProcessError(err)
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return utils.ProcessError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bindings: sync server responded %d", resp.StatusCode)
	}
	return nil
}