package inventory_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	invsvc "vfinancy/backend/internal/application/services/inventory"
	"vfinancy/backend/internal/domain/entities/inventory"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// fakeBatchRepo is the in-memory implementation of InventoryBatchRepository
// sufficient for the tests.
type fakeBatchRepo struct {
	batches map[uuid.UUID]*inventory.InventoryBatch
}

func newFakeBatchRepo() *fakeBatchRepo { return &fakeBatchRepo{batches: make(map[uuid.UUID]*inventory.InventoryBatch)} }

func (f *fakeBatchRepo) Create(ctx context.Context, b *inventory.InventoryBatch) error {
	if _, exists := f.batches[b.ID]; exists {
		return repositories.ErrDuplicate
	}
	f.batches[b.ID] = b
	return nil
}

func (f *fakeBatchRepo) Update(ctx context.Context, b *inventory.InventoryBatch) error {
	f.batches[b.ID] = b
	return nil
}

func (f *fakeBatchRepo) GetByID(ctx context.Context, id uuid.UUID) (*inventory.InventoryBatch, error) {
	b, ok := f.batches[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return b, nil
}

func (f *fakeBatchRepo) List(ctx context.Context, f2 repositories.InventoryBatchFilter) (repositories.Page[*inventory.InventoryBatch], error) {
	out := make([]*inventory.InventoryBatch, 0, len(f.batches))
	for _, b := range f.batches {
		if f2.CompanyID != nil && b.CompanyID != *f2.CompanyID {
			continue
		}
		if f2.OnlyActive && b.Status != "active" {
			continue
		}
		out = append(out, b)
	}
	return repositories.Page[*inventory.InventoryBatch]{Items: out, Total: len(out), Limit: 1000}, nil
}

func (f *fakeBatchRepo) GetStockSummary(ctx context.Context, productID, warehouseID uuid.UUID) (float64, string, error) {
	for _, b := range f.batches {
		if b.ProductID == productID && b.WarehouseID == warehouseID {
			return b.CurrentQuantity.Decimal().InexactFloat64(), b.UnitCost.String(), nil
		}
	}
	return 0, "0", repositories.ErrNotFound
}

func (f *fakeBatchRepo) ListClearance(ctx context.Context, companyID uuid.UUID, at time.Time) ([]*inventory.InventoryBatch, error) {
	out := []*inventory.InventoryBatch{}
	for _, b := range f.batches {
		if matchesCompany(b, companyID) && b.IsClearance(at) {
			out = append(out, b)
		}
	}
	return out, nil
}

func matchesCompany(b *inventory.InventoryBatch, companyID uuid.UUID) bool {
	return b.CompanyID == companyID
}

// fakeMovementRepo is the in-memory implementation of InventoryMovementRepository.
type fakeMovementRepo struct {
	movements []inventory.InventoryMovement
}

func newFakeMovementRepo() *fakeMovementRepo { return &fakeMovementRepo{} }

func (f *fakeMovementRepo) Create(ctx context.Context, m *inventory.InventoryMovement) error {
	f.movements = append(f.movements, *m)
	return nil
}

func (f *fakeMovementRepo) GetByID(ctx context.Context, id uuid.UUID) (*inventory.InventoryMovement, error) {
	return nil, repositories.ErrNotFound
}

func (f *fakeMovementRepo) List(ctx context.Context, f2 repositories.InventoryMovementFilter) (repositories.Page[*inventory.InventoryMovement], error) {
	return repositories.Page[*inventory.InventoryMovement]{}, nil
}

func (f *fakeMovementRepo) StockAt(ctx context.Context, productID, warehouseID uuid.UUID, at time.Time) (float64, error) {
	return 0, nil
}

// fakeUoW is the unit-of-work that exposes the in-memory batch +
// movement repos. Only the methods the service under test calls are
// implemented.
type fakeUoW struct {
	batches    *fakeBatchRepo
	movements  *fakeMovementRepo
}

func (u *fakeUoW) InventoryBatches() repositories.InventoryBatchRepository { return u.batches }
func (u *fakeUoW) InventoryMovements() repositories.InventoryMovementRepository { return u.movements }
func (u *fakeUoW) Customers() repositories.CustomerRepository                  { return nil }
func (u *fakeUoW) Suppliers() repositories.SupplierRepository                  { return nil }
func (u *fakeUoW) Products() repositories.ProductRepository                   { return nil }
func (u *fakeUoW) Categories() repositories.CategoryRepository                { return nil }
func (u *fakeUoW) Brands() repositories.BrandRepository                      { return nil }
func (u *fakeUoW) Warehouses() repositories.WarehouseRepository              { return nil }
func (u *fakeUoW) Currencies() repositories.CurrencyRepository                { return nil }
func (u *fakeUoW) Users() repositories.UserRepository                       { return nil }
func (u *fakeUoW) Roles() repositories.RoleRepository                       { return nil }
func (u *fakeUoW) Permissions() repositories.PermissionRepository           { return nil }
func (u *fakeUoW) UserRoles() repositories.UserRoleRepository               { return nil }
func (u *fakeUoW) PurchaseOrders() repositories.PurchaseRepository           { return nil }
func (u *fakeUoW) SupplierPayments() repositories.SupplierPaymentRepository   { return nil }
func (u *fakeUoW) AccountsPayable() repositories.AccountsPayableRepository    { return nil }
func (u *fakeUoW) Sales() repositories.SalesRepository                       { return nil }
func (u *fakeUoW) CustomerPayments() repositories.CustomerPaymentRepository  { return nil }
func (u *fakeUoW) CustomerAdvances() repositories.CustomerAdvanceRepository  { return nil }
func (u *fakeUoW) AccountsReceivable() repositories.AccountsReceivableRepository { return nil }
func (u *fakeUoW) BankAccounts() repositories.BankAccountRepository           { return nil }
func (u *fakeUoW) CreditCards() repositories.CreditCardRepository             { return nil }
func (u *fakeUoW) BankTransactions() repositories.BankTransactionRepository    { return nil }
func (u *fakeUoW) ExchangeRates() repositories.ExchangeRateRepository        { return nil }
func (u *fakeUoW) ChartOfAccounts() repositories.ChartOfAccountsRepository   { return nil }
func (u *fakeUoW) JournalEntries() repositories.JournalRepository             { return nil }
func (u *fakeUoW) Ledger() repositories.LedgerRepository                     { return nil }

// fakeTxm runs the callback and exposes the UoW via context.
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

func date(t *testing.T, y int, m time.Month, d int) valueobjects.Date {
	t.Helper()
	date, err := valueobjects.NewDate(y, m, d)
	if err != nil {
		t.Fatalf("date: %v", err)
	}
	return date
}

func newSvc(b *fakeBatchRepo, m *fakeMovementRepo) (*invsvc.InventoryService, *fakeTxm) {
	uow := &fakeUoW{batches: b, movements: m}
	txm := &fakeTxm{uow: uow}
	log := common.NewLogger(slog.New(slog.NewTextHandler(&devNullSink{}, nil)))
	return invsvc.New(b, m, txm, log), txm
}

type devNullSink struct{}

func (devNullSink) Write(p []byte) (int, error) { return len(p), nil }

func buildBatch(t *testing.T, arrival valueobjects.Date, qty string) *inventory.InventoryBatch {
	t.Helper()
	addr, _ := valueobjects.NewAddress("Main 1")
	b, err := inventory.NewInventoryBatch(time.Now().UTC(), inventory.NewInventoryBatchOptions{
		CompanyID:      uuid.New(),
		ProductID:      uuid.New(),
		WarehouseID:    uuid.New(),
		ArrivalDate:    arrival,
		InitialQuantity: quantity(t, qty),
		UnitCost:        money(t, "10.00"),
		CurrencyCode:    valueobjects.MustCurrencyCode("PEN"),
	})
	if err != nil {
		t.Fatalf("buildBatch: %v", err)
	}
	_ = addr
	return b
}

func TestInventoryService_Receive_Success(t *testing.T) {
	bRepo := newFakeBatchRepo()
	mRepo := newFakeMovementRepo()
	svc, txm := newSvc(bRepo, mRepo)
	ctx := context.Background()

	b, err := svc.Receive(ctx, invsvc.ReceiveInput{
		CompanyID:   uuid.New(),
		ProductID:   uuid.New(),
		WarehouseID: uuid.New(),
		ArrivalDate:  date(t, 2026, 1, 15),
		Quantity:    quantity(t, "100"),
		UnitCost:    money(t, "10.00"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
	})
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	if bRepo.batches[b.ID] == nil {
		t.Fatal("batch not persisted")
	}
	if len(mRepo.movements) != 1 {
		t.Errorf("movements: got %d want 1", len(mRepo.movements))
	}
	if txm.commits != 1 {
		t.Errorf("commits: %d", txm.commits)
	}
}

func TestInventoryService_25DayClearanceRule(t *testing.T) {
	// The 25-day rule: arrival + 25 days is the boundary.
	arrival := date(t, 2026, 1, 1)
	b := buildBatch(t, arrival, "100")

	// Day 0: not clearance.
	day0 := date(t, 2026, 1, 1)
	if b.IsClearance(day0) {
		t.Error("day 0 should not be clearance")
	}
	// Day 25 = boundary = clearance.
	day25 := date(t, 2026, 1, 26)
	if !b.IsClearance(day25) {
		t.Error("day 25 should be clearance (boundary)")
	}
	// Day 30 = clearance.
	day30 := date(t, 2026, 1, 31)
	if !b.IsClearance(day30) {
		t.Error("day 30 should be clearance")
	}
	// Maximum sale date is arrival + 25.
	if got := b.MaximumSaleDate(); !got.Equal(date(t, 2026, 1, 26)) {
		t.Errorf("max sale date: got %s want 2026-01-26", got)
	}
}

func TestInventoryService_NegativeStockPrevention(t *testing.T) {
	b := buildBatch(t, date(t, 2026, 1, 1), "10")
	if err := b.Consume(quantity(t, "5")); err != nil {
		t.Fatalf("consume: %v", err)
	}
	// Try to consume more than available.
	if err := b.Consume(quantity(t, "10")); err == nil {
		t.Error("consume beyond stock should fail")
	}
}

func TestInventoryService_TransferSameProductOnly(t *testing.T) {
	bRepo := newFakeBatchRepo()
	mRepo := newFakeMovementRepo()
	svc, _ := newSvc(bRepo, mRepo)
	ctx := context.Background()

	from, _ := svc.Receive(ctx, invsvc.ReceiveInput{
		CompanyID:   uuid.New(),
		ProductID:   uuid.New(),
		WarehouseID: uuid.New(),
		ArrivalDate:  date(t, 2026, 1, 1),
		Quantity:    quantity(t, "50"),
		UnitCost:    money(t, "10"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
	})
	to, _ := svc.Receive(ctx, invsvc.ReceiveInput{
		CompanyID:   from.CompanyID,
		ProductID:   from.ProductID,
		WarehouseID: uuid.New(),
		ArrivalDate:  date(t, 2026, 1, 1),
		Quantity:    quantity(t, "1"),
		UnitCost:    money(t, "10"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
	})
	err := svc.Transfer(ctx, invsvc.TransferInput{
		FromBatchID: from.ID,
		ToBatchID:   to.ID,
		Quantity:    quantity(t, "1"),
	})
	if err != nil {
		t.Errorf("same product transfer should succeed: %v", err)
	}
	// Mismatched product: from=ProductA, to=ProductB → should fail.
	from2, _ := svc.Receive(ctx, invsvc.ReceiveInput{
		CompanyID:   from.CompanyID,
		ProductID:   uuid.New(),
		WarehouseID: uuid.New(),
		ArrivalDate:  date(t, 2026, 1, 1),
		Quantity:    quantity(t, "5"),
		UnitCost:    money(t, "10"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
	})
	err = svc.Transfer(ctx, invsvc.TransferInput{
		FromBatchID: from2.ID,
		ToBatchID:   to.ID,
		Quantity:    quantity(t, "1"),
	})
	if err == nil {
		t.Error("mismatched-product transfer should fail")
	}
}

func TestInventoryService_ClearanceCandidates(t *testing.T) {
	bRepo := newFakeBatchRepo()
	mRepo := newFakeMovementRepo()
	svc, _ := newSvc(bRepo, mRepo)
	ctx := context.Background()

	// batch1: arrival 2025-12-01 — clearance
	// batch2: arrival 2026-01-20 — not clearance
	now := time.Now().UTC()
	oldArrival := valueobjects.Date{}
	_ = oldArrival
	_ = now

	// Use Receive to create both.
	companyID := uuid.New()
	for i, arrival := range []valueobjects.Date{
		date(t, 2025, 12, 1),
		date(t, 2026, 1, 20),
	} {
		_, err := svc.Receive(ctx, invsvc.ReceiveInput{
			CompanyID:   companyID,
			ProductID:   uuid.New(),
			WarehouseID: uuid.New(),
			ArrivalDate:  arrival,
			Quantity:    quantity(t, "10"),
			UnitCost:    money(t, "5"),
			CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		})
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	today := date(t, 2026, 1, 30)
	candidates, err := svc.GenerateClearanceCandidates(ctx, companyID, today)
	if err != nil {
		t.Fatalf("candidates: %v", err)
	}
	// The first batch (arrival 2025-12-01) is clearance by 2026-01-30
	// (60 days old, > 25). The second (2026-01-20) is 10 days old, not
	// clearance. So we expect 1.
	if len(candidates) != 1 {
		t.Errorf("clearance candidates: got %d want 1", len(candidates))
	}
}
