package repositories

import (
	"context"
)

// UnitOfWork is the application's "I'm currently inside a transaction"
// handle. It exposes one method per aggregate that participates in a
// transaction, returning a repository bound to the active transaction
// rather than the default (auto-commit) connection.
//
// The default UnitOfWork implementation simply returns the regular
// repositories; the transaction-bound implementation wraps each
// repository so its SQL runs on the active transaction. The
// application layer calls UnitOfWork inside a TransactionManager
// callback to keep the workflow on a single transaction.
type UnitOfWork interface {
	Customers() CustomerRepository
	Suppliers() SupplierRepository
	Products() ProductRepository
	Categories() CategoryRepository
	Brands() BrandRepository
	Warehouses() WarehouseRepository
	Currencies() CurrencyRepository
	Users() UserRepository
	Roles() RoleRepository
	Permissions() PermissionRepository
	UserRoles() UserRoleRepository

	InventoryBatches() InventoryBatchRepository
	InventoryMovements() InventoryMovementRepository

	PurchaseOrders() PurchaseRepository
	SupplierPayments() SupplierPaymentRepository
	AccountsPayable() AccountsPayableRepository

	Sales() SalesRepository
	CustomerPayments() CustomerPaymentRepository
	CustomerAdvances() CustomerAdvanceRepository
	AccountsReceivable() AccountsReceivableRepository

	BankAccounts() BankAccountRepository
	CreditCards() CreditCardRepository
	BankTransactions() BankTransactionRepository
	ExchangeRates() ExchangeRateRepository

	ChartOfAccounts() ChartOfAccountsRepository
	JournalEntries() JournalRepository
	Ledger() LedgerRepository

	Sessions() SessionRepository
	Profiles() ProfileRepository
	Settings() SettingRepository
	Countries() CountryRepository
	Taxes() TaxRepository
	AuditEvents() AuditEventRepository
}

// WithUnitOfWork is a small helper that the application layer uses to
// run a closure against the active UnitOfWork. The default
// implementation, in the absence of an explicit WithUnitOfWork call,
// simply returns ctx unchanged; the transaction-bound implementation
// stamps the context with the active UoW.
//
// Repositories that need to participate in a transaction call
// UnitOfWorkFromContext(ctx) and use the result. The default behavior
// (no UoW) is to fall back to the regular repositories injected at
// construction time.
type ctxKey int

const (
	uowKey ctxKey = iota
)

// ContextWithUnitOfWork returns a new context that carries the given
// UnitOfWork. The transaction manager uses this internally.
func ContextWithUnitOfWork(parent context.Context, uow UnitOfWork) context.Context {
	return context.WithValue(parent, uowKey, uow)
}

// UnitOfWorkFromContext returns the UnitOfWork stored in the context,
// or nil if none is present.
func UnitOfWorkFromContext(ctx context.Context) UnitOfWork {
	if v, ok := ctx.Value(uowKey).(UnitOfWork); ok {
		return v
	}
	return nil
}
