package bindings

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type BackupResultDTO struct {
	Path         string `json:"path"`
	FallbackUsed bool   `json:"fallbackUsed"`
	Warning      string `json:"warning"`
}

// ChooseBackupFolder opens the native OS directory picker.
func (a *App) ChooseBackupFolder() (string, error) {
	dir, err := wailsruntime.OpenDirectoryDialog(a.rawContext(), wailsruntime.OpenDialogOptions{
		Title: "Carpeta de respaldos",
	})
	if err != nil {
		return "", fmt.Errorf("directory picker: %w", err)
	}
	return dir, nil
}

// SetBackupFolder stores the destination directory preference.
func (a *App) SetBackupFolder(folder string) error {
	_, err := a.UpdatePreference("backup_folder", folder)
	return err
}

// SetBackupFrequency stores the automatic backup policy
// (off | on_close | daily | weekly).
func (a *App) SetBackupFrequency(frequency string) error {
	_, err := a.UpdatePreference("backup_frequency", frequency)
	return err
}

// CreateBackup copies the SQLite database to the configured folder with
// a timestamped name (backup_YYYYMMDD_HHMMSS.db). An inaccessible
// destination falls back to the emergency directory (~/.vfinancy/backups)
// and the result reports it (edge case: unplugged backup drive).
func (a *App) CreateBackup() (BackupResultDTO, error) {
	ctx := a.Context()
	prefs, err := a.settingsSvc.GetPreferences(ctx)
	if err != nil {
		return BackupResultDTO{}, err
	}

	target := prefs.BackupFolder
	if target == "" {
		target = defaultBackupDir()
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		target = defaultBackupDir()
	}

	if err := checkDiskSpace(target, 64<<20); err != nil {
		return BackupResultDTO{}, fmt.Errorf("backup: %w", err)
	}

	if err := a.db.Checkpoint(); err != nil {
		return BackupResultDTO{}, fmt.Errorf("backup: checkpoint: %w", err)
	}

	name := fmt.Sprintf("backup_%s.db", time.Now().UTC().Format("20060102_150405"))
	dest := filepath.Join(target, name)
	warning := ""
	if err := copyDBFile(a.cfg.Database.Path, dest); err != nil {
		fallback := defaultBackupDir()
		if err := os.MkdirAll(fallback, 0o755); err != nil {
			return BackupResultDTO{}, fmt.Errorf("backup: emergency dir: %w", err)
		}
		dest = filepath.Join(fallback, name)
		if err := copyDBFile(a.cfg.Database.Path, dest); err != nil {
			return BackupResultDTO{}, fmt.Errorf("backup: copy: %w", err)
		}
		warning = "No se pudo escribir en la carpeta configurada; se creó una copia de emergencia en " + fallback
		return BackupResultDTO{Path: dest, FallbackUsed: true, Warning: warning}, nil
	}
	return BackupResultDTO{Path: dest}, nil
}

// backupOnClose runs the on-close backup when the user opted in.
func (a *App) backupOnClose() {
	if a.db == nil || a.settingsSvc == nil {
		return
	}
	prefs, err := a.settingsSvc.GetPreferences(context.Background())
	if err != nil || prefs.BackupFrequency != "on_close" {
		return
	}
	if _, err := a.CreateBackup(); err != nil {
		a.log.Warn("on-close backup failed", "error", err.Error())
	}
}

// startBackupWorker runs the periodic automatic backup. Daily/weekly
// frequency compares the newest backup file against the interval.
func (a *App) startBackupWorker(ctx context.Context) {
	wctx, cancel := context.WithCancel(ctx)
	a.backupCancel = cancel
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-wctx.Done():
				return
			case <-ticker.C:
				a.runScheduledBackup()
			}
		}
	}()
}

func (a *App) runScheduledBackup() {
	if a.workspaceSvc == nil || !a.workspaceSvc.IsUnlocked() {
		return
	}
	prefs, err := a.settingsSvc.GetPreferences(context.Background())
	if err != nil {
		return
	}
	interval := map[string]time.Duration{
		"daily":  24 * time.Hour,
		"weekly": 7 * 24 * time.Hour,
	}[prefs.BackupFrequency]
	if interval == 0 {
		return
	}
	folder := prefs.BackupFolder
	if folder == "" {
		folder = defaultBackupDir()
	}
	latest, err := newestFile(folder)
	if err == nil && time.Since(latest.ModTime()) < interval {
		return
	}
	if _, err := a.CreateBackup(); err != nil {
		a.log.Warn("scheduled backup failed", "error", err.Error())
	}
}

func defaultBackupDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(home, ".vfinancy", "backups")
}

// copyDBFile copies the (checkpointed) SQLite file. The WAL/SHM
// companions are not needed after a TRUNCATE checkpoint.
func copyDBFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return out.Sync()
}

// checkDiskSpace refuses to start a backup with less than the minimum
// free space at the destination (risk matrix mitigation).
func checkDiskSpace(dir string, minBytes uint64) error {
	free, err := freeDiskBytes(dir)
	if err != nil {
		return nil // ponytail: unknown filesystem -> skip the check
	}
	if free < minBytes {
		return fmt.Errorf("insufficient disk space at %s", dir)
	}
	return nil
}

func newestFile(dir string) (os.FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var newest os.FileInfo
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".db" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if newest == nil || info.ModTime().After(newest.ModTime()) {
			newest = info
		}
	}
	if newest == nil {
		return nil, fmt.Errorf("no backups yet")
	}
	return newest, nil
}
