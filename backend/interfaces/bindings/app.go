package bindings

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/postgres"
	"vfinancy/backend/internal/features/administration"
	adminpostgres "vfinancy/backend/internal/features/administration/postgres"
	"vfinancy/backend/internal/features/auth"
	authpostgres "vfinancy/backend/internal/features/auth/postgres"
	sharedlogger "vfinancy/backend/internal/shared/logger"
)

var demoCompanyID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type App struct {
	ctx context.Context
	db  *database.DB
	cfg *config.Config
	log *logger.Logger

	appLog *sharedlogger.Logger

	authSvc     *auth.AuthenticationService
	settingsSvc *administration.SettingsService
	profileSvc  *auth.ProfileService
	auditSvc    *administration.AuditService
	sessionSvc  *auth.SessionService

	users auth.UserRepository
}

func New(cfg *config.Config, log *logger.Logger) *App {
	return &App{cfg: cfg, log: log}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(ctx context.Context) {
	if a.db != nil {
		_ = a.db.Close()
	}
}

func (a *App) Context() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

func (a *App) Init() error {
	ctx := a.Context()

	if err := postgres.EnsureDatabase(ctx, &a.cfg.Database, a.log); err != nil {
		return fmt.Errorf("ensure database: %w", err)
	}

	db, err := postgres.Connect(ctx, &a.cfg.Database, a.log)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	a.db = db

	runner := migrations.NewRunner(a.cfg.Database.MigrationDir, db.DB, a.log)
	if err := runner.Up(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	a.appLog = sharedlogger.NewLogger(a.log.Logger)

	users := authpostgres.NewUserRepository(db.DB)
	sessions := authpostgres.NewSessionRepository(db.DB)
	userRoles := authpostgres.NewUserRoleRepository(db.DB)
	profiles := authpostgres.NewProfileRepository(db.DB)
	settings := adminpostgres.NewSettingRepository(db.DB)
	currencies := adminpostgres.NewCurrencyRepository(db.DB)
	taxes := adminpostgres.NewTaxRepository(db.DB)
	countries := adminpostgres.NewCountryRepository(db.DB)
	auditEvents := adminpostgres.NewAuditEventRepository(db.DB)

	a.users = users

	argonParams := &auth.Argon2Params{
		Memory:      a.cfg.Auth.ArgonMemory,
		Iterations:  a.cfg.Auth.ArgonIterations,
		Parallelism: a.cfg.Auth.ArgonParallelism,
		SaltLength:  a.cfg.Auth.ArgonSaltLength,
		KeyLength:   a.cfg.Auth.ArgonKeyLength,
	}

	a.sessionSvc = auth.NewSessionService(sessions, a.cfg.Auth.SessionTTL, a.appLog)
	a.settingsSvc = administration.NewSettingsService(settings, currencies, taxes, countries, a.appLog)
	a.profileSvc = auth.NewProfileService(profiles, users, a.appLog)
	a.auditSvc = administration.NewAuditService(auditEvents, a.appLog)

	a.authSvc = auth.NewAuthenticationService(users, userRoles, a.sessionSvc, a.auditSvc, argonParams, a.appLog, a.cfg.Auth.MaxLoginAttempts, a.cfg.Auth.LockoutTTL)

	a.log.Info("bindings initialized")
	return nil
}
