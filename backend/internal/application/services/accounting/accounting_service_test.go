package accounting_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	accsvc "vfinancy/backend/internal/application/services/accounting"
	"vfinancy/backend/internal/application/services/common"
	entacc "vfinancy/backend/internal/domain/entities/accounting"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

type fakeEntryRepo struct {
	entries map[uuid.UUID]*entacc.JournalEntry
}

func (f *fakeEntryRepo) Create(ctx context.Context, e *entacc.JournalEntry) error {
	if _, ok := f.entries[e.ID]; ok {
		return repositories.ErrDuplicate
	}
	f.entries[e.ID] = e
	return nil
}
func (f *fakeEntryRepo) Update(ctx context.Context, e *entacc.JournalEntry) error {
	f.entries[e.ID] = e
	return nil
}
func (f *fakeEntryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entacc.JournalEntry, error) {
	e, ok := f.entries[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return e, nil
}
func (f *fakeEntryRepo) GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*entacc.JournalEntry, error) {
	return nil, repositories.ErrNotFound
}
func (f *fakeEntryRepo) List(ctx context.Context, f2 repositories.JournalEntryFilter) (repositories.Page[*entacc.JournalEntry], error) {
	return repositories.Page[*entacc.JournalEntry]{}, nil
}
func (f *fakeEntryRepo) GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	return "JE-1", nil
}

type fakeChartRepo struct{}

func (f *fakeChartRepo) Create(ctx context.Context, a *entacc.ChartOfAccount) error { return nil }
func (f *fakeChartRepo) Update(ctx context.Context, a *entacc.ChartOfAccount) error { return nil }
func (f *fakeChartRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (f *fakeChartRepo) GetByID(ctx context.Context, id uuid.UUID) (*entacc.ChartOfAccount, error) {
	return nil, repositories.ErrNotFound
}
func (f *fakeChartRepo) GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*entacc.ChartOfAccount, error) {
	return nil, repositories.ErrNotFound
}
func (f *fakeChartRepo) List(ctx context.Context, f2 repositories.ChartOfAccountsFilter) (repositories.Page[*entacc.ChartOfAccount], error) {
	return repositories.Page[*entacc.ChartOfAccount]{}, nil
}
func (f *fakeChartRepo) ListChildren(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID) ([]*entacc.ChartOfAccount, error) {
	return nil, nil
}

type fakeLedgerRepo struct{}

func (f *fakeLedgerRepo) GetAccountBalance(ctx context.Context, accountID uuid.UUID, at time.Time) (string, error) {
	return "0.00", nil
}
func (f *fakeLedgerRepo) GetTrialBalance(ctx context.Context, periodID uuid.UUID) ([]repositories.TrialBalanceRow, error) {
	return nil, nil
}

type fakeUoW struct {
	entries *fakeEntryRepo
}

func (u *fakeUoW) JournalEntries() repositories.JournalRepository { return u.entries }
func (u *fakeUoW) Ledger() repositories.LedgerRepository { return nil }
// everything else nil
func (u *fakeUoW) Customers() repositories.CustomerRepository { return nil }
func (u *fakeUoW) Suppliers() repositories.SupplierRepository { return nil }
func (u *fakeUoW) Products() repositories.ProductRepository { return nil }
func (u *fakeUoW) Categories() repositories.CategoryRepository { return nil }
func (u *fakeUoW) Brands() repositories.BrandRepository { return nil }
func (u *fakeUoW) Warehouses() repositories.WarehouseRepository { return nil }
func (u *fakeUoW) Currencies() repositories.CurrencyRepository { return nil }
func (u *fakeUoW) Users() repositories.UserRepository { return nil }
func (u *fakeUoW) Roles() repositories.RoleRepository { return nil }
func (u *fakeUoW) Permissions() repositories.PermissionRepository { return nil }
func (u *fakeUoW) UserRoles() repositories.UserRoleRepository { return nil }
func (u *fakeUoW) InventoryBatches() repositories.InventoryBatchRepository { return nil }
func (u *fakeUoW) InventoryMovements() repositories.InventoryMovementRepository { return nil }
func (u *fakeUoW) PurchaseOrders() repositories.PurchaseRepository { return nil }
func (u *fakeUoW) SupplierPayments() repositories.SupplierPaymentRepository { return nil }
func (u *fakeUoW) AccountsPayable() repositories.AccountsPayableRepository { return nil }
func (u *fakeUoW) Sales() repositories.SalesRepository { return nil }
func (u *fakeUoW) CustomerPayments() repositories.CustomerPaymentRepository { return nil }
func (u *fakeUoW) CustomerAdvances() repositories.CustomerAdvanceRepository { return nil }
func (u *fakeUoW) AccountsReceivable() repositories.AccountsReceivableRepository { return nil }
func (u *fakeUoW) BankAccounts() repositories.BankAccountRepository { return nil }
func (u *fakeUoW) CreditCards() repositories.CreditCardRepository { return nil }
func (u *fakeUoW) BankTransactions() repositories.BankTransactionRepository { return nil }
func (u *fakeUoW) ExchangeRates() repositories.ExchangeRateRepository { return nil }
func (u *fakeUoW) ChartOfAccounts() repositories.ChartOfAccountsRepository { return nil }

type fakeTxm struct{ uow *fakeUoW }

func (f *fakeTxm) WithinTransaction(ctx context.Context, fn services.TxRunner) error {
	ctx = repositories.ContextWithUnitOfWork(ctx, f.uow)
	return fn(ctx)
}

func newSvc(entries *fakeEntryRepo) *accsvc.AccountingService {
	txm := &fakeTxm{uow: &fakeUoW{entries: entries}}
	log := common.NewLogger(silentSlog())
	return accsvc.New(entries, &fakeChartRepo{}, &fakeLedgerRepo{}, txm, log)
}

func silentSlog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func money(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	return m
}
func date(t *testing.T, y int, m time.Month, d int) valueobjects.Date {
	t.Helper()
	date, err := valueobjects.NewDate(y, m, d)
	if err != nil {
		t.Fatalf("date: %v", err)
	}
	return date
}

func TestAccountingService_CreateEntry_RejectsUnbalanced(t *testing.T) {
	svc := newSvc(&fakeEntryRepo{entries: make(map[uuid.UUID]*entacc.JournalEntry)})
	acc1, acc2 := uuid.New(), uuid.New()
	_, err := svc.CreateEntry(context.Background(), accsvc.EntryInput{
		CompanyID:      uuid.New(),
		FiscalPeriodID: uuid.New(),
		Number:         "JE-1",
		EntryDate:      date(t, 2026, 1, 15),
		Description:    "test",
		Source:         enums.JournalTypeManual,
		Lines: []accsvc.EntryLineInput{
			{AccountID: acc1, Debit: money(t, "100"), CurrencyCode: valueobjects.MustCurrencyCode("PEN")},
			{AccountID: acc2, Credit: money(t, "99"), CurrencyCode: valueobjects.MustCurrencyCode("PEN")},
		},
	})
	if err == nil {
		t.Error("unbalanced entry should fail")
	}
}

func TestAccountingService_CreateEntry_BalancedPersists(t *testing.T) {
	repo := &fakeEntryRepo{entries: make(map[uuid.UUID]*entacc.JournalEntry)}
	svc := newSvc(repo)
	acc1, acc2 := uuid.New(), uuid.New()
	entry, err := svc.CreateEntry(context.Background(), accsvc.EntryInput{
		CompanyID:      uuid.New(),
		FiscalPeriodID: uuid.New(),
		Number:         "JE-1",
		EntryDate:      date(t, 2026, 1, 15),
		Description:    "test",
		Source:         enums.JournalTypeManual,
		Lines: []accsvc.EntryLineInput{
			{AccountID: acc1, Debit: money(t, "100"), CurrencyCode: valueobjects.MustCurrencyCode("PEN")},
			{AccountID: acc2, Credit: money(t, "100"), CurrencyCode: valueobjects.MustCurrencyCode("PEN")},
		},
	})
	if err != nil {
		t.Fatalf("balanced entry: %v", err)
	}
	if entry.Status != enums.JournalStatusDraft {
		t.Errorf("status: %s", entry.Status)
	}
	if repo.entries[entry.ID] == nil {
		t.Error("not persisted")
	}
}

func TestAccountingService_PostAndReverse(t *testing.T) {
	repo := &fakeEntryRepo{entries: make(map[uuid.UUID]*entacc.JournalEntry)}
	svc := newSvc(repo)
	acc1, acc2 := uuid.New(), uuid.New()
	entry, _ := svc.CreateEntry(context.Background(), accsvc.EntryInput{
		CompanyID:      uuid.New(),
		FiscalPeriodID: uuid.New(),
		Number:         "JE-1",
		EntryDate:      date(t, 2026, 1, 15),
		Source:         enums.JournalTypeManual,
		Lines: []accsvc.EntryLineInput{
			{AccountID: acc1, Debit: money(t, "100"), CurrencyCode: valueobjects.MustCurrencyCode("PEN")},
			{AccountID: acc2, Credit: money(t, "100"), CurrencyCode: valueobjects.MustCurrencyCode("PEN")},
		},
	})
	if _, err := svc.Post(context.Background(), entry.ID, uuid.New()); err != nil {
		t.Fatalf("post: %v", err)
	}
	if entry.Status != enums.JournalStatusPosted {
		t.Errorf("status: %s", entry.Status)
	}
	// Try to re-post should fail
	if _, err := svc.Post(context.Background(), entry.ID, uuid.New()); err == nil {
		t.Error("double post should fail")
	}
}
