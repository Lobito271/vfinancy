package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/internal/shared/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/infrastructure/postgres"
	"vfinancy/backend/internal/features/sync"
	syncpostgres "vfinancy/backend/internal/features/sync/postgres"
)

// The sync server runs on the cloud PostgreSQL mirror and serves the
// desktop clients. Run with DB_DRIVER=postgres and DB_MIGRATION_DIR
// pointing at migrations/postgres. The client-side sync is off here,
// so set SYNC_ENABLED=false (the default validation also requires
// SYNC_SERVER_URL when sync is enabled).
func main() {
	cfg, err := config.Load()
	if err != nil {
		fail("config error: %v", err)
	}
	if cfg.Database.Driver != "postgres" {
		fail("syncserver requires DB_DRIVER=postgres, got %q", cfg.Database.Driver)
	}
	cfg.Sync.Enabled = false
	log := logger.New(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.Output)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	persistence.SetDialect(persistence.DialectPostgres)
	if err := postgres.EnsureDatabase(ctx, &cfg.Database, log); err != nil {
		fail("ensure database: %v", err)
	}
	db, err := postgres.Connect(ctx, &cfg.Database, log)
	if err != nil {
		fail("connect postgres: %v", err)
	}
	defer db.Close()

	runner := migrations.NewRunner(cfg.Database.MigrationDir, db.DB, log, "postgres")
	if err := runner.Up(ctx); err != nil {
		fail("migrate: %v", err)
	}

	repo := syncpostgres.NewSyncRepository(db.DB)
	server := sync.NewServer(repo, log.Logger)

	srv := &http.Server{
		Addr:              cfg.ListenAddr(),
		Handler:           server.Routes(cfg.Sync.APIKey),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("sync server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fail("serve: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fail("shutdown: %v", err)
	}
}

func fail(format string, args ...any) {
	_, _ = os.Stderr.WriteString(time.Now().Format(time.RFC3339) + " " + fmt.Sprintf(format, args...) + "\n")
	os.Exit(1)
}
