package bindings

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"vfinancy/backend/internal/utils"
)

// CreateBackup flushes the WAL, then copies the SQLite database file to
// the configured backup folder (pref "backup.folder", fallback
// ~/.vfinancy/backups). The filename includes a UTC timestamp. Returns
// the absolute path of the created file.
func (a *App) CreateBackup() (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("bindings: database not open")
	}
	ctx := context.Background()
	if _, err := a.db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return "", utils.ProcessError(err)
	}

	src := a.cfg.Database.Path
	if !filepath.IsAbs(src) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", utils.ProcessError(err)
		}
		src = filepath.Join(cwd, src)
	}

	dir := a.resolveBackupFolder(ctx)

	name := fmt.Sprintf("vfinancy_%s.sqlite", time.Now().UTC().Format("20060102_150405"))
	dst := filepath.Join(dir, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", utils.ProcessError(err)
	}
	if err := copyFile(src, dst); err != nil {
		return "", utils.ProcessError(err)
	}
	return dst, nil
}

func (a *App) ExportBackup() (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("bindings: database not open")
	}
	ctx := a.rawContext()

	path, err := wailsRuntime.SaveFileDialog(ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Exportar respaldo",
		DefaultFilename: fmt.Sprintf("vfinancy-backup-%s.db", time.Now().Format("2006-01-02")),
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Base de datos SQLite", Pattern: "*.db"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("backup: file dialog: %w", err)
	}
	if path == "" {
		return "", nil
	}

	if _, err := a.db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return "", utils.ProcessError(err)
	}

	if _, err := a.db.ExecContext(ctx, fmt.Sprintf("VACUUM INTO '%s'", path)); err != nil {
		return "", utils.ProcessError(err)
	}

	return path, nil
}

func (a *App) ImportBackup() error {
	if a.db == nil {
		return fmt.Errorf("bindings: database not open")
	}
	ctx := a.rawContext()

	path, err := wailsRuntime.OpenFileDialog(ctx, wailsRuntime.OpenDialogOptions{
		Title: "Importar respaldo",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Base de datos SQLite", Pattern: "*.db"},
		},
	})
	if err != nil {
		return fmt.Errorf("backup: file dialog: %w", err)
	}
	if path == "" {
		return nil
	}

	if err := validateSQLiteFile(path); err != nil {
		return fmt.Errorf("backup: invalid file: %w", err)
	}

	a.stopWorkers()

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("backup: close db: %w", err)
	}
	a.db = nil

	dbPath := a.cfg.Database.Path
	if !filepath.IsAbs(dbPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("backup: get cwd: %w", err)
		}
		dbPath = filepath.Join(cwd, dbPath)
	}

	backupPath := dbPath + ".bak"
	if err := os.Rename(dbPath, backupPath); err != nil {
		return fmt.Errorf("backup: rename current: %w", err)
	}

	if err := copyFile(path, dbPath); err != nil {
		_ = os.Rename(backupPath, dbPath)
		return fmt.Errorf("backup: copy file: %w", err)
	}

	if err := os.Remove(backupPath); err != nil {
		a.log.Warn("backup: remove .bak failed", "error", err.Error())
	}

	if err := a.openDB(); err != nil {
		return fmt.Errorf("backup: reopen db: %w", err)
	}

	if err := a.initializeServices(ctx); err != nil {
		return fmt.Errorf("backup: reinitialize: %w", err)
	}

	a.log.Info("backup restored successfully", "path", path)
	return nil
}

func validateSQLiteFile(path string) error {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro&_pragma=foreign_keys(0)", path))
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer db.Close()

	var result string
	if err := db.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("integrity check: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}
	return nil
}

func (a *App) resolveBackupFolder(ctx context.Context) string {
	id := a.companyID()
	if id != uuid.Nil {
		if prefs, err := a.settingsSvc.GetPreferences(ctx, id); err == nil && prefs != nil {
			if folder := prefs.BackupFolder; folder != "" {
				return folder
			}
		}
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		return filepath.Join(os.TempDir(), "vfinancy", "backups")
	}
	return filepath.Join(home, ".vfinancy", "backups")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
