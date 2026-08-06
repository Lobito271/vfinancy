package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/infrastructure/logger"
)

const DriverName = "pgx"

func Connect(ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger) (*database.DB, error) {
	dsn := cfg.DSN()
	log.Info("connecting to postgres",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Name,
		"sslmode", cfg.SSLMode,
	)

	db, err := database.Open(DriverName, dsn, database.Options{
		MaxOpenConns:    cfg.MaxOpen,
		MaxIdleConns:    cfg.MaxIdle,
		ConnMaxLifetime: cfg.MaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	log.Info("postgres connection established")
	return db, nil
}

func EnsureDatabase(ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger) error {
	admin, err := sql.Open(DriverName, cfg.AdminDSN())
	if err != nil {
		return fmt.Errorf("postgres: open admin: %w", err)
	}
	defer admin.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := admin.PingContext(pingCtx); err != nil {
		return fmt.Errorf("postgres: ping admin: %w", err)
	}

	var exists bool
	row := admin.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.Name)
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("postgres: check database: %w", err)
	}
	if exists {
		log.Info("postgres database already exists", "name", cfg.Name)
		return nil
	}

	safeName := sanitizeIdentifier(cfg.Name)
	stmt := fmt.Sprintf("CREATE DATABASE %s", safeName)
	if _, err := admin.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("postgres: create database: %w", err)
	}
	log.Info("postgres database created", "name", cfg.Name)
	return nil
}

func sanitizeIdentifier(name string) string {
	if name == "" {
		return name
	}
	quoted := strings.ReplaceAll(name, `"`, `""`)
	return `"` + quoted + `"`
}

var ErrNotFound = errors.New("postgres: record not found")
