package inventory_test

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
	derrors "vfinancy/backend/internal/domain/errors"
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
	"vfinancy/backend/internal/features/sales"
	salespostgres "vfinancy/backend/internal/features/sales/postgres"
	"vfinancy/backend/internal/features/supplier"
	supplierpostgres "vfinancy/backend/internal/features/supplier/postgres"
)

var (
	probeCompany   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	probeBranch    = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	probeWarehouse = uuid.MustParse("00000000-0000-0000-0000-0000000000c1")
)

type probeEnv struct {
	ctx          context.Context
	db           *database.DB
	batches      inventory.InventoryBatchRepository
	movements    inventory.InventoryMovementRepository
	salesSvc     *sales.SalesService
	purchSvc     *purchasing.PurchasingService
	customersSvc *customer.CustomerService
	suppliersSvc *supplier.SupplierService
	productsSvc  *product.ProductService
}

func newProbeEnv(t *testing.T) *probeEnv {
	t.Helper()
	log := logger.New("error", "text", "stdout")
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "probe.db"), database.Options{})
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
	orders := salespostgres.NewSaleRepository(db.DB)
	purchaseOrders := purchasingpostgres.NewPurchaseRepository(db.DB)
	supplierPayments := purchasingpostgres.NewSupplierPaymentRepository(db.DB)

	invSvc := inventory.New(batches, movements, warehouseResolver, productClassifier, txm, log)
	salesSvc := sales.New(orders, customers, invSvc, productClassifier, txm, log)
	purchSvc := purchasing.New(purchaseOrders, supplierPayments, invSvc, txm, log)
	customersSvc := customer.New(customers, txm, log)
	suppliersSvc := supplier.New(suppliers, txm, log)
	productsSvc := product.New(products, txm, log)

	return &probeEnv{
		ctx:          ctx,
		db:           db,
		batches:      batches,
		movements:    movements,
		salesSvc:     salesSvc,
		purchSvc:     purchSvc,
		customersSvc: customersSvc,
		suppliersSvc: suppliersSvc,
		productsSvc:  productsSvc,
	}
}

func (e *probeEnv) batchQty(t *testing.T, id uuid.UUID) valueobjects.Quantity {
	t.Helper()
	b, err := e.batches.GetByID(e.ctx, id)
	if err != nil {
		t.Fatalf("get batch %s: %v", id, err)
	}
	return b.CurrentQuantity
}

func (e *probeEnv) batchForLine(t *testing.T, lineID uuid.UUID) *inventory.InventoryBatch {
	t.Helper()
	page, err := e.batches.List(e.ctx, inventory.InventoryBatchFilter{
		CompanyID:      &probeCompany,
		PurchaseLineID: &lineID,
		PageRequest:    repositories.PageRequest{Limit: 10},
	})
	if err != nil {
		t.Fatalf("list batches by line: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected exactly one batch for line %s, got %d", lineID, len(page.Items))
	}
	return page.Items[0]
}

func (e *probeEnv) movementsRef(t *testing.T, refType string, refID uuid.UUID) []*inventory.InventoryMovement {
	t.Helper()
	page, err := e.movements.List(e.ctx, inventory.InventoryMovementFilter{
		CompanyID:     &probeCompany,
		ReferenceType: refType,
		ReferenceID:   &refID,
		PageRequest:   repositories.PageRequest{Limit: 100},
	})
	if err != nil {
		t.Fatalf("list movements: %v", err)
	}
	return page.Items
}

func (e *probeEnv) scalarID(t *testing.T, q string, args ...any) uuid.UUID {
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

func money(s string) valueobjects.Money {
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		panic(err)
	}
	return m
}

func qty(s string) valueobjects.Quantity {
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		panic(err)
	}
	return q
}

func pct(s string) valueobjects.Percentage {
	p, err := valueobjects.PercentageFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func pen(s string) valueobjects.CurrencyCode {
	c, err := valueobjects.NewCurrencyCode(s)
	if err != nil {
		panic(err)
	}
	return c
}

func TestBridgeEndToEnd(t *testing.T) {
	env := newProbeEnv(t)
	ctx := env.ctx
	today := valueobjects.NewDateFromTime(time.Now().UTC())
	yesterday := valueobjects.NewDateFromTime(time.Now().UTC().AddDate(0, 0, -1))

	for _, table := range []string{
		"customers", "suppliers", "products",
		"sales", "sale_items", "purchase_orders", "purchase_order_items",
		"inventory_batches", "inventory_movements",
	} {
		var n int
		if err := env.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if n != 0 {
			t.Fatalf("table %s has %d rows: seed data is still being loaded", table, n)
		}
	}

	taxID := env.scalarID(t, "SELECT id FROM taxes WHERE company_id = $1 AND code = 'IGV' LIMIT 1", probeCompany)
	unitID := env.scalarID(t, "SELECT id FROM units_of_measure WHERE company_id = $1 AND code = 'UND' LIMIT 1", probeCompany)

	cust, err := env.customersSvc.Create(ctx, customer.CreateInput{
		CompanyID:       probeCompany,
		DocumentType:    enums.DocumentTypeRUC,
		DocumentNumber:  "20612345678",
		BusinessName:    "Probe Customer S.A.C.",
		TaxCategory:     enums.TaxCategoryTaxed,
		CreditLimit:     money("50000.00"),
		PaymentTermDays: 30,
		Email:           "probe@customer.com",
		Address:         "Av. Probe 100, Lima",
		BranchID:        &probeBranch,
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	sup, err := env.suppliersSvc.Create(ctx, supplier.CreateInput{
		CompanyID:       probeCompany,
		DocumentType:    enums.DocumentTypeRUC,
		DocumentNumber:  "20101234001",
		BusinessName:    "Probe Supplier S.A.C.",
		TradeName:       "Probe Supplier",
		TaxID:           "20101234001",
		DefaultCurrency: pen("PEN"),
		PaymentTermDays: 30,
		Email:           "probe@supplier.com",
		Address:         "Av. Probe 200, Lima",
	})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}

	sku, err := valueobjects.NewSKU("PROBE-CAFE-001")
	if err != nil {
		t.Fatalf("new sku: %v", err)
	}
	prod, err := env.productsSvc.Create(ctx, product.CreateInput{
		CompanyID:    probeCompany,
		SKU:          sku,
		Description:  "Café Probe",
		UnitID:       unitID,
		TaxID:        taxID,
		Cost:         money("9.80"),
		SalePrice:    money("15.00"),
		SaleCurrency: pen("PEN"),
		MinStock:     qty("0"),
		MaxStock:     qty("0"),
		IsService:    false,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	po1, err := env.purchSvc.Create(ctx, purchasing.CreateInput{
		CompanyID:    probeCompany,
		SupplierID:   sup.ID,
		CurrencyCode: pen("PEN"),
		ExchangeRate: valueobjects.One(),
		OrderDate:    yesterday,
		Number:       "OC-PROBE-0001",
		Items: []purchasing.CreateItemInput{
			{
				ProductID:   prod.ID,
				Quantity:    qty("125"),
				UnitPrice:   money("9.80"),
				TaxRate:     pct("18"),
				TaxAmount:   money("220.50"),
				Description: "Café Probe",
			},
		},
	})
	if err != nil {
		t.Fatalf("create purchase 1: %v", err)
	}
	line1 := po1.Items[0].ID
	batchA := env.batchForLine(t, line1)
	if !batchA.CurrentQuantity.Equals(qty("125.0000")) {
		t.Fatalf("batch A qty = %s, want 125.0000", batchA.CurrentQuantity)
	}
	if !batchA.UnitCost.Equals(money("9.80")) {
		t.Fatalf("batch A unit cost = %s, want 9.80", batchA.UnitCost)
	}
	if batchA.WarehouseID != probeWarehouse {
		t.Fatalf("warehouse = %s, want %s", batchA.WarehouseID, probeWarehouse)
	}

	po2, err := env.purchSvc.Create(ctx, purchasing.CreateInput{
		CompanyID:    probeCompany,
		SupplierID:   sup.ID,
		CurrencyCode: pen("PEN"),
		ExchangeRate: valueobjects.One(),
		OrderDate:    today,
		Number:       "OC-PROBE-0002",
		Items: []purchasing.CreateItemInput{
			{
				ProductID:      prod.ID,
				Quantity:       qty("50"),
				UnitPrice:      money("12.00"),
				DiscountAmount: money("75.00"),
				TaxRate:        pct("18"),
				TaxAmount:      money("108.00"),
				Description:    "Café Probe",
			},
		},
	})
	if err != nil {
		t.Fatalf("create purchase 2: %v", err)
	}
	line2 := po2.Items[0].ID
	batchB := env.batchForLine(t, line2)
	if !batchB.CurrentQuantity.Equals(qty("50.0000")) {
		t.Fatalf("batch B qty = %s, want 50.0000", batchB.CurrentQuantity)
	}
	if !batchB.UnitCost.Equals(money("10.50")) {
		t.Fatalf("batch B unit cost = %s, want 10.50", batchB.UnitCost)
	}
	if string(batchB.LotNumber) != po2.Number {
		t.Fatalf("lot = %q, want %q", batchB.LotNumber, po2.Number)
	}
	if got := env.batchQty(t, batchA.ID); !got.Equals(qty("125.0000")) {
		t.Fatalf("batch A changed by second receipt: %s", got)
	}
	recv := env.movementsRef(t, "purchase", line2)
	if len(recv) != 1 || recv[0].Type != enums.MovementTypePurchaseReceipt {
		t.Fatalf("expected 1 purchase_receipt movement, got %d", len(recv))
	}
	if !recv[0].Quantity.Equals(qty("50.0000")) {
		t.Fatalf("purchase_receipt qty = %s, want 50.0000", recv[0].Quantity)
	}

	if err := env.purchSvc.Approve(ctx, po2.ID); err != nil {
		t.Fatalf("approve purchase 2: %v", err)
	}
	env.batchForLine(t, line2)

	res, err := env.salesSvc.Create(ctx, sales.CreateInput{
		CompanyID:    probeCompany,
		Number:       "V-PROBE-0001",
		CustomerID:   cust.ID,
		CurrencyCode: pen("PEN"),
		ExchangeRate: valueobjects.One(),
		DueDate:      &today,
		Items: []sales.CreateItemInput{
			{
				ProductID:   prod.ID,
				Quantity:    qty("160"),
				UnitPrice:   money("15.00"),
				TaxRate:     pct("18"),
				TaxAmount:   money("432.00"),
				Description: "Café Probe",
			},
		},
	})
	if err != nil {
		t.Fatalf("create sale: %v", err)
	}
	saleID := res.Sale.ID
	if got := res.Sale.Items[0].CostSnapshot; !got.Equals(money("9.95")) {
		t.Fatalf("cost snapshot = %s, want 9.95", got)
	}
	if got := env.batchQty(t, batchA.ID); !got.Equals(qty("0.0000")) {
		t.Fatalf("batch A after sale = %s, want 0", got)
	}
	if got := env.batchQty(t, batchB.ID); !got.Equals(qty("15.0000")) {
		t.Fatalf("batch B after sale = %s, want 15.0000", got)
	}
	refSale := env.movementsRef(t, "sale", saleID)
	if len(refSale) != 2 {
		t.Fatalf("expected 2 sale movements, got %d", len(refSale))
	}
	byBatch := map[uuid.UUID]*inventory.InventoryMovement{}
	for _, m := range refSale {
		if m.Type != enums.MovementTypeSale || m.BatchID == nil {
			t.Fatalf("unexpected movement %s", m.Type)
		}
		byBatch[*m.BatchID] = m
	}
	mA := byBatch[batchA.ID]
	if mA == nil || !mA.Quantity.Equals(qty("-125.0000")) {
		t.Fatalf("sale movement for A: %v", mA)
	}
	mB := byBatch[batchB.ID]
	if mB == nil || !mB.Quantity.Equals(qty("-35.0000")) {
		t.Fatalf("sale movement for B: %v", mB)
	}
	if !mB.UnitCost.Equals(money("10.50")) {
		t.Fatalf("batch B sale unit cost = %s, want 10.50", mB.UnitCost)
	}

	_, err = env.salesSvc.Create(ctx, sales.CreateInput{
		CompanyID:    probeCompany,
		Number:       "V-PROBE-0002",
		CustomerID:   cust.ID,
		CurrencyCode: pen("PEN"),
		ExchangeRate: valueobjects.One(),
		DueDate:      &today,
		Items: []sales.CreateItemInput{
			{
				ProductID:   prod.ID,
				Quantity:    qty("200"),
				UnitPrice:   money("15.00"),
				TaxRate:     pct("18"),
				TaxAmount:   money("540.00"),
				Description: "Café Probe",
			},
		},
	})
	if err == nil {
		t.Fatal("expected shortfall error")
	}
	if !derrors.IsCode(err, derrors.ErrInsufficientStock.Code()) {
		t.Fatalf("expected INSUFFICIENT_STOCK, got %v", err)
	}
	if got := env.batchQty(t, batchA.ID); !got.Equals(qty("0.0000")) {
		t.Fatalf("shortfall rolled back batch A to %s", got)
	}
	if got := env.batchQty(t, batchB.ID); !got.Equals(qty("15.0000")) {
		t.Fatalf("shortfall rolled back batch B to %s", got)
	}

	if _, err := env.salesSvc.Cancel(ctx, sales.CancelInput{ID: saleID, Reason: "probe"}); err != nil {
		t.Fatalf("cancel sale: %v", err)
	}
	if got := env.batchQty(t, batchA.ID); !got.Equals(qty("125.0000")) {
		t.Fatalf("batch A after sale void = %s, want 125", got)
	}
	if got := env.batchQty(t, batchB.ID); !got.Equals(qty("50.0000")) {
		t.Fatalf("batch B after sale void = %s, want 50", got)
	}
	refSale = env.movementsRef(t, "sale", saleID)
	if len(refSale) != 4 {
		t.Fatalf("expected 4 movements after void, got %d", len(refSale))
	}
	if !hasMovement(refSale, batchA.ID, enums.MovementTypeVoidSale, "125.0000") {
		t.Fatal("missing void_sale +125 for batch A")
	}
	if !hasMovement(refSale, batchB.ID, enums.MovementTypeVoidSale, "35.0000") {
		t.Fatal("missing void_sale +35 for batch B")
	}

	if err := env.purchSvc.Cancel(ctx, po2.ID, "probe"); err != nil {
		t.Fatalf("cancel purchase 2: %v", err)
	}
	if got := env.batchQty(t, batchB.ID); !got.Equals(qty("0.0000")) {
		t.Fatalf("batch B after purchase void = %s, want 0", got)
	}
	if got := env.batchQty(t, batchA.ID); !got.Equals(qty("125.0000")) {
		t.Fatalf("batch A changed by purchase void: %s", got)
	}
	refPurch := env.movementsRef(t, "purchase", line2)
	if len(refPurch) != 2 {
		t.Fatalf("expected 2 purchase movements after void, got %d", len(refPurch))
	}
	if !hasMovement(refPurch, batchB.ID, enums.MovementTypeVoidPurchase, "-50.0000") {
		t.Fatalf("missing void_purchase -50 for batch %s", batchB.ID)
	}
}

func hasMovement(items []*inventory.InventoryMovement, batchID uuid.UUID, typ enums.InventoryMovementType, qtyStr string) bool {
	for _, m := range items {
		if m.BatchID != nil && *m.BatchID == batchID && m.Type == typ && m.Quantity.Equals(qty(qtyStr)) {
			return true
		}
	}
	return false
}
