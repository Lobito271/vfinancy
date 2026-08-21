package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/internal/shared/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/infrastructure/postgres"
	"vfinancy/backend/infrastructure/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}
	log := logger.New(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.Output)

	switch os.Args[1] {
	case "migrate":
		runMigrations(cfg, log, os.Args[2:])
	case "status":
		runStatus(cfg, log, os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

// target decides which database the command operates on. The default
// is the local SQLite runtime database; "--postgres" (or "--cloud")
// targets the cloud PostgreSQL mirror.
func target(args []string, cfg *config.Config) (dialect, dir string) {
	dialect = "sqlite"
	dir = cfg.Database.MigrationDir
	for _, a := range args {
		switch strings.TrimPrefix(a, "--") {
		case "postgres", "cloud":
			dialect = "postgres"
			dir = "migrations/postgres"
		}
	}
	return
}

func openLocal(cfg *config.Config) (*database.DB, error) {
	return sqlite.Open(cfg.Database.Path, database.Options{
		MaxOpenConns: 4, MaxIdleConns: 2,
	})
}

func openCloud(ctx context.Context, cfg *config.Config, log *logger.Logger) (*database.DB, error) {
	if err := postgres.EnsureDatabase(ctx, &cfg.Database, log); err != nil {
		return nil, err
	}
	return postgres.Connect(ctx, &cfg.Database, log)
}

func openTarget(ctx context.Context, cfg *config.Config, log *logger.Logger, dialect string) (*database.DB, error) {
	if dialect == "postgres" {
		return openCloud(ctx, cfg, log)
	}
	return openLocal(cfg)
}

func runMigrations(cfg *config.Config, log *logger.Logger, args []string) {
	ctx := context.Background()
	dialect, dir := target(args, cfg)
	persistence.SetDialect(persistence.Dialect(dialect))

	db, err := openTarget(ctx, cfg, log, dialect)
	if err != nil {
		log.Error("connect failed", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	runner := migrations.NewRunner(dir, db.DB, log, dialect)
	if err := runner.Up(ctx); err != nil {
		log.Error("migrate up failed", "error", err.Error())
		os.Exit(1)
	}
}

func runStatus(cfg *config.Config, log *logger.Logger, args []string) {
	ctx := context.Background()
	dialect, dir := target(args, cfg)
	persistence.SetDialect(persistence.Dialect(dialect))

	db, err := openTarget(ctx, cfg, log, dialect)
	if err != nil {
		log.Error("connect failed", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()
	if err := migrations.NewRunner(dir, db.DB, log, dialect).Status(ctx); err != nil {
		log.Error("status failed", "error", err.Error())
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "vfinancy backend CLI")
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  vfinancy migrate [--postgres]   apply pending migrations (default: local SQLite)")
	fmt.Fprintln(os.Stderr, "  vfinancy status  [--postgres]   show migration status")
}
