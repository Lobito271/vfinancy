package postgres

import (
	"database/sql"

	"vfinancy/backend/internal/domain/repositories"
)

// unitOfWork is the repositories.UnitOfWork implementation. It
// hands out either a "real" (auto-commit) repository bound to *sql.DB
// or a "transaction-bound" repository bound to *sql.Tx, depending on
// whether the constructor was called from a transaction manager or
// the application container.
type unitOfWork struct {
	db     *sql.DB
	tx     *sql.Tx
	hasTx  bool

	// The repos are populated lazily so callers pay only for what
	// they use. They are safe to share within a single request
	// because each holds no state — every method takes the ctx.
}

// newAutoCommitUoW returns a UoW whose repositories run against the
// given *sql.DB (auto-commit mode). This is the UoW used outside of
// a TransactionManager.WithinTransaction block.
func newAutoCommitUoW(db *sql.DB) *unitOfWork {
	return &unitOfWork{db: db}
}

// newUnitOfWork returns a UoW whose repositories run against the
// given txBox (a transaction-bound Querier).
func newUnitOfWork(tb *txBox) *unitOfWork {
	return &unitOfWork{tx: tb.tx, hasTx: true}
}

// execOn runs the given function against either the tx or the db
// depending on whether this UoW is in transaction mode. The closure
// receives a Querier (an interface satisfied by both *sql.DB and
// *sql.Tx).
func (u *unitOfWork) execOn(fn func(q Querier) error) error {
	if u.hasTx {
		return fn(&txBox{tx: u.tx})
	}
	return fn(&dbBox{db: u.db})
}

// --- Repositories (lazy, one per bounded context) ---

func (u *unitOfWork) Customers() repositories.CustomerRepository {
	var q Querier
	if u.hasTx {
		q = &txBox{tx: u.tx}
	} else {
		q = &dbBox{db: u.db}
	}
	return &customerRepository{q: q}
}

func (u *unitOfWork) Suppliers() repositories.SupplierRepository {
	if u.hasTx {
		return newSupplierRepositoryTx(u.tx)
	}
	return newSupplierRepository(u.db)
}

func (u *unitOfWork) Products() repositories.ProductRepository {
	if u.hasTx {
		return newProductRepositoryTx(u.tx)
	}
	return newProductRepository(u.db)
}

func (u *unitOfWork) Categories() repositories.CategoryRepository {
	if u.hasTx {
		return newCategoryRepositoryTx(u.tx)
	}
	return newCategoryRepository(u.db)
}

func (u *unitOfWork) Brands() repositories.BrandRepository {
	if u.hasTx {
		return newBrandRepositoryTx(u.tx)
	}
	return newBrandRepository(u.db)
}

func (u *unitOfWork) Warehouses() repositories.WarehouseRepository {
	if u.hasTx {
		return newWarehouseRepositoryTx(u.tx)
	}
	return newWarehouseRepository(u.db)
}

func (u *unitOfWork) Currencies() repositories.CurrencyRepository {
	if u.hasTx {
		return newCurrencyRepositoryTx(u.tx)
	}
	return newCurrencyRepository(u.db)
}

func (u *unitOfWork) Users() repositories.UserRepository {
	if u.hasTx {
		return newUserRepositoryTx(u.tx)
	}
	return newUserRepository(u.db)
}

func (u *unitOfWork) Roles() repositories.RoleRepository {
	if u.hasTx {
		return newRoleRepositoryTx(u.tx)
	}
	return newRoleRepository(u.db)
}

func (u *unitOfWork) Permissions() repositories.PermissionRepository {
	if u.hasTx {
		return newPermissionRepositoryTx(u.tx)
	}
	return newPermissionRepository(u.db)
}

func (u *unitOfWork) UserRoles() repositories.UserRoleRepository {
	if u.hasTx {
		return newUserRoleRepositoryTx(u.tx)
	}
	return newUserRoleRepository(u.db)
}

func (u *unitOfWork) InventoryBatches() repositories.InventoryBatchRepository {
	if u.hasTx {
		return newInventoryBatchRepositoryTx(u.tx)
	}
	return newInventoryBatchRepository(u.db)
}

func (u *unitOfWork) InventoryMovements() repositories.InventoryMovementRepository {
	if u.hasTx {
		return newInventoryMovementRepositoryTx(u.tx)
	}
	return newInventoryMovementRepository(u.db)
}

func (u *unitOfWork) PurchaseOrders() repositories.PurchaseRepository {
	if u.hasTx {
		return newPurchaseRepositoryTx(u.tx)
	}
	return newPurchaseRepository(u.db)
}

func (u *unitOfWork) SupplierPayments() repositories.SupplierPaymentRepository {
	if u.hasTx {
		return newSupplierPaymentRepositoryTx(u.tx)
	}
	return newSupplierPaymentRepository(u.db)
}

func (u *unitOfWork) AccountsPayable() repositories.AccountsPayableRepository {
	if u.hasTx {
		return newAccountsPayableRepositoryTx(u.tx)
	}
	return newAccountsPayableRepository(u.db)
}

func (u *unitOfWork) Sales() repositories.SalesRepository {
	if u.hasTx {
		return newSalesRepositoryTx(u.tx)
	}
	return newSalesRepository(u.db)
}

func (u *unitOfWork) CustomerPayments() repositories.CustomerPaymentRepository {
	if u.hasTx {
		return newCustomerPaymentRepositoryTx(u.tx)
	}
	return newCustomerPaymentRepository(u.db)
}

func (u *unitOfWork) CustomerAdvances() repositories.CustomerAdvanceRepository {
	if u.hasTx {
		return newCustomerAdvanceRepositoryTx(u.tx)
	}
	return newCustomerAdvanceRepository(u.db)
}

func (u *unitOfWork) AccountsReceivable() repositories.AccountsReceivableRepository {
	if u.hasTx {
		return newAccountsReceivableRepositoryTx(u.tx)
	}
	return newAccountsReceivableRepository(u.db)
}

func (u *unitOfWork) BankAccounts() repositories.BankAccountRepository {
	if u.hasTx {
		return newBankAccountRepositoryTx(u.tx)
	}
	return newBankAccountRepository(u.db)
}

func (u *unitOfWork) CreditCards() repositories.CreditCardRepository {
	if u.hasTx {
		return newCreditCardRepositoryTx(u.tx)
	}
	return newCreditCardRepository(u.db)
}

func (u *unitOfWork) BankTransactions() repositories.BankTransactionRepository {
	if u.hasTx {
		return newBankTransactionRepositoryTx(u.tx)
	}
	return newBankTransactionRepository(u.db)
}

func (u *unitOfWork) ExchangeRates() repositories.ExchangeRateRepository {
	if u.hasTx {
		return newExchangeRateRepositoryTx(u.tx)
	}
	return newExchangeRateRepository(u.db)
}

func (u *unitOfWork) ChartOfAccounts() repositories.ChartOfAccountsRepository {
	if u.hasTx {
		return newChartOfAccountsRepositoryTx(u.tx)
	}
	return newChartOfAccountsRepository(u.db)
}

func (u *unitOfWork) JournalEntries() repositories.JournalRepository {
	if u.hasTx {
		return newJournalRepositoryTx(u.tx)
	}
	return newJournalRepository(u.db)
}

func (u *unitOfWork) Ledger() repositories.LedgerRepository {
	if u.hasTx {
		return newLedgerRepositoryTx(u.tx)
	}
	return newLedgerRepository(u.db)
}

func (u *unitOfWork) Sessions() repositories.SessionRepository {
	if u.hasTx {
		return newSessionRepositoryTx(u.tx)
	}
	return newSessionRepository(u.db)
}

func (u *unitOfWork) Profiles() repositories.ProfileRepository {
	if u.hasTx {
		return newProfileRepositoryTx(u.tx)
	}
	return newProfileRepository(u.db)
}

func (u *unitOfWork) Settings() repositories.SettingRepository {
	if u.hasTx {
		return newSettingRepositoryTx(u.tx)
	}
	return newSettingRepository(u.db)
}

func (u *unitOfWork) Countries() repositories.CountryRepository {
	if u.hasTx {
		return newCountryRepositoryTx(u.tx)
	}
	return newCountryRepository(u.db)
}

func (u *unitOfWork) Taxes() repositories.TaxRepository {
	if u.hasTx {
		return newTaxRepositoryTx(u.tx)
	}
	return newTaxRepository(u.db)
}

func (u *unitOfWork) AuditEvents() repositories.AuditEventRepository {
	if u.hasTx {
		return newAuditEventRepositoryTx(u.tx)
	}
	return newAuditEventRepository(u.db)
}
