package customer_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/application/services/customer"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/entities/masterdata"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// fakeCustomerRepo is the minimum in-memory implementation of the
// CustomerRepository needed by the service tests. It does not aim to
// mirror the real repo — only the methods the service actually calls.
type fakeCustomerRepo struct {
	customers map[uuid.UUID]*masterdata.Customer
	byDoc     map[string]uuid.UUID
}

func newFakeCustomerRepo() *fakeCustomerRepo {
	return &fakeCustomerRepo{
		customers: make(map[uuid.UUID]*masterdata.Customer),
		byDoc:     make(map[string]uuid.UUID),
	}
}

func (f *fakeCustomerRepo) key(docType enums.DocumentType, docNum string) string {
	return string(docType) + ":" + docNum
}

func (f *fakeCustomerRepo) Create(ctx context.Context, c *masterdata.Customer) error {
	if _, exists := f.customers[c.ID]; exists {
		return fmt.Errorf("fake: customer %s already exists", c.ID)
	}
	f.customers[c.ID] = c
	f.byDoc[f.key(c.Document.Type(), c.Document.Number())] = c.ID
	return nil
}

func (f *fakeCustomerRepo) Update(ctx context.Context, c *masterdata.Customer) error {
	f.customers[c.ID] = c
	f.byDoc[f.key(c.Document.Type(), c.Document.Number())] = c.ID
	return nil
}

func (f *fakeCustomerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if c, ok := f.customers[id]; ok {
		delete(f.byDoc, f.key(c.Document.Type(), c.Document.Number()))
		delete(f.customers, id)
	}
	return nil
}

func (f *fakeCustomerRepo) GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Customer, error) {
	c, ok := f.customers[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return c, nil
}

func (f *fakeCustomerRepo) GetByDocument(ctx context.Context, companyID uuid.UUID, documentType, documentNumber string) (*masterdata.Customer, error) {
	id, ok := f.byDoc[documentType+":"+documentNumber]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return f.customers[id], nil
}

func (f *fakeCustomerRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := f.customers[id]
	return ok, nil
}

func (f *fakeCustomerRepo) List(ctx context.Context, filter repositories.CustomerFilter) (repositories.Page[*masterdata.Customer], error) {
	var out []*masterdata.Customer
	for _, c := range f.customers {
		if filter.CompanyID != nil && c.CompanyID != *filter.CompanyID {
			continue
		}
		if filter.Status != "" && string(c.Status) != filter.Status {
			continue
		}
		out = append(out, c)
	}
	return repositories.Page[*masterdata.Customer]{Items: out, Total: len(out), Limit: 1000, Offset: 0}, nil
}

func (f *fakeCustomerRepo) GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error) {
	c, ok := f.customers[id]
	if !ok {
		return "", repositories.ErrNotFound
	}
	return c.CurrentDebt.String(), nil
}

// silentLogger discards every log line. The service uses the logger
// only for Info; we drop the output.
var silentLogger = common.NewLogger(slog.New(slog.NewTextHandler(&devNull{}, nil)))

type devNull struct{}

func (devNull) Write(p []byte) (int, error) { return len(p), nil }

// fakeTxManager runs the callback without touching a real database.
// Records commits and rollbacks for assertion.
type fakeTxManager struct {
	commits   int
	rollbacks int
	repo      *fakeCustomerRepo
}

func (f *fakeTxManager) WithinTransaction(ctx context.Context, fn services.TxRunner) error {
	ctx = repositories.ContextWithUnitOfWork(ctx, &fakeUoW{repo: f.repo})
	if err := fn(ctx); err != nil {
		f.rollbacks++
		return err
	}
	f.commits++
	return nil
}

// fakeUoW is the UnitOfWork handed to the callback. Only the Customers
// method is implemented (the service under test only touches it).
type fakeUoW struct {
	repo *fakeCustomerRepo
}

func (u *fakeUoW) Customers() repositories.CustomerRepository { return u.repo }
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
func (u *fakeUoW) JournalEntries() repositories.JournalRepository { return nil }
func (u *fakeUoW) Ledger() repositories.LedgerRepository { return nil }

// buildCustomer creates a valid active customer.
func buildCustomer(t *testing.T) *masterdata.Customer {
	t.Helper()
	doc, _ := valueobjects.NewDocumentNumber(enums.DocumentTypeRUC, "20123456789")
	addr, _ := valueobjects.NewAddress("Calle 1")
	email, _ := valueobjects.NewEmail("buyer@example.com")
	c, err := masterdata.NewCustomer(time.Now().UTC(), masterdata.NewCustomerOptions{
		CompanyID:    uuid.New(),
		Document:     doc,
		BusinessName: valueobjects.MustFullName("Test Co"),
		TaxCategory:  enums.TaxCategoryTaxed,
		CreditLimit:  money(t, "1000.00"),
		Email:        email,
		Address:      addr,
	})
	if err != nil {
		t.Fatalf("buildCustomer: %v", err)
	}
	return c
}

func money(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	return m
}

// newSvc wires the service with the standard fakes.
func newSvc(repo *fakeCustomerRepo, txm *fakeTxManager) *customer.CustomerService {
	txm.repo = repo
	return customer.New(repo, txm, silentLogger)
}

func TestCustomerService_Create_Success(t *testing.T) {
	repo := newFakeCustomerRepo()
	txm := &fakeTxManager{}
	svc := newSvc(repo, txm)

	in := customer.CreateInput{
		CompanyID:    uuid.New(),
		DocumentType: enums.DocumentTypeRUC,
		DocumentNumber:  "20999999999",
		BusinessName: "Acme SAC",
		TaxCategory:  enums.TaxCategoryTaxed,
		CreditLimit:  money(t, "5000.00"),
		Email:        "billing@acme.com",
		Address:      "Av Demo 123",
	}

	c, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.ID == uuid.Nil {
		t.Fatal("expected non-zero ID")
	}
	if txm.commits != 1 {
		t.Errorf("commits: got %d want 1", txm.commits)
	}
	if txm.rollbacks != 0 {
		t.Errorf("rollbacks: got %d want 0", txm.rollbacks)
	}
	if got := len(repo.customers); got != 1 {
		t.Errorf("repo: got %d customers want 1", got)
	}
}

func TestCustomerService_Create_InvalidDocument(t *testing.T) {
	svc := newSvc(newFakeCustomerRepo(), &fakeTxManager{})
	_, err := svc.Create(context.Background(), customer.CreateInput{
		CompanyID:    uuid.New(),
		DocumentType: enums.DocumentTypeRUC,
		DocumentNumber:  "123", // invalid RUC
		BusinessName: "X",
		TaxCategory:  enums.TaxCategoryTaxed,
		CreditLimit:  money(t, "0"),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCustomerService_Block_Unblock(t *testing.T) {
	repo := newFakeCustomerRepo()
	c := buildCustomer(t)
	repo.customers[c.ID] = c
	repo.byDoc[repo.key(c.Document.Type(), c.Document.Number())] = c.ID

	svc := newSvc(repo, &fakeTxManager{})
	if err := svc.Block(context.Background(), c.ID, "credit overdue"); err != nil {
		t.Fatalf("block: %v", err)
	}
	if repo.customers[c.ID].Status != enums.CustomerStatusBlocked {
		t.Errorf("status: %s", repo.customers[c.ID].Status)
	}
	if repo.customers[c.ID].BlockedReason != "credit overdue" {
		t.Errorf("reason: %q", repo.customers[c.ID].BlockedReason)
	}

	if err := svc.Unblock(context.Background(), c.ID); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	if repo.customers[c.ID].Status != enums.CustomerStatusActive {
		t.Errorf("status after unblock: %s", repo.customers[c.ID].Status)
	}
}

func TestCustomerService_RecordSalePayment(t *testing.T) {
	repo := newFakeCustomerRepo()
	c := buildCustomer(t)
	repo.customers[c.ID] = c
	repo.byDoc[repo.key(c.Document.Type(), c.Document.Number())] = c.ID

	svc := newSvc(repo, &fakeTxManager{})

	// Record a sale of 300.
	debt, err := svc.RecordSale(context.Background(), c.ID, money(t, "300.00"))
	if err != nil {
		t.Fatalf("record sale: %v", err)
	}
	if debt.String() != "300.00" {
		t.Errorf("debt: got %s want 300.00", debt)
	}

	// Pay 100.
	debt, err = svc.RecordPayment(context.Background(), c.ID, money(t, "100.00"))
	if err != nil {
		t.Fatalf("record payment: %v", err)
	}
	if debt.String() != "200.00" {
		t.Errorf("debt after payment: got %s want 200.00", debt)
	}
}

func TestCustomerService_RecordPayment_Invalid(t *testing.T) {
	repo := newFakeCustomerRepo()
	c := buildCustomer(t)
	repo.customers[c.ID] = c
	svc := newSvc(repo, &fakeTxManager{})

	if _, err := svc.RecordPayment(context.Background(), c.ID, money(t, "0")); err == nil {
		t.Error("zero payment should fail")
	}
	if _, err := svc.RecordPayment(context.Background(), c.ID, money(t, "-1")); err == nil {
		t.Error("negative payment should fail")
	}
}

func TestCustomerService_AvailableCreditAndIsOverLimit(t *testing.T) {
	repo := newFakeCustomerRepo()
	c := buildCustomer(t)
	repo.customers[c.ID] = c
	svc := newSvc(repo, &fakeTxManager{})

	available, err := svc.AvailableCredit(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("available: %v", err)
	}
	if available.String() != "1000.00" {
		t.Errorf("initial available: %s", available)
	}

	_, _ = svc.RecordSale(context.Background(), c.ID, money(t, "1200.00"))
	over, err := svc.IsOverLimit(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("is over: %v", err)
	}
	if !over {
		t.Error("should be over limit")
	}
}

func TestCustomerService_CanPlaceSale_Guards(t *testing.T) {
	repo := newFakeCustomerRepo()
	c := buildCustomer(t)
	repo.customers[c.ID] = c
	svc := newSvc(repo, &fakeTxManager{})

	// 1500 > 1000 limit, so this should fail.
	if err := svc.CanPlaceSale(context.Background(), c.ID, money(t, "1500.00")); err == nil {
		t.Error("over-limit sale should fail")
	}
	// Block the customer. Now any sale should fail.
	if err := svc.Block(context.Background(), c.ID, "x"); err != nil {
		t.Fatalf("block: %v", err)
	}
	err := svc.CanPlaceSale(context.Background(), c.ID, money(t, "100.00"))
	if err == nil {
		t.Error("blocked customer should not be able to place sale")
	}
	// The service returns services.ErrCustomerBlocked, which wraps
	// the domain's CUSTOMER_INACTIVE error.
	if err == nil {
		t.Fatal("expected error")
	}
	if !derrors.IsCode(err, "CUSTOMER_INACTIVE") {
		t.Errorf("expected CUSTOMER_INACTIVE code, got %v", err)
	}
}

func TestCustomerService_UpdateCreditLimit(t *testing.T) {
	repo := newFakeCustomerRepo()
	c := buildCustomer(t)
	repo.customers[c.ID] = c
	svc := newSvc(repo, &fakeTxManager{})

	if err := svc.UpdateCreditLimit(context.Background(), c.ID, money(t, "5000.00")); err != nil {
		t.Fatalf("update: %v", err)
	}
	if repo.customers[c.ID].CreditLimit.String() != "5000.00" {
		t.Errorf("limit: %s", repo.customers[c.ID].CreditLimit)
	}
	if err := svc.UpdateCreditLimit(context.Background(), c.ID, money(t, "-1")); err == nil {
		t.Error("negative limit should fail")
	}
}

func TestCustomerService_Deactivate_ActiveToInactive(t *testing.T) {
	repo := newFakeCustomerRepo()
	c := buildCustomer(t)
	repo.customers[c.ID] = c
	svc := newSvc(repo, &fakeTxManager{})

	if err := svc.Deactivate(context.Background(), c.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if repo.customers[c.ID].Status != enums.CustomerStatusInactive {
		t.Errorf("status: %s", repo.customers[c.ID].Status)
	}
}
