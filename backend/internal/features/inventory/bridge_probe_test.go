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
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
	inventorypostgres "vfinancy/backend/internal/features/inventory/postgres"
)

var (
	probeProduct = uuid.MustParse("00000000-0000-0000-0000-000000000001")
)

const (
	probeLine  = "00000000-0000-0000-0000-0000000000e1"
	probeOrder = "00000000-0000-0000-0000-0000000000d2"
)

type probeEnv struct {
	ctx     context.Context
	db      *database.DB
	svc     *inventory.InventoryService
	batches inventory.InventoryBatchRepository
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

	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	lineID := uuid.MustParse(probeLine)
	orderID := uuid.MustParse(probeOrder)
	if _, err := db.ExecContext(ctx, `INSERT INTO products (id, sku, description) VALUES (?, 'PROBE-1', 'Probe Product')`, productID); err != nil {
		t.Fatalf("seed product: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO purchase_orders (id, number, order_date) VALUES (?, 'OC-PROBE-1', ?)`, orderID, time.Now().UTC()); err != nil {
		t.Fatalf("seed purchase order: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO purchase_order_items (id, purchase_order_id, description, quantity_ordered, unit_cost_usd) VALUES (?, ?, 'Probe Product', '100.0000', '1.00')`, lineID, orderID); err != nil {
		t.Fatalf("seed purchase line: %v", err)
	}

	batches := inventorypostgres.NewInventoryBatchRepository(db.DB)
	movements := inventorypostgres.NewInventoryMovementRepository(db.DB)
	svc := inventory.New(batches, movements, persistence.NewTxManager(db), log)
	return &probeEnv{ctx: ctx, db: db, svc: svc, batches: batches}
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

func TestReserveForSaleInsufficientStock(t *testing.T) {
	env := newProbeEnv(t)
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	lineID := uuid.MustParse(probeLine)
	today := valueobjects.NewDateFromTime(time.Now().UTC())

	batch, err := env.svc.ReceiveFromPurchase(env.ctx, inventory.ReceiveFromPurchaseInput{
		ProductID:      productID,
		PurchaseLineID: lineID,
		ArrivalDate:    today,
		Quantity:       qty("10"),
		UnitCost:       money("2.00"),
		ExchangeRate:   valueobjects.One(),
	})
	if err != nil {
		t.Fatalf("receive from purchase: %v", err)
	}

	_, err = env.svc.ReserveForSale(env.ctx, inventory.ReserveForSaleInput{
		ProductID: productID,
		Quantity:  qty("11"),
		SaleID:    uuid.New(),
	})
	if err == nil {
		t.Fatal("expected shortfall error")
	}
	if !derrors.IsCode(err, derrors.ErrInsufficientStock.Code()) {
		t.Fatalf("expected INSUFFICIENT_STOCK, got %v", err)
	}
	// The failed reserve must roll back completely.
	got, err := env.batches.GetByID(env.ctx, batch.ID)
	if err != nil {
		t.Fatalf("get batch: %v", err)
	}
	if !got.Quantity.Equals(qty("10.0000")) {
		t.Fatalf("batch quantity after failed reserve = %s, want 10.0000", got.Quantity)
	}
	if got.Status != enums.BatchStatusActive {
		t.Fatalf("batch status after failed reserve = %s, want active", got.Status)
	}
}

func TestAdjustRejectsNonPositiveTarget(t *testing.T) {
	env := newProbeEnv(t)
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	lineID := uuid.MustParse(probeLine)
	today := valueobjects.NewDateFromTime(time.Now().UTC())

	batch, err := env.svc.ReceiveFromPurchase(env.ctx, inventory.ReceiveFromPurchaseInput{
		ProductID:      productID,
		PurchaseLineID: lineID,
		ArrivalDate:    today,
		Quantity:       qty("10"),
		UnitCost:       money("2.00"),
		ExchangeRate:   valueobjects.One(),
	})
	if err != nil {
		t.Fatalf("receive from purchase: %v", err)
	}

	if err := env.svc.Adjust(env.ctx, inventory.AdjustInput{BatchID: batch.ID, NewQuantity: qty("0")}); err == nil {
		t.Fatal("expected error for zero target")
	} else if !derrors.IsCode(err, derrors.ErrNegativeQuantity.Code()) {
		t.Fatalf("expected NEGATIVE_QUANTITY, got %v", err)
	}
	if err := env.svc.Adjust(env.ctx, inventory.AdjustInput{BatchID: batch.ID, NewQuantity: qty("-5")}); err == nil {
		t.Fatal("expected error for negative target")
	}
}

func TestReceiveFromPurchaseRejectsFutureArrival(t *testing.T) {
	env := newProbeEnv(t)
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	lineID := uuid.MustParse(probeLine)
	tomorrow := valueobjects.NewDateFromTime(time.Now().UTC().AddDate(0, 0, 1))

	_, err := env.svc.ReceiveFromPurchase(env.ctx, inventory.ReceiveFromPurchaseInput{
		ProductID:      productID,
		PurchaseLineID: lineID,
		ArrivalDate:    tomorrow,
		Quantity:       qty("10"),
		UnitCost:       money("2.00"),
		ExchangeRate:   valueobjects.One(),
	})
	if err == nil {
		t.Fatal("expected error for future arrival date")
	}
	if !derrors.IsCode(err, derrors.ErrOutOfRange.Code()) {
		t.Fatalf("expected OUT_OF_RANGE, got %v", err)
	}
}
