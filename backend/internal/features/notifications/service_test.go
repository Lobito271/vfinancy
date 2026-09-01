package notifications_test

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
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
	inventorypostgres "vfinancy/backend/internal/features/inventory/postgres"
	"vfinancy/backend/internal/features/notifications"
	notificationspostgres "vfinancy/backend/internal/features/notifications/postgres"
	"vfinancy/backend/internal/features/product"
	productpostgres "vfinancy/backend/internal/features/product/postgres"
)

var (
	ntfCompany   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	ntfBranch    = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	ntfWarehouse = uuid.MustParse("00000000-0000-0000-0000-0000000000c1")
)

type ntfEnv struct {
	ctx         context.Context
	db          *database.DB
	svc         *notifications.NotificationsService
	invSvc      *inventory.InventoryService
	productsSvc *product.ProductService
	batches     inventory.InventoryBatchRepository
	companyID   uuid.UUID
}

func newNtfEnv(t *testing.T) *ntfEnv {
	t.Helper()
	log := logger.New("error", "text", "stdout")
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "ntf.db"), database.Options{})
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
	_, err = db.ExecContext(ctx, `INSERT INTO companies (id, code, legal_name, trade_name, tax_id) VALUES (?, ?, ?, ?, ?)`, ntfCompany, "NTF", "Ntf Company S.A.C.", "Ntf Company", "20600000002")
	if err == nil {
		_, err = db.ExecContext(ctx, `INSERT INTO branches (id, company_id, code, name, is_default) VALUES (?, ?, ?, ?, TRUE)`, ntfBranch, ntfCompany, "MAIN", "Principal")
	}
	if err == nil {
		_, err = db.ExecContext(ctx, `INSERT INTO taxes (id, company_id, code, name, short_name, country_code, default_rate) VALUES (?, ?, 'IGV', 'Impuesto General a las Ventas', 'IGV', 'PE', '0.18')`, uuid.New(), ntfCompany)
	}
	if err == nil {
		_, err = db.ExecContext(ctx, `INSERT INTO units_of_measure (id, company_id, code, name, symbol, allows_decimals) VALUES (?, ?, 'UND', 'Unidad', 'UND', FALSE)`, uuid.MustParse("00000000-0000-0000-0000-0000000000d1"), ntfCompany)
	}
	if err == nil {
		_, err = db.ExecContext(ctx, `INSERT INTO warehouses (id, company_id, branch_id, code, name, is_default, is_active) VALUES (?, ?, ?, 'ALM-01', 'Almacén Principal', TRUE, TRUE)`, ntfWarehouse, ntfCompany, ntfBranch)
	}
	if err != nil {
		t.Fatalf("seed references: %v", err)
	}

	txm := persistence.NewTxManager(db)
	batches := inventorypostgres.NewInventoryBatchRepository(db.DB)
	movements := inventorypostgres.NewInventoryMovementRepository(db.DB)
	warehouseResolver := inventorypostgres.NewWarehouseResolver(db.DB)
	productClassifier := inventorypostgres.NewProductClassifier(db.DB)
	products := productpostgres.NewProductRepository(db.DB)

	invSvc := inventory.New(batches, movements, warehouseResolver, productClassifier, txm, log)
	productsSvc := product.New(products, txm, log)

	svc := notifications.New(notificationspostgres.NewNotificationRepository(db.DB), log)
	svc.SetClearanceSource(invSvc)
	svc.SetClearanceDays(func(context.Context, uuid.UUID) int { return inventory.ClearanceDays })
	svc.SetProductInfo(func(ctx context.Context, productID uuid.UUID) (string, string, error) {
		p, err := productsSvc.GetByID(ctx, productID)
		if err != nil {
			return "", "", err
		}
		return p.Description, p.SKU.String(), nil
	})
	svc.SetActiveCompany(func() uuid.UUID { return ntfCompany })

	return &ntfEnv{
		ctx: ctx, db: db, svc: svc, invSvc: invSvc,
		productsSvc: productsSvc, batches: batches,
		companyID: ntfCompany,
	}
}

func (e *ntfEnv) createProduct(t *testing.T) *product.Product {
	t.Helper()
	taxID := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	unitID := uuid.MustParse("00000000-0000-0000-0000-0000000000d1")
	if err := e.db.QueryRowContext(e.ctx,
		`SELECT id FROM taxes WHERE company_id = ? LIMIT 1`, e.companyID).Scan(&taxID); err != nil {
		t.Fatalf("lookup tax: %v", err)
	}
	sku, err := valueobjects.NewSKU("NTF-PROD-001")
	if err != nil {
		t.Fatalf("new sku: %v", err)
	}
	p, err := e.productsSvc.Create(e.ctx, product.CreateInput{
		CompanyID:    e.companyID,
		SKU:          sku,
		Description:  "Producto Ntf",
		UnitID:       unitID,
		TaxID:        taxID,
		CostUSD:      ntfMoney("9.80"),
		SalePrice:    ntfMoney("15.00"),
		SaleCurrency: ntfPen("PEN"),
		MinStock:     ntfQty("0"),
		MaxStock:     ntfQty("0"),
		IsService:    false,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	return p
}

// createStaleBatch inserts a batch that arrived `ageDays` days ago with
// stock, bypassing the purchase flow.
func (e *ntfEnv) createStaleBatch(t *testing.T, productID uuid.UUID, ageDays int) *inventory.InventoryBatch {
	t.Helper()
	lot, err := valueobjects.NewLotNumber("LOT-NTF-0001")
	if err != nil {
		t.Fatalf("new lot: %v", err)
	}
	b, err := inventory.NewInventoryBatch(time.Now().UTC(), inventory.NewInventoryBatchOptions{
		CompanyID:       e.companyID,
		ProductID:       productID,
		WarehouseID:     ntfWarehouse,
		LotNumber:       lot,
		ArrivalDate:     valueobjects.NewDateFromTime(time.Now().UTC().AddDate(0, 0, -ageDays)),
		InitialQuantity: ntfQty("10"),
		UnitCost:        ntfMoney("5.00"),
		CurrencyCode:    ntfPen("PEN"),
	})
	if err != nil {
		t.Fatalf("new batch: %v", err)
	}
	if err := e.batches.Create(e.ctx, b); err != nil {
		t.Fatalf("create batch: %v", err)
	}
	return b
}

func (e *ntfEnv) unread(t *testing.T) int {
	t.Helper()
	n, err := e.svc.UnreadCount(e.ctx, e.companyID)
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	return n
}

func (e *ntfEnv) listed(t *testing.T, onlyUnread bool) []*notifications.Notification {
	t.Helper()
	page, err := e.svc.List(e.ctx, notifications.ListFilter{CompanyID: e.companyID, UnreadOnly: onlyUnread})
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	return page.Items
}

func TestClearanceNotificationLifecycle(t *testing.T) {
	env := newNtfEnv(t)
	prod := env.createProduct(t)
	batch := env.createStaleBatch(t, prod.ID, 30)

	if env.unread(t) != 0 {
		t.Fatal("expected no notifications before generation")
	}

	created, err := env.svc.Generate(env.ctx)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if created != 1 {
		t.Fatalf("created = %d, want 1", created)
	}
	if env.unread(t) != 1 {
		t.Fatalf("unread = %d, want 1", env.unread(t))
	}

	items := env.listed(t, false)
	if len(items) != 1 {
		t.Fatalf("listed = %d, want 1", len(items))
	}
	n := items[0]
	if n.Type != notifications.TypeClearance || n.RecordType != notifications.RecordTypeBatch {
		t.Fatalf("unexpected notification type %s / %s", n.Type, n.RecordType)
	}
	if n.RecordID == nil || *n.RecordID != batch.ID || n.DedupKey != batch.ID.String() {
		t.Fatalf("notification not tied to the batch: %+v", n)
	}
	if n.Title != "Producto en remate" {
		t.Fatalf("title = %q", n.Title)
	}
	if n.ReadAt != nil {
		t.Fatal("fresh notification must be unread")
	}

	if again, err := env.svc.Generate(env.ctx); err != nil || again != 0 {
		t.Fatalf("second generate created %d (err=%v), want 0", again, err)
	}
	if env.unread(t) != 1 {
		t.Fatalf("unread after regen = %d, want 1", env.unread(t))
	}

	if _, err := env.svc.MarkAllRead(env.ctx, env.companyID); err != nil {
		t.Fatalf("mark all read: %v", err)
	}
	if env.unread(t) != 0 {
		t.Fatalf("unread after markAll = %d, want 0", env.unread(t))
	}
	items = env.listed(t, false)
	if len(items) != 1 || items[0].ReadAt == nil {
		t.Fatalf("read notification lost: %+v", items)
	}
}

func TestStaleClearanceNotificationsRemoved(t *testing.T) {
	env := newNtfEnv(t)
	prod := env.createProduct(t)
	batch := env.createStaleBatch(t, prod.ID, 30)

	if _, err := env.svc.Generate(env.ctx); err != nil {
		t.Fatalf("generate: %v", err)
	}

	if _, err := env.db.ExecContext(env.ctx,
		`UPDATE inventory_batches SET quantity = '0' WHERE id = ?`, batch.ID); err != nil {
		t.Fatalf("deplete batch: %v", err)
	}
	if removed, err := env.svc.Generate(env.ctx); err != nil || removed != 0 {
		t.Fatalf("generate after deplete: removed=%d err=%v", removed, err)
	}
	if got := env.listed(t, false); len(got) != 0 {
		t.Fatalf("unread stale notification not removed, listed=%d", len(got))
	}

	batch2 := env.createStaleBatch(t, prod.ID, 40)
	if _, err := env.svc.Generate(env.ctx); err != nil {
		t.Fatalf("generate 2: %v", err)
	}
	if _, err := env.svc.MarkAllRead(env.ctx, env.companyID); err != nil {
		t.Fatalf("mark all read: %v", err)
	}
	if _, err := env.db.ExecContext(env.ctx,
		`UPDATE inventory_batches SET quantity = '0' WHERE id = ?`, batch2.ID); err != nil {
		t.Fatalf("deplete batch 2: %v", err)
	}
	if _, err := env.svc.Generate(env.ctx); err != nil {
		t.Fatalf("generate 3: %v", err)
	}
	got := env.listed(t, false)
	if len(got) != 1 {
		t.Fatalf("read notification must be kept as history, listed=%d", len(got))
	}
}

func TestRefreshClearanceFlags(t *testing.T) {
	env := newNtfEnv(t)
	prod := env.createProduct(t)
	// A fresh batch is created with is_clearance = FALSE; aging it
	// behind the repo's back simulates the stale-flag window the
	// refresh pass exists for.
	batch := env.createStaleBatch(t, prod.ID, 0)
	aged := valueobjects.NewDateFromTime(time.Now().UTC().AddDate(0, 0, -30))
	if _, err := env.db.ExecContext(env.ctx,
		`UPDATE inventory_batches SET arrival_date = ?, clearance_date = NULL, is_clearance = FALSE WHERE id = ?`,
		aged, batch.ID); err != nil {
		t.Fatalf("age batch: %v", err)
	}

	flagged, err := env.invSvc.RefreshClearanceFlags(env.ctx, env.companyID, time.Now().UTC())
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if flagged != 1 {
		t.Fatalf("flagged = %d, want 1", flagged)
	}
	stored, err := env.batches.GetByID(env.ctx, batch.ID)
	if err != nil {
		t.Fatalf("reload batch: %v", err)
	}
	if !stored.IsClearance(time.Now().UTC()) {
		t.Fatal("stored batch must keep is_clearance set after refresh")
	}
	if again, err := env.invSvc.RefreshClearanceFlags(env.ctx, env.companyID, time.Now().UTC()); err != nil || again != 0 {
		t.Fatalf("second refresh flagged %d (err=%v), want 0", again, err)
	}
}

func TestGenerateWithoutActiveCompany(t *testing.T) {
	env := newNtfEnv(t)
	prod := env.createProduct(t)
	env.createStaleBatch(t, prod.ID, 30)
	env.svc.SetActiveCompany(func() uuid.UUID { return uuid.Nil })

	if created, err := env.svc.Generate(env.ctx); err != nil || created != 0 {
		t.Fatalf("generate without company: created=%d err=%v, want 0/nil", created, err)
	}
}

func ntfMoney(s string) valueobjects.Money {
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		panic(err)
	}
	return m
}

func ntfQty(s string) valueobjects.Quantity {
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		panic(err)
	}
	return q
}

func ntfPen(s string) valueobjects.CurrencyCode {
	c, err := valueobjects.NewCurrencyCode(s)
	if err != nil {
		panic(err)
	}
	return c
}
