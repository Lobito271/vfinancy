package inventory

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/valueobjects"
)

func mustDate(t *testing.T, y int, m time.Month, d int) valueobjects.Date {
	t.Helper()
	date, err := valueobjects.NewDate(y, m, d)
	if err != nil {
		t.Fatalf("date: %v", err)
	}
	return date
}

func mustMoney(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	return m
}

func mustQuantity(t *testing.T, s string) valueobjects.Quantity {
	t.Helper()
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		t.Fatalf("qty: %v", err)
	}
	return q
}

func newBatch(t *testing.T, arrival valueobjects.Date, qty string) *InventoryBatch {
	t.Helper()
	b, err := NewInventoryBatch(time.Now(), NewInventoryBatchOptions{
		CompanyID:       uuid.New(),
		ProductID:       uuid.New(),
		WarehouseID:     uuid.New(),
		ArrivalDate:     arrival,
		InitialQuantity: mustQuantity(t, qty),
		UnitCost:        mustMoney(t, "10.00"),
		CurrencyCode:    valueobjects.MustCurrencyCode("PEN"),
	})
	if err != nil {
		t.Fatalf("NewInventoryBatch: %v", err)
	}
	return b
}

func TestBatchNewValidation(t *testing.T) {
	arrival := mustDate(t, 2026, 1, 15)
	if _, err := NewInventoryBatch(time.Now(), NewInventoryBatchOptions{
		CompanyID:       uuid.New(),
		ProductID:       uuid.New(),
		WarehouseID:     uuid.New(),
		ArrivalDate:     arrival,
		InitialQuantity: mustQuantity(t, "0"),
		UnitCost:        mustMoney(t, "0"),
		CurrencyCode:    valueobjects.MustCurrencyCode("PEN"),
	}); err == nil {
		t.Error("zero initial quantity should fail")
	}
}

func TestBatchMaximumSaleDate(t *testing.T) {
	// 2026-01-15 + 25 days = 2026-02-09
	arrival := mustDate(t, 2026, 1, 15)
	b := newBatch(t, arrival, "100")
	expected := mustDate(t, 2026, 2, 9)
	if !b.MaximumSaleDate().Equal(expected) {
		t.Errorf("max sale: got %s, want %s", b.MaximumSaleDate(), expected)
	}
}

func TestBatchIsClearance(t *testing.T) {
	arrival := mustDate(t, 2026, 1, 1)
	b := newBatch(t, arrival, "100")

	// Day 0: not clearance
	day0 := mustDate(t, 2026, 1, 1)
	if b.IsClearance(day0) {
		t.Error("day 0 should not be clearance")
	}

	// Day 25 (the boundary): IS clearance (>=)
	day25 := mustDate(t, 2026, 1, 26)
	if !b.IsClearance(day25) {
		t.Error("day 25 should be clearance")
	}

	// Day 30: clearance
	day30 := mustDate(t, 2026, 1, 31)
	if !b.IsClearance(day30) {
		t.Error("day 30 should be clearance")
	}

	// Zero quantity: NOT clearance (depleted, not clearance)
	b.Consume(mustQuantity(t, "100"))
	if b.IsClearance(day30) {
		t.Error("depleted batch should not be clearance")
	}
}

func TestBatchDaysInStockAndUntilClearance(t *testing.T) {
	arrival := mustDate(t, 2026, 1, 1)
	b := newBatch(t, arrival, "100")
	today := mustDate(t, 2026, 1, 11)
	if got := b.DaysInStock(today); got != 10 {
		t.Errorf("days in stock: %d", got)
	}
	// days until clearance = 25 - 10 = 15
	if got := b.DaysUntilClearance(today); got != 15 {
		t.Errorf("days until clearance: %d", got)
	}
}

func TestBatchConsumeAndReceive(t *testing.T) {
	b := newBatch(t, mustDate(t, 2026, 1, 1), "100")

	if err := b.Consume(mustQuantity(t, "30")); err != nil {
		t.Errorf("consume: %v", err)
	}
	if b.CurrentQuantity.String() != "70.0000" {
		t.Errorf("after consume: %s", b.CurrentQuantity)
	}

	// Cannot consume more than available
	if err := b.Consume(mustQuantity(t, "100")); err == nil {
		t.Error("consuming more than available should fail")
	}

	// Negative consume fails
	if err := b.Consume(mustQuantity(t, "0")); err == nil {
		t.Error("zero consume should fail")
	}

	// Receive adds back
	if _, err := b.Receive(mustQuantity(t, "20")); err != nil {
		t.Errorf("receive: %v", err)
	}
	if b.CurrentQuantity.String() != "90.0000" {
		t.Errorf("after receive: %s", b.CurrentQuantity)
	}

	// Consume all -> depleted
	if err := b.Consume(mustQuantity(t, "90")); err != nil {
		t.Errorf("consume all: %v", err)
	}
	if b.Status != "depleted" {
		t.Errorf("status after depleted: %s", b.Status)
	}

	// Receive again re-activates
	if _, err := b.Receive(mustQuantity(t, "10")); err != nil {
		t.Errorf("receive after depleted: %v", err)
	}
	if b.Status != "active" {
		t.Errorf("status after receive: %s", b.Status)
	}
}

func TestBatchWriteOff(t *testing.T) {
	b := newBatch(t, mustDate(t, 2026, 1, 1), "100")
	b.WriteOff()
	if !b.CurrentQuantity.IsZero() {
		t.Errorf("writeoff quantity: %s", b.CurrentQuantity)
	}
	if b.Status != "written_off" {
		t.Errorf("status: %s", b.Status)
	}
}

func TestBatchNeedsClearanceSoon(t *testing.T) {
	arrival := mustDate(t, 2026, 1, 1)
	b := newBatch(t, arrival, "100")

	// Day 20: 5 days left -> within 3? No, 5 > 3
	day20 := mustDate(t, 2026, 1, 21)
	if b.NeedsClearanceSoon(day20) {
		t.Error("5 days left should not be 'soon'")
	}

	// Day 23: 2 days left -> within 3
	day23 := mustDate(t, 2026, 1, 24)
	if !b.NeedsClearanceSoon(day23) {
		t.Error("2 days left should be 'soon'")
	}

	// Day 27: already clearance, not "soon"
	day27 := mustDate(t, 2026, 1, 28)
	if b.NeedsClearanceSoon(day27) {
		t.Error("already clearance should not be 'soon'")
	}

	// Depleted: not soon
	b.Consume(mustQuantity(t, "100"))
	if b.NeedsClearanceSoon(day23) {
		t.Error("depleted should not be 'soon'")
	}
}
