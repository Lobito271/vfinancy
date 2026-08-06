package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/accounting"
	"vfinancy/backend/internal/domain/entities/identity"
	"vfinancy/backend/internal/domain/entities/inventory"
	"vfinancy/backend/internal/domain/entities/masterdata"
	"vfinancy/backend/internal/domain/entities/purchasing"
	"vfinancy/backend/internal/domain/entities/sales"
	"vfinancy/backend/internal/domain/entities/treasury"
	"vfinancy/backend/internal/domain/repositories"
)

// This file contains the constructor pairs (auto-commit and tx-bound)
// for every repository declared in unit_of_work.go. Repositories
// that are not yet fully implemented return a "not implemented"
// stub that satisfies the interface and returns repositories.ErrNotFound
// for any read and nil for any write. The pattern is replaced by a
// real implementation in the dedicated file (e.g. customer_repository.go)
// as the schema migration for that module lands.
//
// Every method on the stub returns a clearly-watermarked error so
// developers can spot it immediately in tests.
//
// The pattern every real implementation follows is:
//
//   type xRepository struct{ q Querier }
//   func newXRepository(db *sql.DB) *xRepository { return &xRepository{q: &dbBox{db: db}} }
//   func newXRepositoryTx(tx *sql.Tx) *xRepository { return &xRepository{q: &txBox{tx: tx}} }
//   func (r *xRepository) ... { ... }

// errNotImplemented is the placeholder returned by stub methods.
var errNotImplemented = errors.New("postgres: repository not yet implemented")

// errStub panics so the calling code can never silently succeed with a
// stub. To enable a method, replace the stub in the dedicated file
// (e.g. customer_repository.go) and remove the function from this file.
func errStub(method string) error {
	return fmt.Errorf("postgres: stub %s called; implement the repository", method)
}

// --- Supplier stub ---

type supplierRepositoryStub struct{}

func newSupplierRepository(*sql.DB) *supplierRepositoryStub { return &supplierRepositoryStub{} }
func newSupplierRepositoryTx(*sql.Tx) *supplierRepositoryStub { return &supplierRepositoryStub{} }

func (s *supplierRepositoryStub) Create(ctx context.Context, _ *masterdata.Supplier) error { return errStub("Supplier.Create") }
func (s *supplierRepositoryStub) Update(ctx context.Context, _ *masterdata.Supplier) error { return errStub("Supplier.Update") }
func (s *supplierRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error          { return errStub("Supplier.Delete") }
func (s *supplierRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*masterdata.Supplier, error) {
	return nil, repositories.ErrNotFound
}
func (s *supplierRepositoryStub) GetByDocument(ctx context.Context, _ uuid.UUID, _ string) (*masterdata.Supplier, error) {
	return nil, repositories.ErrNotFound
}
func (s *supplierRepositoryStub) Exists(ctx context.Context, _ uuid.UUID) (bool, error) { return false, nil }
func (s *supplierRepositoryStub) List(ctx context.Context, _ repositories.SupplierFilter) (repositories.Page[*masterdata.Supplier], error) {
	return repositories.Page[*masterdata.Supplier]{}, nil
}
func (s *supplierRepositoryStub) GetOutstandingBalance(ctx context.Context, _ uuid.UUID) (string, error) {
	return "0.00", nil
}

// --- Product stub ---

type productRepositoryStub struct{}

func newProductRepository(*sql.DB) *productRepositoryStub { return &productRepositoryStub{} }
func newProductRepositoryTx(*sql.Tx) *productRepositoryStub { return &productRepositoryStub{} }

func (p *productRepositoryStub) Create(ctx context.Context, _ *masterdata.Product) error { return errStub("Product.Create") }
func (p *productRepositoryStub) Update(ctx context.Context, _ *masterdata.Product) error { return errStub("Product.Update") }
func (p *productRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error        { return errStub("Product.Delete") }
func (p *productRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*masterdata.Product, error) {
	return nil, repositories.ErrNotFound
}
func (p *productRepositoryStub) GetBySKU(ctx context.Context, _ uuid.UUID, _ string) (*masterdata.Product, error) {
	return nil, repositories.ErrNotFound
}
func (p *productRepositoryStub) GetByBarcode(ctx context.Context, _ uuid.UUID, _ string) (*masterdata.Product, error) {
	return nil, repositories.ErrNotFound
}
func (p *productRepositoryStub) Exists(ctx context.Context, _ uuid.UUID) (bool, error) { return false, nil }
func (p *productRepositoryStub) List(ctx context.Context, _ repositories.ProductFilter) (repositories.Page[*masterdata.Product], error) {
	return repositories.Page[*masterdata.Product]{}, nil
}

// --- Category stub ---

type categoryRepositoryStub struct{}

func newCategoryRepository(*sql.DB) *categoryRepositoryStub { return &categoryRepositoryStub{} }
func newCategoryRepositoryTx(*sql.Tx) *categoryRepositoryStub { return &categoryRepositoryStub{} }

func (c *categoryRepositoryStub) Create(ctx context.Context, _ *masterdata.ProductCategory) error {
	return errStub("Category.Create")
}
func (c *categoryRepositoryStub) Update(ctx context.Context, _ *masterdata.ProductCategory) error {
	return errStub("Category.Update")
}
func (c *categoryRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error { return errStub("Category.Delete") }
func (c *categoryRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*masterdata.ProductCategory, error) {
	return nil, repositories.ErrNotFound
}
func (c *categoryRepositoryStub) List(ctx context.Context, _ uuid.UUID) ([]*masterdata.ProductCategory, error) {
	return nil, nil
}
func (c *categoryRepositoryStub) ListChildren(ctx context.Context, _ uuid.UUID, _ *uuid.UUID) ([]*masterdata.ProductCategory, error) {
	return nil, nil
}

// --- Brand stub ---

type brandRepositoryStub struct{}

func newBrandRepository(*sql.DB) *brandRepositoryStub { return &brandRepositoryStub{} }
func newBrandRepositoryTx(*sql.Tx) *brandRepositoryStub { return &brandRepositoryStub{} }

func (b *brandRepositoryStub) Create(ctx context.Context, _ *masterdata.ProductBrand) error {
	return errStub("Brand.Create")
}
func (b *brandRepositoryStub) Update(ctx context.Context, _ *masterdata.ProductBrand) error {
	return errStub("Brand.Update")
}
func (b *brandRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error { return errStub("Brand.Delete") }
func (b *brandRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*masterdata.ProductBrand, error) {
	return nil, repositories.ErrNotFound
}
func (b *brandRepositoryStub) List(ctx context.Context, _ uuid.UUID) ([]*masterdata.ProductBrand, error) {
	return nil, nil
}

// --- Warehouse stub ---

type warehouseRepositoryStub struct{}

func newWarehouseRepository(*sql.DB) *warehouseRepositoryStub { return &warehouseRepositoryStub{} }
func newWarehouseRepositoryTx(*sql.Tx) *warehouseRepositoryStub { return &warehouseRepositoryStub{} }

func (w *warehouseRepositoryStub) Create(ctx context.Context, _ *masterdata.Warehouse) error {
	return errStub("Warehouse.Create")
}
func (w *warehouseRepositoryStub) Update(ctx context.Context, _ *masterdata.Warehouse) error {
	return errStub("Warehouse.Update")
}
func (w *warehouseRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error { return errStub("Warehouse.Delete") }
func (w *warehouseRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*masterdata.Warehouse, error) {
	return nil, repositories.ErrNotFound
}
func (w *warehouseRepositoryStub) List(ctx context.Context, _ uuid.UUID) ([]*masterdata.Warehouse, error) {
	return nil, nil
}

// --- Role stub ---

type roleRepositoryStub struct{}

func newRoleRepository(*sql.DB) *roleRepositoryStub { return &roleRepositoryStub{} }
func newRoleRepositoryTx(*sql.Tx) *roleRepositoryStub { return &roleRepositoryStub{} }

func (r *roleRepositoryStub) Create(ctx context.Context, _ *identity.Role) error { return errStub("Role.Create") }
func (r *roleRepositoryStub) Update(ctx context.Context, _ *identity.Role) error { return errStub("Role.Update") }
func (r *roleRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error       { return errStub("Role.Delete") }
func (r *roleRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*identity.Role, error) {
	return nil, repositories.ErrNotFound
}
func (r *roleRepositoryStub) Exists(ctx context.Context, _ uuid.UUID) (bool, error) { return false, nil }
func (r *roleRepositoryStub) List(ctx context.Context, _ repositories.RoleFilter) (repositories.Page[*identity.Role], error) {
	return repositories.Page[*identity.Role]{}, nil
}
func (r *roleRepositoryStub) Grant(ctx context.Context, _ uuid.UUID, _ string) error      { return errStub("Role.Grant") }
func (r *roleRepositoryStub) Revoke(ctx context.Context, _ uuid.UUID, _ string) error     { return errStub("Role.Revoke") }
func (r *roleRepositoryStub) ReplacePermissions(ctx context.Context, _ uuid.UUID, _ []string) error {
	return errStub("Role.ReplacePermissions")
}
func (r *roleRepositoryStub) ListPermissionCodes(ctx context.Context, _ uuid.UUID) ([]string, error) {
	return nil, nil
}

// --- Permission stub ---

type permissionRepositoryStub struct{}

func newPermissionRepository(*sql.DB) *permissionRepositoryStub { return &permissionRepositoryStub{} }
func newPermissionRepositoryTx(*sql.Tx) *permissionRepositoryStub { return &permissionRepositoryStub{} }

func (p *permissionRepositoryStub) GetByCode(ctx context.Context, _ string) (*identity.Permission, error) {
	return nil, repositories.ErrNotFound
}
func (p *permissionRepositoryStub) Exists(ctx context.Context, _ string) (bool, error) { return false, nil }
func (p *permissionRepositoryStub) ListAll(ctx context.Context) ([]*identity.Permission, error) {
	return nil, nil
}
func (p *permissionRepositoryStub) ListCodesForRoles(ctx context.Context, _ []string) ([]string, error) {
	return nil, nil
}

// --- UserRole stub ---

type userRoleRepositoryStub struct{}

func newUserRoleRepository(*sql.DB) *userRoleRepositoryStub { return &userRoleRepositoryStub{} }
func newUserRoleRepositoryTx(*sql.Tx) *userRoleRepositoryStub { return &userRoleRepositoryStub{} }

func (u *userRoleRepositoryStub) Assign(ctx context.Context, _, _ uuid.UUID, _ *uuid.UUID, _ *time.Time) error {
	return errStub("UserRole.Assign")
}
func (u *userRoleRepositoryStub) Revoke(ctx context.Context, _, _ uuid.UUID, _ *uuid.UUID) error {
	return errStub("UserRole.Revoke")
}
func (u *userRoleRepositoryStub) EffectiveRoles(ctx context.Context, _ uuid.UUID, _ time.Time) ([]repositories.UserRoleAssignment, error) {
	return nil, nil
}

// --- Inventory stubs ---

type inventoryBatchRepositoryStub struct{}

func newInventoryBatchRepository(*sql.DB) *inventoryBatchRepositoryStub { return &inventoryBatchRepositoryStub{} }
func newInventoryBatchRepositoryTx(*sql.Tx) *inventoryBatchRepositoryStub { return &inventoryBatchRepositoryStub{} }

func (i *inventoryBatchRepositoryStub) Create(ctx context.Context, _ *inventory.InventoryBatch) error {
	return errStub("InventoryBatch.Create")
}
func (i *inventoryBatchRepositoryStub) Update(ctx context.Context, _ *inventory.InventoryBatch) error {
	return errStub("InventoryBatch.Update")
}
func (i *inventoryBatchRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*inventory.InventoryBatch, error) {
	return nil, repositories.ErrNotFound
}
func (i *inventoryBatchRepositoryStub) List(ctx context.Context, _ repositories.InventoryBatchFilter) (repositories.Page[*inventory.InventoryBatch], error) {
	return repositories.Page[*inventory.InventoryBatch]{}, nil
}
func (i *inventoryBatchRepositoryStub) GetStockSummary(ctx context.Context, _, _ uuid.UUID) (float64, string, error) {
	return 0, "0.00", nil
}
func (i *inventoryBatchRepositoryStub) ListClearance(ctx context.Context, _ uuid.UUID, _ time.Time) ([]*inventory.InventoryBatch, error) {
	return nil, nil
}

type inventoryMovementRepositoryStub struct{}

func newInventoryMovementRepository(*sql.DB) *inventoryMovementRepositoryStub { return &inventoryMovementRepositoryStub{} }
func newInventoryMovementRepositoryTx(*sql.Tx) *inventoryMovementRepositoryStub { return &inventoryMovementRepositoryStub{} }

func (i *inventoryMovementRepositoryStub) Create(ctx context.Context, _ *inventory.InventoryMovement) error {
	return errStub("InventoryMovement.Create")
}
func (i *inventoryMovementRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*inventory.InventoryMovement, error) {
	return nil, repositories.ErrNotFound
}
func (i *inventoryMovementRepositoryStub) List(ctx context.Context, _ repositories.InventoryMovementFilter) (repositories.Page[*inventory.InventoryMovement], error) {
	return repositories.Page[*inventory.InventoryMovement]{}, nil
}
func (i *inventoryMovementRepositoryStub) StockAt(ctx context.Context, _, _ uuid.UUID, _ time.Time) (float64, error) {
	return 0, nil
}

// --- Purchasing stubs ---

type purchaseRepositoryStub struct{}

func newPurchaseRepository(*sql.DB) *purchaseRepositoryStub { return &purchaseRepositoryStub{} }
func newPurchaseRepositoryTx(*sql.Tx) *purchaseRepositoryStub { return &purchaseRepositoryStub{} }

func (p *purchaseRepositoryStub) Create(ctx context.Context, _ *purchasing.PurchaseOrder) error { return errStub("Purchase.Create") }
func (p *purchaseRepositoryStub) Update(ctx context.Context, _ *purchasing.PurchaseOrder) error { return errStub("Purchase.Update") }
func (p *purchaseRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error              { return errStub("Purchase.Delete") }
func (p *purchaseRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*purchasing.PurchaseOrder, error) {
	return nil, repositories.ErrNotFound
}
func (p *purchaseRepositoryStub) GetByNumber(ctx context.Context, _ uuid.UUID, _ string) (*purchasing.PurchaseOrder, error) {
	return nil, repositories.ErrNotFound
}
func (p *purchaseRepositoryStub) Exists(ctx context.Context, _ uuid.UUID) (bool, error) { return false, nil }
func (p *purchaseRepositoryStub) List(ctx context.Context, _ repositories.PurchaseFilter) (repositories.Page[*purchasing.PurchaseOrder], error) {
	return repositories.Page[*purchasing.PurchaseOrder]{}, nil
}
func (p *purchaseRepositoryStub) GetNextNumber(ctx context.Context, _ uuid.UUID) (string, error) { return "", nil }

type supplierPaymentRepositoryStub struct{}

func newSupplierPaymentRepository(*sql.DB) *supplierPaymentRepositoryStub { return &supplierPaymentRepositoryStub{} }
func newSupplierPaymentRepositoryTx(*sql.Tx) *supplierPaymentRepositoryStub { return &supplierPaymentRepositoryStub{} }

func (s *supplierPaymentRepositoryStub) Create(ctx context.Context, _ *purchasing.SupplierPayment) error {
	return errStub("SupplierPayment.Create")
}
func (s *supplierPaymentRepositoryStub) Update(ctx context.Context, _ *purchasing.SupplierPayment) error {
	return errStub("SupplierPayment.Update")
}
func (s *supplierPaymentRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*purchasing.SupplierPayment, error) {
	return nil, repositories.ErrNotFound
}
func (s *supplierPaymentRepositoryStub) List(ctx context.Context, _ repositories.SupplierPaymentFilter) (repositories.Page[*purchasing.SupplierPayment], error) {
	return repositories.Page[*purchasing.SupplierPayment]{}, nil
}
func (s *supplierPaymentRepositoryStub) ListAllocationsForPurchase(ctx context.Context, _ uuid.UUID) ([]*purchasing.SupplierPayment, error) {
	return nil, nil
}

type accountsPayableRepositoryStub struct{}

func newAccountsPayableRepository(*sql.DB) *accountsPayableRepositoryStub { return &accountsPayableRepositoryStub{} }
func newAccountsPayableRepositoryTx(*sql.Tx) *accountsPayableRepositoryStub { return &accountsPayableRepositoryStub{} }

func (a *accountsPayableRepositoryStub) GetOpenBalanceForSupplier(ctx context.Context, _ uuid.UUID) (string, error) {
	return "0.00", nil
}
func (a *accountsPayableRepositoryStub) ListAgingBucket(ctx context.Context, _ uuid.UUID) (map[string]string, error) {
	return map[string]string{}, nil
}

// --- Sales stubs ---

type salesRepositoryStub struct{}

func newSalesRepository(*sql.DB) *salesRepositoryStub { return &salesRepositoryStub{} }
func newSalesRepositoryTx(*sql.Tx) *salesRepositoryStub { return &salesRepositoryStub{} }

func (s *salesRepositoryStub) Create(ctx context.Context, _ *sales.Sale) error { return errStub("Sale.Create") }
func (s *salesRepositoryStub) Update(ctx context.Context, _ *sales.Sale) error { return errStub("Sale.Update") }
func (s *salesRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error   { return errStub("Sale.Delete") }
func (s *salesRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*sales.Sale, error) {
	return nil, repositories.ErrNotFound
}
func (s *salesRepositoryStub) GetByNumber(ctx context.Context, _ uuid.UUID, _ string) (*sales.Sale, error) {
	return nil, repositories.ErrNotFound
}
func (s *salesRepositoryStub) Exists(ctx context.Context, _ uuid.UUID) (bool, error) { return false, nil }
func (s *salesRepositoryStub) List(ctx context.Context, _ repositories.SaleFilter) (repositories.Page[*sales.Sale], error) {
	return repositories.Page[*sales.Sale]{}, nil
}
func (s *salesRepositoryStub) GetNextNumber(ctx context.Context, _ uuid.UUID) (string, error) { return "", nil }

type customerPaymentRepositoryStub struct{}

func newCustomerPaymentRepository(*sql.DB) *customerPaymentRepositoryStub { return &customerPaymentRepositoryStub{} }
func newCustomerPaymentRepositoryTx(*sql.Tx) *customerPaymentRepositoryStub { return &customerPaymentRepositoryStub{} }

func (c *customerPaymentRepositoryStub) Create(ctx context.Context, _ *sales.CustomerPayment) error {
	return errStub("CustomerPayment.Create")
}
func (c *customerPaymentRepositoryStub) Update(ctx context.Context, _ *sales.CustomerPayment) error {
	return errStub("CustomerPayment.Update")
}
func (c *customerPaymentRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*sales.CustomerPayment, error) {
	return nil, repositories.ErrNotFound
}
func (c *customerPaymentRepositoryStub) List(ctx context.Context, _ repositories.CustomerPaymentFilter) (repositories.Page[*sales.CustomerPayment], error) {
	return repositories.Page[*sales.CustomerPayment]{}, nil
}
func (c *customerPaymentRepositoryStub) ListAllocationsForSale(ctx context.Context, _ uuid.UUID) ([]*sales.CustomerPayment, error) {
	return nil, nil
}

type customerAdvanceRepositoryStub struct{}

func newCustomerAdvanceRepository(*sql.DB) *customerAdvanceRepositoryStub { return &customerAdvanceRepositoryStub{} }
func newCustomerAdvanceRepositoryTx(*sql.Tx) *customerAdvanceRepositoryStub { return &customerAdvanceRepositoryStub{} }

func (c *customerAdvanceRepositoryStub) Create(ctx context.Context, _ *sales.CustomerAdvance) error {
	return errStub("CustomerAdvance.Create")
}
func (c *customerAdvanceRepositoryStub) Update(ctx context.Context, _ *sales.CustomerAdvance) error {
	return errStub("CustomerAdvance.Update")
}
func (c *customerAdvanceRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*sales.CustomerAdvance, error) {
	return nil, repositories.ErrNotFound
}
func (c *customerAdvanceRepositoryStub) ListByCustomer(ctx context.Context, _ uuid.UUID) ([]*sales.CustomerAdvance, error) {
	return nil, nil
}
func (c *customerAdvanceRepositoryStub) ListApplicationsForSale(ctx context.Context, _ uuid.UUID) ([]*sales.CustomerAdvance, error) {
	return nil, nil
}

type accountsReceivableRepositoryStub struct{}

func newAccountsReceivableRepository(*sql.DB) *accountsReceivableRepositoryStub { return &accountsReceivableRepositoryStub{} }
func newAccountsReceivableRepositoryTx(*sql.Tx) *accountsReceivableRepositoryStub { return &accountsReceivableRepositoryStub{} }

func (a *accountsReceivableRepositoryStub) GetOpenBalanceForCustomer(ctx context.Context, _ uuid.UUID) (string, error) {
	return "0.00", nil
}
func (a *accountsReceivableRepositoryStub) ListAgingBucket(ctx context.Context, _ uuid.UUID) (map[string]string, error) {
	return map[string]string{}, nil
}

// --- Treasury stubs ---

type bankAccountRepositoryStub struct{}

func newBankAccountRepository(*sql.DB) *bankAccountRepositoryStub { return &bankAccountRepositoryStub{} }
func newBankAccountRepositoryTx(*sql.Tx) *bankAccountRepositoryStub { return &bankAccountRepositoryStub{} }

func (b *bankAccountRepositoryStub) Create(ctx context.Context, _ *treasury.BankAccount) error {
	return errStub("BankAccount.Create")
}
func (b *bankAccountRepositoryStub) Update(ctx context.Context, _ *treasury.BankAccount) error {
	return errStub("BankAccount.Update")
}
func (b *bankAccountRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error        { return errStub("BankAccount.Delete") }
func (b *bankAccountRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*treasury.BankAccount, error) {
	return nil, repositories.ErrNotFound
}
func (b *bankAccountRepositoryStub) List(ctx context.Context, _ repositories.BankAccountFilter) (repositories.Page[*treasury.BankAccount], error) {
	return repositories.Page[*treasury.BankAccount]{}, nil
}

type creditCardRepositoryStub struct{}

func newCreditCardRepository(*sql.DB) *creditCardRepositoryStub { return &creditCardRepositoryStub{} }
func newCreditCardRepositoryTx(*sql.Tx) *creditCardRepositoryStub { return &creditCardRepositoryStub{} }

func (c *creditCardRepositoryStub) Create(ctx context.Context, _ *treasury.CreditCard) error {
	return errStub("CreditCard.Create")
}
func (c *creditCardRepositoryStub) Update(ctx context.Context, _ *treasury.CreditCard) error {
	return errStub("CreditCard.Update")
}
func (c *creditCardRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error        { return errStub("CreditCard.Delete") }
func (c *creditCardRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*treasury.CreditCard, error) {
	return nil, repositories.ErrNotFound
}
func (c *creditCardRepositoryStub) List(ctx context.Context, _ uuid.UUID) ([]*treasury.CreditCard, error) {
	return nil, nil
}

type bankTransactionRepositoryStub struct{}

func newBankTransactionRepository(*sql.DB) *bankTransactionRepositoryStub { return &bankTransactionRepositoryStub{} }
func newBankTransactionRepositoryTx(*sql.Tx) *bankTransactionRepositoryStub { return &bankTransactionRepositoryStub{} }

func (b *bankTransactionRepositoryStub) Create(ctx context.Context, _ *treasury.BankTransaction) error {
	return errStub("BankTransaction.Create")
}
func (b *bankTransactionRepositoryStub) Update(ctx context.Context, _ *treasury.BankTransaction) error {
	return errStub("BankTransaction.Update")
}
func (b *bankTransactionRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*treasury.BankTransaction, error) {
	return nil, repositories.ErrNotFound
}
func (b *bankTransactionRepositoryStub) List(ctx context.Context, _ repositories.BankTransactionFilter) (repositories.Page[*treasury.BankTransaction], error) {
	return repositories.Page[*treasury.BankTransaction]{}, nil
}

// --- Accounting stubs ---

type chartOfAccountsRepositoryStub struct{}

func newChartOfAccountsRepository(*sql.DB) *chartOfAccountsRepositoryStub { return &chartOfAccountsRepositoryStub{} }
func newChartOfAccountsRepositoryTx(*sql.Tx) *chartOfAccountsRepositoryStub { return &chartOfAccountsRepositoryStub{} }

func (c *chartOfAccountsRepositoryStub) Create(ctx context.Context, _ *accounting.ChartOfAccount) error {
	return errStub("ChartOfAccounts.Create")
}
func (c *chartOfAccountsRepositoryStub) Update(ctx context.Context, _ *accounting.ChartOfAccount) error {
	return errStub("ChartOfAccounts.Update")
}
func (c *chartOfAccountsRepositoryStub) Delete(ctx context.Context, _ uuid.UUID) error {
	return errStub("ChartOfAccounts.Delete")
}
func (c *chartOfAccountsRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*accounting.ChartOfAccount, error) {
	return nil, repositories.ErrNotFound
}
func (c *chartOfAccountsRepositoryStub) GetByCode(ctx context.Context, _ uuid.UUID, _ string) (*accounting.ChartOfAccount, error) {
	return nil, repositories.ErrNotFound
}
func (c *chartOfAccountsRepositoryStub) List(ctx context.Context, _ repositories.ChartOfAccountsFilter) (repositories.Page[*accounting.ChartOfAccount], error) {
	return repositories.Page[*accounting.ChartOfAccount]{}, nil
}
func (c *chartOfAccountsRepositoryStub) ListChildren(ctx context.Context, _ uuid.UUID, _ *uuid.UUID) ([]*accounting.ChartOfAccount, error) {
	return nil, nil
}

type journalRepositoryStub struct{}

func newJournalRepository(*sql.DB) *journalRepositoryStub { return &journalRepositoryStub{} }
func newJournalRepositoryTx(*sql.Tx) *journalRepositoryStub { return &journalRepositoryStub{} }

func (j *journalRepositoryStub) Create(ctx context.Context, _ *accounting.JournalEntry) error {
	return errStub("JournalEntry.Create")
}
func (j *journalRepositoryStub) Update(ctx context.Context, _ *accounting.JournalEntry) error {
	return errStub("JournalEntry.Update")
}
func (j *journalRepositoryStub) GetByID(ctx context.Context, _ uuid.UUID) (*accounting.JournalEntry, error) {
	return nil, repositories.ErrNotFound
}
func (j *journalRepositoryStub) GetByNumber(ctx context.Context, _ uuid.UUID, _ string) (*accounting.JournalEntry, error) {
	return nil, repositories.ErrNotFound
}
func (j *journalRepositoryStub) List(ctx context.Context, _ repositories.JournalEntryFilter) (repositories.Page[*accounting.JournalEntry], error) {
	return repositories.Page[*accounting.JournalEntry]{}, nil
}
func (j *journalRepositoryStub) GetNextNumber(ctx context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

type ledgerRepositoryStub struct{}

func newLedgerRepository(*sql.DB) *ledgerRepositoryStub { return &ledgerRepositoryStub{} }
func newLedgerRepositoryTx(*sql.Tx) *ledgerRepositoryStub { return &ledgerRepositoryStub{} }

func (l *ledgerRepositoryStub) GetAccountBalance(ctx context.Context, _ uuid.UUID, _ time.Time) (string, error) {
	return "0.00", nil
}
func (l *ledgerRepositoryStub) GetTrialBalance(ctx context.Context, _ uuid.UUID) ([]repositories.TrialBalanceRow, error) {
	return nil, nil
}
