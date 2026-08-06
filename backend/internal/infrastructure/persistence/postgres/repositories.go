package postgres

import (
	"database/sql"

	"vfinancy/backend/internal/domain/repositories"
)

// Repositories is the public entry point for the persistence layer.
// The application layer constructs a Repositories value once at
// startup and uses the typed fields to access individual
// repositories. To run a workflow inside a transaction, use
// Repositories.Tx().WithinTransaction(ctx, fn) and rely on the
// unit-of-work stamped on the context.
type Repositories struct {
	Tx repositories.TransactionManager

	Customers     repositories.CustomerRepository
	Suppliers     repositories.SupplierRepository
	Products      repositories.ProductRepository
	Categories    repositories.CategoryRepository
	Brands        repositories.BrandRepository
	Warehouses    repositories.WarehouseRepository
	Currencies    repositories.CurrencyRepository
	Users         repositories.UserRepository
	Roles         repositories.RoleRepository
	Permissions   repositories.PermissionRepository
	UserRoles     repositories.UserRoleRepository

	InventoryBatches    repositories.InventoryBatchRepository
	InventoryMovements  repositories.InventoryMovementRepository

	PurchaseOrders    repositories.PurchaseRepository
	SupplierPayments   repositories.SupplierPaymentRepository
	AccountsPayable    repositories.AccountsPayableRepository

	Sales                 repositories.SalesRepository
	CustomerPayments      repositories.CustomerPaymentRepository
	CustomerAdvances      repositories.CustomerAdvanceRepository
	AccountsReceivable    repositories.AccountsReceivableRepository

	BankAccounts       repositories.BankAccountRepository
	CreditCards         repositories.CreditCardRepository
	BankTransactions    repositories.BankTransactionRepository
	ExchangeRates       repositories.ExchangeRateRepository

	ChartOfAccounts repositories.ChartOfAccountsRepository
	JournalEntries  repositories.JournalRepository
	Ledger          repositories.LedgerRepository

	Sessions    repositories.SessionRepository
	Profiles    repositories.ProfileRepository
	Settings    repositories.SettingRepository
	Countries   repositories.CountryRepository
	Taxes       repositories.TaxRepository
	AuditEvents repositories.AuditEventRepository
}

// NewRepositories builds the Repositories facade bound to db (the
// *sql.DB exposed by infrastructure/database.DB.DB). Every repository
// is initialized in auto-commit mode.
func NewRepositories(db *sql.DB) *Repositories {
	uow := newAutoCommitUoW(db)
	return &Repositories{
		Tx:                  nil, // populated below; use the *TxManager method
		Customers:           uow.Customers(),
		Suppliers:           uow.Suppliers(),
		Products:            uow.Products(),
		Categories:          uow.Categories(),
		Brands:              uow.Brands(),
		Warehouses:          uow.Warehouses(),
		Currencies:          uow.Currencies(),
		Users:               uow.Users(),
		Roles:               uow.Roles(),
		Permissions:         uow.Permissions(),
		UserRoles:           uow.UserRoles(),
		InventoryBatches:    uow.InventoryBatches(),
		InventoryMovements:  uow.InventoryMovements(),
		PurchaseOrders:      uow.PurchaseOrders(),
		SupplierPayments:    uow.SupplierPayments(),
		AccountsPayable:     uow.AccountsPayable(),
		Sales:               uow.Sales(),
		CustomerPayments:    uow.CustomerPayments(),
		CustomerAdvances:    uow.CustomerAdvances(),
		AccountsReceivable:  uow.AccountsReceivable(),
		BankAccounts:       uow.BankAccounts(),
		CreditCards:         uow.CreditCards(),
		BankTransactions:    uow.BankTransactions(),
		ExchangeRates:       uow.ExchangeRates(),
		ChartOfAccounts:     uow.ChartOfAccounts(),
		JournalEntries:      uow.JournalEntries(),
		Ledger:              uow.Ledger(),
		Sessions:            uow.Sessions(),
		Profiles:            uow.Profiles(),
		Settings:            uow.Settings(),
		Countries:           uow.Countries(),
		Taxes:               uow.Taxes(),
		AuditEvents:         uow.AuditEvents(),
	}
}

// SetTransactionManager wires the transaction manager into the facade
// (it is set separately to avoid a cyclic dependency between this
// package's NewRepositories and NewTxManager).
func (r *Repositories) SetTransactionManager(tx repositories.TransactionManager) {
	r.Tx = tx
}
