package bindings

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/infrastructure/sqlite"
	"vfinancy/backend/internal/features/accounting"
	accountingpostgres "vfinancy/backend/internal/features/accounting/postgres"
	"vfinancy/backend/internal/features/administration"
	adminpostgres "vfinancy/backend/internal/features/administration/postgres"
	"vfinancy/backend/internal/features/auth"
	authpostgres "vfinancy/backend/internal/features/auth/postgres"
	"vfinancy/backend/internal/features/customer"
	customerpostgres "vfinancy/backend/internal/features/customer/postgres"
	"vfinancy/backend/internal/features/customerpayments"
	"vfinancy/backend/internal/features/inventory"
	inventorypostgres "vfinancy/backend/internal/features/inventory/postgres"
	"vfinancy/backend/internal/features/product"
	productpostgres "vfinancy/backend/internal/features/product/postgres"
	"vfinancy/backend/internal/features/purchasing"
	purchasingpostgres "vfinancy/backend/internal/features/purchasing/postgres"
	"vfinancy/backend/internal/features/sales"
	salespostgres "vfinancy/backend/internal/features/sales/postgres"
	"vfinancy/backend/internal/features/supplier"
	supplierpostgres "vfinancy/backend/internal/features/supplier/postgres"
	"vfinancy/backend/internal/features/sync"
	syncpostgres "vfinancy/backend/internal/features/sync/postgres"
	"vfinancy/backend/internal/features/treasury"
	treasurypostgres "vfinancy/backend/internal/features/treasury/postgres"
)

var demoCompanyID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type App struct {
	ctx context.Context
	db  *database.DB
	cfg *config.Config
	log *logger.Logger

	migrationsFS fs.FS

	authSvc     *auth.AuthenticationService
	settingsSvc *administration.SettingsService
	profileSvc  *auth.ProfileService
	auditSvc    *administration.AuditService
	sessionSvc  *auth.SessionService

	treasurySvc    *treasury.TreasuryService
	salesSvc       *sales.SalesService
	paymentSvc     *customerpayments.CustomerPaymentService
	inventorySvc   *inventory.InventoryService
	accountingSvc  *accounting.AccountingService
	purchasingSvc  *purchasing.PurchasingService
	customersSvc   *customer.CustomerService
	productsSvc    *product.ProductService
	suppliersSvc   *supplier.SupplierService

	syncCancel context.CancelFunc

	users auth.UserRepository
}

func New(cfg *config.Config, log *logger.Logger, migrationsFS fs.FS) *App {
	return &App{cfg: cfg, log: log, migrationsFS: migrationsFS}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(ctx context.Context) {
	if a.syncCancel != nil {
		a.syncCancel()
	}
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

	if a.cfg.Database.Driver == "postgres" {
		return fmt.Errorf("bindings: postgres driver is not supported for the desktop runtime; use sqlite (DB_DRIVER=sqlite)")
	}

	db, err := sqlite.Open(a.cfg.Database.Path, database.Options{
		MaxOpenConns:    a.cfg.Database.MaxOpen,
		MaxIdleConns:    a.cfg.Database.MaxIdle,
		ConnMaxLifetime: a.cfg.Database.MaxLifetime,
	})
	if err != nil {
		return fmt.Errorf("connect sqlite: %w", err)
	}
	a.db = db
	persistence.SetDialect(persistence.DialectSQLite)

	runner := migrations.NewRunnerFS(a.migrationsFS, db.DB, a.log, "sqlite")
	if err := runner.Up(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	users := authpostgres.NewUserRepository(db.DB)
	sessions := authpostgres.NewSessionRepository(db.DB)
	userRoles := authpostgres.NewUserRoleRepository(db.DB)
	profiles := authpostgres.NewProfileRepository(db.DB)
	settings := adminpostgres.NewSettingRepository(db.DB)
	currencies := adminpostgres.NewCurrencyRepository(db.DB)
	taxes := adminpostgres.NewTaxRepository(db.DB)
	countries := adminpostgres.NewCountryRepository(db.DB)
	auditEvents := adminpostgres.NewAuditEventRepository(db.DB)

	txm := persistence.NewTxManager(db)
	customers := customerpostgres.NewCustomerRepository(db.DB)
	products := productpostgres.NewProductRepository(db.DB)
	suppliers := supplierpostgres.NewSupplierRepository(db.DB)

	bankAccounts := treasurypostgres.NewBankAccountRepository(db.DB)
	creditCards := treasurypostgres.NewCreditCardRepository(db.DB)
	bankTransactions := treasurypostgres.NewBankTransactionRepository(db.DB)
	exchangeRates := treasurypostgres.NewExchangeRateRepository(db.DB)

	batches := inventorypostgres.NewInventoryBatchRepository(db.DB)
	movements := inventorypostgres.NewInventoryMovementRepository(db.DB)
	warehouseResolver := inventorypostgres.NewWarehouseResolver(db.DB)
	productClassifier := inventorypostgres.NewProductClassifier(db.DB)

	orders := salespostgres.NewSaleRepository(db.DB)
	payments := salespostgres.NewCustomerPaymentRepository(db.DB)
	advances := salespostgres.NewCustomerAdvanceRepository(db.DB)

	purchaseOrders := purchasingpostgres.NewPurchaseRepository(db.DB)
	supplierPayments := purchasingpostgres.NewSupplierPaymentRepository(db.DB)

	journalEntries := accountingpostgres.NewJournalRepository(db.DB)
	chartOfAccounts := accountingpostgres.NewChartOfAccountRepository(db.DB)
	ledger := accountingpostgres.NewLedgerRepository(db.DB)
	fiscalPeriods := accountingpostgres.NewFiscalPeriodRepository(db.DB)

	a.users = users

	argonParams := &auth.Argon2Params{
		Memory:      a.cfg.Auth.ArgonMemory,
		Iterations:  a.cfg.Auth.ArgonIterations,
		Parallelism: a.cfg.Auth.ArgonParallelism,
		SaltLength:  a.cfg.Auth.ArgonSaltLength,
		KeyLength:   a.cfg.Auth.ArgonKeyLength,
	}

	a.sessionSvc = auth.NewSessionService(sessions, a.cfg.Auth.SessionTTL, a.log)
	a.settingsSvc = administration.NewSettingsService(settings, currencies, taxes, countries, a.log)
	a.profileSvc = auth.NewProfileService(profiles, users, a.log)
	a.auditSvc = administration.NewAuditService(auditEvents, a.log)

	a.authSvc = auth.NewAuthenticationService(users, userRoles, a.sessionSvc, a.auditSvc, argonParams, a.log, a.cfg.Auth.MaxLoginAttempts, a.cfg.Auth.LockoutTTL)

	a.treasurySvc = treasury.New(bankAccounts, creditCards, bankTransactions, exchangeRates, txm, a.log)
	a.inventorySvc = inventory.New(batches, movements, warehouseResolver, productClassifier, txm, a.log)
	a.salesSvc = sales.New(orders, customers, a.inventorySvc, productClassifier, txm, a.log)
	a.paymentSvc = customerpayments.New(payments, advances, orders, customers, txm, a.log)
	a.accountingSvc = accounting.New(journalEntries, chartOfAccounts, ledger, fiscalPeriods, txm, a.log)
	a.purchasingSvc = purchasing.New(purchaseOrders, supplierPayments, a.inventorySvc, txm, a.log)
	a.customersSvc = customer.New(customers, txm, a.log)
	a.productsSvc = product.New(products, txm, a.log)
	a.suppliersSvc = supplier.New(suppliers, txm, a.log)

	if err := a.seedAuthData(ctx, users, userRoles); err != nil {
		a.log.Warn("auth seeder failed; continuing without demo admin", "error", err.Error())
	}

	a.startSyncWorker(ctx)

	a.log.Info("bindings initialized")
	return nil
}

// startSyncWorker launches the background replication loop when sync is
// enabled. It is best-effort: any failure is logged and the worker
// retries on the next tick, so the app keeps running fully offline.
func (a *App) startSyncWorker(ctx context.Context) {
	if !a.cfg.Sync.Enabled {
		return
	}
	repo := syncpostgres.NewSyncRepository(a.db.DB)
	client := sync.NewHTTPClient(a.cfg.Sync.ServerURL, a.cfg.Sync.APIKey)
	svc := sync.NewService(repo, client, a.log.Logger, "vfinancy-desktop", "desktop")

	wctx, cancel := context.WithCancel(ctx)
	a.syncCancel = cancel
	go func() {
		run := func() {
			if err := svc.RunOnce(wctx); err != nil {
				a.log.Warn("sync: run failed", "error", err.Error())
			}
		}
		run()
		ticker := time.NewTicker(a.cfg.Sync.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-wctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
	a.log.Info("sync worker started",
		"server", a.cfg.Sync.ServerURL,
		"interval", a.cfg.Sync.PollInterval.String(),
	)
}
