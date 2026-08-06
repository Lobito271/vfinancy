package sales_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	custsvc "vfinancy/backend/internal/application/services/customer"
	salessvc "vfinancy/backend/internal/application/services/sales"
	"vfinancy/backend/internal/domain/entities/masterdata"
	entsales "vfinancy/backend/internal/domain/entities/sales"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

type fakeSalesRepo struct {
	sales    map[uuid.UUID]*entsales.Sale
	byNumber map[string]uuid.UUID
}

func newFakeSalesRepo() *fakeSalesRepo {
	return &fakeSalesRepo{sales: make(map[uuid.UUID]*entsales.Sale), byNumber: make(map[string]uuid.UUID)}
}

func (f *fakeSalesRepo) Create(ctx context.Context, s *entsales.Sale) error {
	if _, exists := f.sales[s.ID]; exists {
		return repositories.ErrDuplicate
	}
	f.sales[s.ID] = s
	f.byNumber[s.Number] = s.ID
	return nil
}
func (f *fakeSalesRepo) Update(ctx context.Context, s *entsales.Sale) error {
	f.sales[s.ID] = s
	return nil
}
func (f *fakeSalesRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (f *fakeSalesRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := f.sales[id]
	return ok, nil
}
func (f *fakeSalesRepo) GetByID(ctx context.Context, id uuid.UUID) (*entsales.Sale, error) {
	s, ok := f.sales[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return s, nil
}
func (f *fakeSalesRepo) GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*entsales.Sale, error) {
	id, ok := f.byNumber[number]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return f.sales[id], nil
}
func (f *fakeSalesRepo) List(ctx context.Context, f2 repositories.SaleFilter) (repositories.Page[*entsales.Sale], error) {
	out := make([]*entsales.Sale, 0, len(f.sales))
	for _, s := range f.sales {
		if f2.CompanyID != nil && s.CompanyID != *f2.CompanyID {
			continue
		}
		out = append(out, s)
	}
	return repositories.Page[*entsales.Sale]{Items: out, Total: len(out), Limit: 1000}, nil
}
func (f *fakeSalesRepo) GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	return "F001-1", nil
}

type fakeCustomerRepo struct {
	customers map[uuid.UUID]*masterdata.Customer
}

func (f *fakeCustomerRepo) Create(ctx context.Context, c *masterdata.Customer) error {
	f.customers[c.ID] = c
	return nil
}
func (f *fakeCustomerRepo) Update(ctx context.Context, c *masterdata.Customer) error {
	f.customers[c.ID] = c
	return nil
}
func (f *fakeCustomerRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (f *fakeCustomerRepo) GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Customer, error) {
	c, ok := f.customers[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return c, nil
}
func (f *fakeCustomerRepo) GetByDocument(ctx context.Context, companyID uuid.UUID, documentType, documentNumber string) (*masterdata.Customer, error) {
	return nil, repositories.ErrNotFound
}
func (f *fakeCustomerRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := f.customers[id]
	return ok, nil
}
func (f *fakeCustomerRepo) List(ctx context.Context, f2 repositories.CustomerFilter) (repositories.Page[*masterdata.Customer], error) {
	return repositories.Page[*masterdata.Customer]{}, nil
}
func (f *fakeCustomerRepo) GetOutstandingBalance(ctx context.Context, id uuid.UUID) (string, error) {
	return "0.00", nil
}

type fakeUoW struct {
	sales    *fakeSalesRepo
	customers *fakeCustomerRepo
}

func (u *fakeUoW) Sales() repositories.SalesRepository { return u.sales }
func (u *fakeUoW) Customers() repositories.CustomerRepository { return u.customers }
// the rest are nil — service under test only uses the two above
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

type fakeTxm struct {
	commits, rollbacks int
	uow                *fakeUoW
}

func (f *fakeTxm) WithinTransaction(ctx context.Context, fn services.TxRunner) error {
	ctx = repositories.ContextWithUnitOfWork(ctx, f.uow)
	if err := fn(ctx); err != nil {
		f.rollbacks++
		return err
	}
	f.commits++
	return nil
}

func newSvc(salesRepo *fakeSalesRepo, custRepo *fakeCustomerRepo) (*salessvc.SalesService, *fakeTxm) {
	uow := &fakeUoW{sales: salesRepo, customers: custRepo}
	txm := &fakeTxm{uow: uow}
	return salessvc.New(salesRepo, txm, common.NewLogger(slog.New(slog.NewTextHandler(&devNull{}, nil)))), txm
}

type devNull struct{}

func (devNull) Write(p []byte) (int, error) { return len(p), nil }

func money(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	return m
}
func quantity(t *testing.T, s string) valueobjects.Quantity {
	t.Helper()
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		t.Fatalf("quantity: %v", err)
	}
	return q
}
func buildCustomer(t *testing.T, companyID uuid.UUID) *masterdata.Customer {
	t.Helper()
	doc, _ := valueobjects.NewDocumentNumber(enums.DocumentTypeRUC, "20123456789")
	addr, _ := valueobjects.NewAddress("Calle 1")
	email, _ := valueobjects.NewEmail("buyer@example.com")
	c, err := masterdata.NewCustomer(time.Now().UTC(), masterdata.NewCustomerOptions{
		CompanyID:    companyID,
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

func TestSalesService_Create_EmptyDocument(t *testing.T) {
	repo := newFakeSalesRepo()
	custRepo := &fakeCustomerRepo{customers: make(map[uuid.UUID]*masterdata.Customer)}
	svc, _ := newSvc(repo, custRepo)
	_, err := svc.Create(context.Background(), salessvc.CreateInput{
		CompanyID:   uuid.New(),
		Number:      "F001-1",
		CustomerID:  uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: valueobjects.One(),
		Items:       nil,
	})
	if err == nil {
		t.Fatal("expected empty-document error")
	}
}

func TestSalesService_Create_DuplicateItem(t *testing.T) {
	repo := newFakeSalesRepo()
	custRepo := &fakeCustomerRepo{customers: make(map[uuid.UUID]*masterdata.Customer)}
	svc, _ := newSvc(repo, custRepo)
	companyID := uuid.New()
	customer := buildCustomer(t, companyID)
	custRepo.customers[customer.ID] = customer
	product1 := uuid.New()
	product2 := uuid.New() // ignored: duplicate of product1's
	_ = product2
	// Add two lines for the same product: should fail.
	_, err := svc.Create(context.Background(), salessvc.CreateInput{
		CompanyID:   companyID,
		Number:      "F001-1",
		CustomerID:  customer.ID,
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: valueobjects.One(),
		Items: []salessvc.CreateItemInput{
			{ProductID: product1, Quantity: quantity(t, "1"), UnitPrice: money(t, "100"), CostSnapshot: money(t, "60")},
			{ProductID: product1, Quantity: quantity(t, "1"), UnitPrice: money(t, "100"), CostSnapshot: money(t, "60")},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate-line error")
	}
}

func TestSalesService_Cancel(t *testing.T) {
	repo := newFakeSalesRepo()
	custRepo := &fakeCustomerRepo{customers: make(map[uuid.UUID]*masterdata.Customer)}
	svc, _ := newSvc(repo, custRepo)
	companyID := uuid.New()
	customer := buildCustomer(t, companyID)
	custRepo.customers[customer.ID] = customer
	productID := uuid.New()
	res, err := svc.Create(context.Background(), salessvc.CreateInput{
		CompanyID:   companyID,
		Number:      "F001-1",
		CustomerID:  customer.ID,
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: valueobjects.One(),
		Items: []salessvc.CreateItemInput{
			{ProductID: productID, Quantity: quantity(t, "1"), UnitPrice: money(t, "100"), CostSnapshot: money(t, "60")},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Cancel(context.Background(), salessvc.CancelInput{ID: res.Sale.ID, Reason: "test"}); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if !res.Sale.IsCancelled() {
		t.Error("sale should be cancelled")
	}
}

func TestSalesService_ApplyPayment_AutoStatus(t *testing.T) {
	repo := newFakeSalesRepo()
	custRepo := &fakeCustomerRepo{customers: make(map[uuid.UUID]*masterdata.Customer)}
	svc, _ := newSvc(repo, custRepo)
	companyID := uuid.New()
	customer := buildCustomer(t, companyID)
	custRepo.customers[customer.ID] = customer
	res, err := svc.Create(context.Background(), salessvc.CreateInput{
		CompanyID:   companyID,
		Number:      "F001-1",
		CustomerID:  customer.ID,
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: valueobjects.One(),
		Items: []salessvc.CreateItemInput{
			{ProductID: uuid.New(), Quantity: quantity(t, "1"), UnitPrice: money(t, "100"), CostSnapshot: money(t, "60")},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	balance, err := svc.ApplyPayment(context.Background(), res.Sale.ID, money(t, "100"))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if balance.String() != "0.00" {
		t.Errorf("balance: %s", balance)
	}
	if !res.Sale.IsPaid() {
		t.Error("sale should be paid")
	}
}

func TestSalesService_OutstandingBalance(t *testing.T) {
	repo := newFakeSalesRepo()
	custRepo := &fakeCustomerRepo{customers: make(map[uuid.UUID]*masterdata.Customer)}
	svc, _ := newSvc(repo, custRepo)
	companyID := uuid.New()
	customer := buildCustomer(t, companyID)
	custRepo.customers[customer.ID] = customer
	res, err := svc.Create(context.Background(), salessvc.CreateInput{
		CompanyID:   companyID,
		Number:      "F001-1",
		CustomerID:  customer.ID,
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: valueobjects.One(),
		Items: []salessvc.CreateItemInput{
			{ProductID: uuid.New(), Quantity: quantity(t, "1"), UnitPrice: money(t, "100"), CostSnapshot: money(t, "60")},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	bal, err := svc.OutstandingBalance(context.Background(), res.Sale.ID)
	if err != nil {
		t.Fatalf("outstanding: %v", err)
	}
	if bal.String() != "100.00" {
		t.Errorf("outstanding: %s", bal)
	}
}

// silence unused imports
var _ = custsvc.CustomerService{}
