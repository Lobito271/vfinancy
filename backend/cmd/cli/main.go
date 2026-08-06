package main

import (
	"context"
	"fmt"
	"os"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/postgres"
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
		runStatus(cfg, log)
	default:
		printUsage()
		os.Exit(1)
	}
}

func runMigrations(cfg *config.Config, log *logger.Logger, args []string) {
	ctx := context.Background()

	if err := postgres.EnsureDatabase(ctx, &cfg.Database, log); err != nil {
		log.Error("ensure database failed", "error", err.Error())
		os.Exit(1)
	}

	db, err := postgres.Connect(ctx, &cfg.Database, log)
	if err != nil {
		log.Error("connect failed", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	runner := migrations.NewRunner(cfg.Database.MigrationDir, db.DB, log)
	if err := runner.Up(ctx); err != nil {
		log.Error("migrate up failed", "error", err.Error())
		os.Exit(1)
	}
}

func runStatus(cfg *config.Config, log *logger.Logger) {
	ctx := context.Background()
	if err := postgres.EnsureDatabase(ctx, &cfg.Database, log); err != nil {
		log.Error("ensure database failed", "error", err.Error())
		os.Exit(1)
	}
	db, err := postgres.Connect(ctx, &cfg.Database, log)
	if err != nil {
		log.Error("connect failed", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()
	if err := migrations.NewRunner(cfg.Database.MigrationDir, db.DB, log).Status(ctx); err != nil {
		log.Error("status failed", "error", err.Error())
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "vfinancy backend CLI")
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  vfinancy migrate         apply pending migrations")
	fmt.Fprintln(os.Stderr, "  vfinancy status          show migration status")
}
