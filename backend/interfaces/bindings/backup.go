package bindings

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

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
