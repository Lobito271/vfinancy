package bindings

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) ExportBackup() (string, error) {
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

	_, err = a.db.ExecContext(ctx, fmt.Sprintf("VACUUM INTO '%s'", path))
	if err != nil {
		return "", fmt.Errorf("backup: vacuum into: %w", err)
	}

	return path, nil
}

func (a *App) ImportBackup() error {
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

	backupPath := a.cfg.Database.Path + ".bak"
	if err := os.Rename(a.cfg.Database.Path, backupPath); err != nil {
		return fmt.Errorf("backup: rename current: %w", err)
	}

	if err := copyFile(path, a.cfg.Database.Path); err != nil {
		_ = os.Rename(backupPath, a.cfg.Database.Path)
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
