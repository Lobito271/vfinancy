package bindings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type persistedSettings struct {
	Sync *SyncConfigDTO `json:"sync,omitempty"`
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