package purchasing_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/infrastructure/sqlite"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/customer"
	customerpostgres "vfinancy/backend/internal/features/customer/postgres"
	"vfinancy/backend/internal/features/inventory"
	inventorypostgres "vfinancy/backend/internal/features/inventory/postgres"
	"vfinancy/backend/internal/features/product"
	productpostgres "vfinancy/backend/internal/features/product/postgres"
	"vfinancy/backend/internal/features/purchasing"
	purchasingpostgres "vfinancy/backend/internal/features/purchasing/postgres"
	"vfinancy/backend/internal/features/supplier"
	supplierpostgres "vfinancy/backend/internal/features/supplier/postgres"
)

const coCompanyID = "00000000-0000-0000-0000-000000000001"

type coEnv struct {
	ctx          context.Context
	db           *database.DB
	purchSvc     *purchasing.PurchasingService
	customersSvc *customer.CustomerService
	suppliersSvc *supplier.SupplierService
	productsSvc  *product.ProductService
	orders       purchasing.PurchaseRepository
	companyID    uuid.UUID
	cardID       uuid.UUID
}

func newCoEnv(t *testing.T) *coEnv {
	t.Helper()
	log := logger.New("error", "text", "stdout")
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "co.db"), database.Options{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	persistence.SetDialect(persistence.DialectSQLite)
	runner := migrations.NewRunner("../../../../backend/migrations/sqlite", db.DB, log, "sqlite")
	if err := runner.Up(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	txm := persistence.NewTxManager(db)
	customers := customerpostgres.NewCustomerRepository(db.DB)
	suppliers := supplierpostgres.NewSupplierRepository(db.DB)
	products := productpostgres.NewProductRepository(db.DB)
	batches := inventorypostgres.NewInventoryBatchRepository(db.DB)
	movements := inventorypostgres.NewInventoryMovementRepository(db.DB)
	warehouseResolver := inventorypostgres.NewWarehouseResolver(db.DB)
	productClassifier := inventorypostgres.NewProductClassifier(db.DB)
	orders := purchasingpostgres.NewPurchaseRepository(db.DB)
	supplierPayments := purchasingpostgres.NewSupplierPaymentRepository(db.DB)

	invSvc := inventory.New(batches, movements, warehouseResolver, productClassifier, txm, log)
	purchSvc := purchasing.New(orders, supplierPayments, invSvc, txm, log)
	customersSvc := customer.New(customers, txm, log)
	suppliersSvc := supplier.New(suppliers, txm, log)
	productsSvc := product.New(products, txm, log)

	return &coEnv{
		ctx:          ctx,
		db:           db,
		purchSvc:     purchSvc,
		customersSvc: customersSvc,
		suppliersSvc: suppliersSvc,
		productsSvc:  productsSvc,
		orders:       orders,
		companyID:    uuid.MustParse(coCompanyID),
		cardID:       uuid.MustParse("00000000-0000-0000-0000-00000000ca11"),
	}
}

func (e *coEnv) scalarID(t *testing.T, q string, args ...any) uuid.UUID {
	t.Helper()
	var s string
	if err := e.db.QueryRowContext(e.ctx, q, args...).Scan(&s); err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("bad uuid %q: %v", s, err)
	}
	return id
}

func (e *coEnv) seedCustomerAndSupplier(t *testing.T) (uuid.UUID, uuid.UUID) {
	t.Helper()
	cust, err := e.customersSvc.Create(e.ctx, customer.CreateInput{
		CompanyID:       e.companyID,
		DocumentType:    enums.DocumentTypeRUC,
		DocumentNumber:  "20611111111",
		BusinessName:    "CO Customer S.A.C.",
		TaxCategory:     enums.TaxCategoryTaxed,
		CreditLimit:     coMoney("50000.00"),
		PaymentTermDays: 30,
		Email:           "co@customer.com",
		Address:         "Av. CO 100, Lima",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}
	sup, err := e.suppliersSvc.Create(e.ctx, supplier.CreateInput{
		CompanyID:       e.companyID,
		DocumentType:    enums.DocumentTypeRUC,
		DocumentNumber:  "20105555555",
		BusinessName:    "CO Supplier S.A.C.",
		TradeName:       "CO Supplier",
		TaxID:           "20105555555",
		DefaultCurrency: coPen("PEN"),
		PaymentTermDays: 30,
		Email:           "co@supplier.com",
		Address:         "Av. CO 200, Lima",
	})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}
	return cust.ID, sup.ID
}

func (e *coEnv) seedProduct(t *testing.T) uuid.UUID {
	t.Helper()
	taxID := e.scalarID(t, "SELECT id FROM taxes WHERE company_id = $1 AND code = 'IGV' LIMIT 1", e.companyID)
	unitID := e.scalarID(t, "SELECT id FROM units_of_measure WHERE company_id = $1 AND code = 'UND' LIMIT 1", e.companyID)
	sku, err := valueobjects.NewSKU("CO-PROD-001")
	if err != nil {
		t.Fatalf("new sku: %v", err)
	}
	prod, err := e.productsSvc.Create(e.ctx, product.CreateInput{
		CompanyID:    e.companyID,
		SKU:          sku,
		Description:  "CO Producto",
		UnitID:       unitID,
		TaxID:        taxID,
		CostUSD:      coMoney("10.00"),
		SalePrice:    coMoney("15.00"),
		SaleCurrency: coPen("PEN"),
		MinStock:     coQty("0"),
		MaxStock:     coQty("0"),
		IsService:    false,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	return prod.ID
}

func (e *coEnv) newOrder(t *testing.T, customerID, supplierID, productID uuid.UUID, salePrice, anticipo string) *purchasing.PurchaseOrder {
	t.Helper()
	po, err := e.purchSvc.Create(e.ctx, purchasing.CreateInput{
		CompanyID:    e.companyID,
		SupplierID:   supplierID,
		CustomerID:   &customerID,
		CreditCardID: &e.cardID,
		OrderType:    enums.OrderTypeCustomer,
		CurrencyCode: coPen("PEN"),
		ExchangeRate: valueobjects.One(),
		OrderDate:    valueobjects.NewDateFromTime(time.Now().UTC()),
		CostUSD:      coMoney("500.00"),
		SalePricePEN: coMoney(salePrice),
		Anticipo:     coMoney(anticipo),
		AnticipoDate: coDatePtr(time.Now().UTC()),
		Items: []purchasing.CreateItemInput{
			{
				ProductID:   productID,
				Quantity:    coQty("50"),
				UnitPrice:   coMoney("10.00"),
				TaxRate:     coPct("18"),
				TaxAmount:   coMoney("90.00"),
				Description: "CO Producto",
			},
		},
	})
	if err != nil {
		t.Fatalf("create customer order: %v", err)
	}
	return po
}

func coMoney(s string) valueobjects.Money {
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		panic(err)
	}
	return m
}

func coQty(s string) valueobjects.Quantity {
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		panic(err)
	}
	return q
}

func coPct(s string) valueobjects.Percentage {
	p, err := valueobjects.PercentageFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func coPen(s string) valueobjects.CurrencyCode {
	c, err := valueobjects.NewCurrencyCode(s)
	if err != nil {
		panic(err)
	}
	return c
}

func coDatePtr(t time.Time) *valueobjects.Date {
	d := valueobjects.NewDateFromTime(t)
	return &d
}

func TestCustomerOrderLifecycle(t *testing.T) {
	env := newCoEnv(t)
	ctx := env.ctx
	customerID, supplierID := env.seedCustomerAndSupplier(t)
	productID := env.seedProduct(t)

	po := env.newOrder(t, customerID, supplierID, productID, "1000.00", "200.00")

	if po.OrderType != enums.OrderTypeCustomer {
		t.Fatalf("order type = %s, want customer", po.OrderType)
	}
	if po.CustomerID == nil || *po.CustomerID != customerID {
		t.Fatalf("customer id not set on the order")
	}
	if !po.Anticipo.Equals(coMoney("200.00")) {
		t.Fatalf("anticipo = %s, want 200.00", po.Anticipo)
	}
	if want := coMoney("800.00"); !po.PorCobrar().Equals(want) {
		t.Fatalf("por cobrar = %s, want 800.00", po.PorCobrar())
	}
	if want := coMoney("535.00"); !po.RealCostPEN.Equals(want) {
		t.Fatalf("real cost = %s, want 535.00 (500 * 1.07)", po.RealCostPEN)
	}
	if want := coMoney("465.00"); !po.ProjectedProfitPEN.Equals(want) {
		t.Fatalf("projected profit = %s, want 465.00", po.ProjectedProfitPEN)
	}

	loaded, err := env.orders.GetByID(ctx, po.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if len(loaded.CustomerPayments) != 1 {
		t.Fatalf("expected 1 initial anticipo payment, got %d", len(loaded.CustomerPayments))
	}
	initial := loaded.CustomerPayments[0]
	if !initial.Amount.Equals(coMoney("200.00")) || initial.IsRefunded() {
		t.Fatalf("initial payment wrong: amount=%s refunded=%v", initial.Amount, initial.IsRefunded())
	}

	pm, err := env.purchSvc.RegisterCustomerOrderPayment(ctx, po.ID, purchasing.CustomerPaymentInput{
		CompanyID:    env.companyID,
		PaymentDate:  valueobjects.NewDateFromTime(time.Now().UTC()),
		Amount:       coMoney("800.00"),
		CurrencyCode: coPen("PEN"),
		ExchangeRate: valueobjects.One(),
		Method:       enums.PaymentMethodBankTransfer,
		Reference:    "Pago final",
	})
	if err != nil {
		t.Fatalf("register final payment: %v", err)
	}
	if pm.PurchaseOrderID != po.ID {
		t.Fatalf("payment belongs to %s, want %s", pm.PurchaseOrderID, po.ID)
	}

	final, err := env.orders.GetByID(ctx, po.ID)
	if err != nil {
		t.Fatalf("get order after payment: %v", err)
	}
	if !final.Anticipo.Equals(coMoney("1000.00")) {
		t.Fatalf("anticipo after final payment = %s, want 1000.00", final.Anticipo)
	}
	if !final.PorCobrar().Equals(valueobjects.Zero()) {
		t.Fatalf("por cobrar after final payment = %s, want 0.00", final.PorCobrar())
	}
	if len(final.CustomerPayments) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(final.CustomerPayments))
	}

	if _, err := env.purchSvc.RegisterCustomerOrderPayment(ctx, po.ID, purchasing.CustomerPaymentInput{
		CompanyID:    env.companyID,
		PaymentDate:  valueobjects.NewDateFromTime(time.Now().UTC()),
		Amount:       coMoney("1.00"),
		CurrencyCode: coPen("PEN"),
		ExchangeRate: valueobjects.One(),
		Method:       enums.PaymentMethodCash,
	}); err == nil {
		t.Fatalf("expected error when paying a fully paid order")
	}
}

func TestCustomerOrderMarkFaultyRefundsAnticipos(t *testing.T) {
	env := newCoEnv(t)
	ctx := env.ctx
	customerID, supplierID := env.seedCustomerAndSupplier(t)
	productID := env.seedProduct(t)

	po := env.newOrder(t, customerID, supplierID, productID, "1200.00", "300.00")

	out, err := env.purchSvc.MarkFaulty(ctx, purchasing.FaultyInput{
		ID:          po.ID,
		ArrivalDate: valueobjects.NewDateFromTime(time.Now().UTC()),
		Reason:      "Caja abollada al llegar",
	})
	if err != nil {
		t.Fatalf("mark faulty: %v", err)
	}
	if !out.IsCancelled() {
		t.Fatalf("order should be cancelled after faulty arrival, got %s", out.Status)
	}
	if !out.Faulty {
		t.Fatalf("order should be flagged as faulty")
	}
	if !out.RefundedAmount.Equals(coMoney("300.00")) {
		t.Fatalf("refunded amount = %s, want 300.00", out.RefundedAmount)
	}

	loaded, err := env.orders.GetByID(ctx, po.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if len(loaded.CustomerPayments) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(loaded.CustomerPayments))
	}
	if !loaded.CustomerPayments[0].IsRefunded() {
		t.Fatalf("expected the anticipo to be refunded")
	}
	if loaded.CustomerPayments[0].RefundReason == "" {
		t.Fatalf("expected a refund reason")
	}
}

func TestCustomerOrderCancelRefundsAnticipos(t *testing.T) {
	env := newCoEnv(t)
	ctx := env.ctx
	customerID, supplierID := env.seedCustomerAndSupplier(t)
	productID := env.seedProduct(t)

	po := env.newOrder(t, customerID, supplierID, productID, "900.00", "150.00")

	if err := env.purchSvc.Cancel(ctx, po.ID, "El cliente desistió"); err != nil {
		t.Fatalf("cancel order: %v", err)
	}

	loaded, err := env.orders.GetByID(ctx, po.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if !loaded.IsCancelled() {
		t.Fatalf("order should be cancelled")
	}
	if !loaded.RefundedAmount.Equals(coMoney("150.00")) {
		t.Fatalf("refunded amount = %s, want 150.00", loaded.RefundedAmount)
	}
	if len(loaded.CustomerPayments) != 1 || !loaded.CustomerPayments[0].IsRefunded() {
		t.Fatalf("expected the anticipo to be refunded")
	}
}

func TestCustomerOrderListFilterByOrderType(t *testing.T) {
	env := newCoEnv(t)
	ctx := env.ctx
	customerID, supplierID := env.seedCustomerAndSupplier(t)
	productID := env.seedProduct(t)

	env.newOrder(t, customerID, supplierID, productID, "1000.00", "0.00")

	if _, err := env.purchSvc.Create(ctx, purchasing.CreateInput{
		CompanyID:    env.companyID,
		SupplierID:   supplierID,
		CreditCardID: &env.cardID,
		OrderType:    enums.OrderTypeGeneral,
		CurrencyCode: coPen("PEN"),
		ExchangeRate: valueobjects.One(),
		OrderDate:    valueobjects.NewDateFromTime(time.Now().UTC()),
		Items: []purchasing.CreateItemInput{
			{ProductID: productID, Quantity: coQty("10"), UnitPrice: coMoney("10.00"), Description: "CO Producto"},
		},
	}); err != nil {
		t.Fatalf("create general order: %v", err)
	}

	customers, err := env.orders.List(ctx, purchasing.PurchaseFilter{
		CompanyID:   &env.companyID,
		OrderType:   enums.OrderTypeCustomer.String(),
		PageRequest: repositories.PageRequest{Limit: 25},
	})
	if err != nil {
		t.Fatalf("list customer orders: %v", err)
	}
	if len(customers.Items) != 1 {
		t.Fatalf("expected 1 customer order, got %d", len(customers.Items))
	}
	if customers.Items[0].CustomerID == nil {
		t.Fatalf("customer order should carry the customer id")
	}

	general, err := env.orders.List(ctx, purchasing.PurchaseFilter{
		CompanyID:   &env.companyID,
		OrderType:   enums.OrderTypeGeneral.String(),
		PageRequest: repositories.PageRequest{Limit: 25},
	})
	if err != nil {
		t.Fatalf("list general orders: %v", err)
	}
	if len(general.Items) != 1 {
		t.Fatalf("expected 1 general order, got %d", len(general.Items))
	}
}
