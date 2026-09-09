package sales_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/features/purchasing"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/customer"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/internal/features/product"
	"vfinancy/backend/internal/features/sales"
	"vfinancy/backend/internal/shared/apperrors"
)

type fakeTx struct{}

func (fakeTx) WithinTransaction(ctx context.Context, fn repositories.TxRunner) error { return fn(ctx) }

func money(s string) valueobjects.Money {
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		panic(err)
	}
	return m
}

type fakeSalesRepo struct {
	sales.SaleRepository
	seq     int
	created []*sales.Sale
	updated []*sales.Sale
	byID    map[uuid.UUID]*sales.Sale
}

func (f *fakeSalesRepo) Create(_ context.Context, s *sales.Sale, items []*sales.SaleItem) error {
	f.created = append(f.created, s)
	f.byID[s.ID] = s
	return nil
}

func (f *fakeSalesRepo) Update(_ context.Context, s *sales.Sale) error {
	f.updated = append(f.updated, s)
	f.byID[s.ID] = s
	return nil
}

func (f *fakeSalesRepo) GetByID(_ context.Context, id uuid.UUID) (*sales.Sale, error) {
	s, ok := f.byID[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return s, nil
}

func (f *fakeSalesRepo) NextNumber(context.Context) (string, error) {
	f.seq++
	return fmt.Sprintf("V-2026-%05d", f.seq), nil
}

type fakePaymentsRepo struct {
	sales.CustomerPaymentRepository
	seq      int
	created  []*sales.CustomerPayment
	allocs   [][]sales.PaymentAllocation
	statuses map[uuid.UUID]string
}

func (f *fakePaymentsRepo) Create(_ context.Context, p *sales.CustomerPayment, allocations []sales.PaymentAllocation) error {
	f.created = append(f.created, p)
	f.allocs = append(f.allocs, allocations)
	return nil
}

func (f *fakePaymentsRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	f.statuses[id] = status
	return nil
}

func (f *fakePaymentsRepo) NextNumber(context.Context) (string, error) {
	f.seq++
	return fmt.Sprintf("CP-2026-%05d", f.seq), nil
}

type fakeCustomerRepo struct {
	customer.CustomerRepository
	byID map[uuid.UUID]*customer.Customer
}

func (f *fakeCustomerRepo) GetByID(_ context.Context, id uuid.UUID) (*customer.Customer, error) {
	c, ok := f.byID[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return c, nil
}

func (f *fakeCustomerRepo) Update(_ context.Context, c *customer.Customer) error {
	f.byID[c.ID] = c
	return nil
}

type fakeProductRepo struct {
	product.ProductRepository
	byID map[uuid.UUID]*product.Product
}

func (f *fakeProductRepo) GetByID(_ context.Context, id uuid.UUID) (*product.Product, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return p, nil
}

type fakeStock struct {
	err      error
	reserved []inventory.ReserveForSaleInput
	voided   []uuid.UUID
}

func (f *fakeStock) ReserveForSale(_ context.Context, in inventory.ReserveForSaleInput) (valueobjects.Money, error) {
	if f.err != nil {
		return valueobjects.Zero(), f.err
	}
	f.reserved = append(f.reserved, in)
	return money("2.00"), nil
}

func (f *fakeStock) ReturnVoidedSale(_ context.Context, saleID uuid.UUID) error {
	f.voided = append(f.voided, saleID)
	return nil
}

type fakeClientOrders struct {
	calls [][]purchasing.ClientOrderLine
}

func (f *fakeClientOrders) CreateClientOrder(_ context.Context, _ uuid.UUID, _ uuid.UUID, lines []purchasing.ClientOrderLine) error {
	f.calls = append(f.calls, lines)
	return nil
}

type harness struct {
	svc          *sales.SalesService
	salesRepo    *fakeSalesRepo
	paymentsRepo *fakePaymentsRepo
	customers    *fakeCustomerRepo
	products     *fakeProductRepo
	stock        *fakeStock
	clientOrders *fakeClientOrders
}

func newHarness() *harness {
	log := logger.NewLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	h := &harness{
		salesRepo:    &fakeSalesRepo{byID: map[uuid.UUID]*sales.Sale{}},
		paymentsRepo: &fakePaymentsRepo{statuses: map[uuid.UUID]string{}},
		customers:    &fakeCustomerRepo{byID: map[uuid.UUID]*customer.Customer{}},
		products:     &fakeProductRepo{byID: map[uuid.UUID]*product.Product{}},
		stock:        &fakeStock{},
		clientOrders: &fakeClientOrders{},
	}
	h.customers.byID[uuid.MustParse("00000000-0000-0000-0000-000000000001")] = &customer.Customer{
		ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		CurrentDebt: valueobjects.Zero(),
	}
	h.svc = sales.New(h.salesRepo, h.paymentsRepo, customer.NewService(h.customers, fakeTx{}, log), product.NewService(h.products, fakeTx{}, log), h.stock, h.clientOrders, fakeTx{}, log)
	return h
}

var testCustomerID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func stockInput(due *time.Time, initial valueobjects.Money) sales.CreateInput {
	return sales.CreateInput{
		CustomerID:     testCustomerID,
		SaleType:       enums.SaleTypeStock,
		Date:           time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
		DueDate:        due,
		PaymentMethod:  enums.PaymentMethodCash,
		Items:          []sales.ItemInput{{ProductID: uuid.New(), Quantity: valueobjects.QuantityFromInt64(2), UnitPrice: money("10.00")}},
		InitialPayment: initial,
	}
}

func TestCreateStockInsufficientStockAborts(t *testing.T) {
	h := newHarness()
	h.stock.err = derrors.ErrInsufficientStock
	_, err := h.svc.Create(context.Background(), stockInput(nil, valueobjects.Zero()))
	if !errors.Is(err, derrors.ErrInsufficientStock) {
		t.Fatalf("want insufficient stock, got %v", err)
	}
	if len(h.salesRepo.created) != 0 {
		t.Fatalf("sale persisted despite stock failure")
	}
	if len(h.paymentsRepo.created) != 0 {
		t.Fatalf("payment persisted despite stock failure")
	}
	if len(h.clientOrders.calls) != 0 {
		t.Fatalf("client order created despite stock failure")
	}
	if debt := h.customers.byID[testCustomerID].CurrentDebt; !debt.IsZero() {
		t.Fatalf("customer debt changed to %s", debt)
	}
}

func TestCreateClientOrderDoesNotReserveStock(t *testing.T) {
	h := newHarness()
	prod := &product.Product{ID: uuid.New(), Description: "Cosa importada", CostUSD: money("4.00")}
	h.products.byID[prod.ID] = prod
	in := stockInput(nil, valueobjects.Zero())
	in.SaleType = enums.SaleTypeClientOrder
	in.Items[0].ProductID = prod.ID
	in.Items[0].UnitPrice = money("50.00")
	due := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	in.DueDate = &due

	res, err := h.svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(h.stock.reserved) != 0 {
		t.Fatalf("stock reserved for client order sale")
	}
	if len(h.clientOrders.calls) != 1 {
		t.Fatalf("client orders called %d times, want 1", len(h.clientOrders.calls))
	}
	line := h.clientOrders.calls[0][0]
	if line.ProductID != prod.ID || line.Description != "Cosa importada" || !line.SalePricePen.Equals(money("50.00")) {
		t.Fatalf("client order line = %+v", line)
	}
	if !res.Items[0].CostSnapshot.Equals(money("4.00")) {
		t.Fatalf("cost snapshot = %s, want 4.00", res.Items[0].CostSnapshot)
	}
	if !res.Sale.Profit.Equals(money("92.00")) {
		t.Fatalf("profit = %s, want 92.00", res.Sale.Profit)
	}
	if res.Sale.Status != enums.SaleStatusPending {
		t.Fatalf("status = %s, want pending", res.Sale.Status)
	}
	if len(h.paymentsRepo.created) != 0 {
		t.Fatalf("payment recorded for credit sale without initial payment")
	}
	if !h.customers.byID[testCustomerID].CurrentDebt.Equals(money("100.00")) {
		t.Fatalf("debt = %s, want 100.00", h.customers.byID[testCustomerID].CurrentDebt)
	}
}

func TestCreateContadoForcesFullPayment(t *testing.T) {
	h := newHarness()
	res, err := h.svc.Create(context.Background(), stockInput(nil, valueobjects.Zero()))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if res.Sale.Status != enums.SaleStatusPaid {
		t.Fatalf("status = %s, want paid", res.Sale.Status)
	}
	if !res.Sale.PaidAmount.Equals(money("20.00")) {
		t.Fatalf("paid = %s, want 20.00", res.Sale.PaidAmount)
	}
	if len(h.paymentsRepo.created) != 1 {
		t.Fatalf("payments = %d, want 1", len(h.paymentsRepo.created))
	}
	if !h.paymentsRepo.created[0].Amount.Equals(money("20.00")) {
		t.Fatalf("payment amount = %s, want 20.00", h.paymentsRepo.created[0].Amount)
	}
	if !h.paymentsRepo.allocs[0][0].AllocatedAmount.Equals(money("20.00")) {
		t.Fatalf("allocation = %s, want 20.00", h.paymentsRepo.allocs[0][0].AllocatedAmount)
	}
	if !h.customers.byID[testCustomerID].CurrentDebt.IsZero() {
		t.Fatalf("debt = %s, want zero", h.customers.byID[testCustomerID].CurrentDebt)
	}

	h2 := newHarness()
	_, err = h2.svc.Create(context.Background(), func() sales.CreateInput {
		in := stockInput(nil, money("5.00"))
		return in
	}())
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("want validation error for partial contado payment, got %v", err)
	}
}

func TestCreateRejectsDueDateBeforeSaleDate(t *testing.T) {
	h := newHarness()
	in := stockInput(nil, valueobjects.Zero())
	past := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	in.DueDate = &past
	_, err := h.svc.Create(context.Background(), in)
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
	if len(h.salesRepo.created) != 0 {
		t.Fatalf("sale persisted despite invalid due date")
	}
}

func TestApplyPaymentUpdatesStatusToPaid(t *testing.T) {
	h := newHarness()
	due := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	res, err := h.svc.Create(context.Background(), stockInput(&due, valueobjects.Zero()))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if res.Sale.Status != enums.SaleStatusPending {
		t.Fatalf("status = %s, want pending", res.Sale.Status)
	}
	remaining, err := h.svc.ApplyPayment(context.Background(), res.Sale.ID, sales.PaymentInput{
		Amount:        money("20.00"),
		PaymentMethod: enums.PaymentMethodTransfer,
		Reference:     "REC-001",
		Date:          time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("apply payment: %v", err)
	}
	if !remaining.IsZero() {
		t.Fatalf("remaining = %s, want zero", remaining)
	}
	sale, err := h.svc.GetByID(context.Background(), res.Sale.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if sale.Status != enums.SaleStatusPaid || !sale.PaidAmount.Equals(money("20.00")) {
		t.Fatalf("status = %s, paid = %s, want paid/20.00", sale.Status, sale.PaidAmount)
	}
	if len(h.paymentsRepo.created) != 1 {
		t.Fatalf("payments = %d, want 1", len(h.paymentsRepo.created))
	}
	if h.paymentsRepo.created[0].PaymentMethod != enums.PaymentMethodTransfer {
		t.Fatalf("payment method = %s, want transfer", h.paymentsRepo.created[0].PaymentMethod)
	}
	if !h.customers.byID[testCustomerID].CurrentDebt.IsZero() {
		t.Fatalf("debt = %s, want zero", h.customers.byID[testCustomerID].CurrentDebt)
	}

	if _, err := h.svc.ApplyPayment(context.Background(), res.Sale.ID, sales.PaymentInput{Amount: money("1.00"), Date: time.Now().UTC()}); !derrors.IsCode(err, derrors.ErrSaleAlreadyPaid.Code()) {
		t.Fatalf("want sale already paid, got %v", err)
	}
}
