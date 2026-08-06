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
	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/auth"
	adminservice "vfinancy/backend/internal/application/services/administration"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/application/usecases"
	authuc "vfinancy/backend/internal/application/usecases/auth"
	repospostgres "vfinancy/backend/internal/infrastructure/persistence/postgres"
)

var demoCompanyID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type txBridge struct {
	inner services.TxManager
}

func (t *txBridge) WithinTransaction(ctx context.Context, fn usecases.TxRunner) error {
	return t.inner.WithinTransaction(ctx, services.TxRunner(fn))
}

type App struct {
	ctx context.Context
	db  *database.DB
	cfg *config.Config
	log *logger.Logger

	repos  *repospostgres.Repositories
	appLog *common.Logger

	loginUC    *authuc.LoginUseCase
	logoutUC   *authuc.LogoutUseCase
	changePwUC *authuc.ChangePasswordUseCase

	settingsSvc *adminservice.SettingsService
	profileSvc  *adminservice.ProfileService
	auditSvc    *adminservice.AuditService
	sessionSvc  *auth.SessionService
	authSvc     *auth.AuthenticationService
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

	a.repos = repospostgres.NewRepositories(db.DB)
	txMgr := repospostgres.NewTxManager(db)
	a.repos.SetTransactionManager(txMgr)

	a.appLog = common.NewLogger(a.log.Logger)

	argonParams := &auth.Argon2Params{
		Memory:      a.cfg.Auth.ArgonMemory,
		Iterations:  a.cfg.Auth.ArgonIterations,
		Parallelism: a.cfg.Auth.ArgonParallelism,
		SaltLength:  a.cfg.Auth.ArgonSaltLength,
		KeyLength:   a.cfg.Auth.ArgonKeyLength,
	}

	a.authSvc = auth.NewAuthenticationService(
		a.repos.Users,
		a.repos.UserRoles,
		argonParams,
		a.appLog,
		a.cfg.Auth.MaxLoginAttempts,
		a.cfg.Auth.LockoutTTL,
	)

	a.sessionSvc = auth.NewSessionService(
		a.repos.Sessions,
		a.cfg.Auth.SessionTTL,
		a.appLog,
	)

	a.settingsSvc = adminservice.NewSettingsService(
		a.repos.Settings,
		a.repos.Currencies,
		a.repos.Taxes,
		a.repos.Countries,
		a.appLog,
	)

	a.profileSvc = adminservice.NewProfileService(
		a.repos.Profiles,
		a.repos.Users,
		a.appLog,
	)

	a.auditSvc = adminservice.NewAuditService(
		a.repos.AuditEvents,
		a.appLog,
	)

	svcTx := services.NewTxManager(txMgr)
	ucTx := &txBridge{inner: svcTx}
	base := usecases.NewBase(ucTx, a.appLog)

	a.loginUC = authuc.NewLoginUseCase(base, a.authSvc, a.sessionSvc, a.auditSvc)
	a.logoutUC = authuc.NewLogoutUseCase(base, a.sessionSvc, a.auditSvc)
	a.changePwUC = authuc.NewChangePasswordUseCase(base, a.authSvc, a.sessionSvc, a.auditSvc)

	a.log.Info("bindings initialized")
	return nil
}
