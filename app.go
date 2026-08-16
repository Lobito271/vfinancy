package main

import (
	"context"
	"io/fs"
	"log"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/interfaces/bindings"
)

type App struct {
	ctx context.Context

	cfg      *config.Config
	log      *logger.Logger
	bindings *bindings.App
}

func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	l := logger.New(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.Output)

	migrationsFS, err := fs.Sub(sqliteMigrations, "backend/migrations/sqlite")
	if err != nil {
		log.Fatalf("migrations: %v", err)
	}

	return &App{
		cfg:      cfg,
		log:      l,
		bindings: bindings.New(cfg, l, migrationsFS),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.bindings.Startup(ctx)

	if err := a.bindings.Init(); err != nil {
		a.log.Error("bootstrap failed", "error", err.Error())
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.bindings.Shutdown(ctx)
}

func (a *App) Config() *config.Config  { return a.cfg }
func (a *App) Logger() *logger.Logger  { return a.log }
func (a *App) Bindings() *bindings.App { return a.bindings }
