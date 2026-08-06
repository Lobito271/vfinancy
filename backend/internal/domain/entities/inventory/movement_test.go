package inventory

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

func TestMovementInbound(t *testing.T) {
	at := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	m, err := NewInventoryMovement(NewInventoryMovementOptions{
		CompanyID:    uuid.New(),
		ProductID:    uuid.New(),
		WarehouseID:  uuid.New(),
		Type:         enums.MovementTypePurchase,
		Quantity:     mustQuantityT(t, "12.5"),
		UnitCost:     mustMoneyT(t, "3.00"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		OccurredAt:   at,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !m.IsInbound() {
		t.Error("purchase should be inbound")
	}
	if m.SignedQuantity() != "+12.5000" {
		t.Errorf("signed: %s", m.SignedQuantity())
	}
}

func TestMovementOutbound(t *testing.T) {
	at := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	m, err := NewInventoryMovement(NewInventoryMovementOptions{
		CompanyID:    uuid.New(),
		ProductID:    uuid.New(),
		WarehouseID:  uuid.New(),
		Type:         enums.MovementTypeSale,
		Quantity:     quantityNeg(mustQuantityT(t, "3")),
		UnitCost:     mustMoneyT(t, "5.00"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		OccurredAt:   at,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if m.Quantity.IsPositive() {
		t.Error("sale quantity should be negative")
	}
	if !m.IsOutbound() {
		t.Error("sale should be outbound")
	}
	if m.SignedQuantity() != "-3.0000" {
		t.Errorf("signed: %s", m.SignedQuantity())
	}
}

func TestMovementSignMismatch(t *testing.T) {
	at := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	// Inbound with negative quantity should fail
	neg, _ := valueobjects.QuantityFromString("5")
	neg = valueobjects.Quantity{}.Sub(neg) // hacky; use a typed Quantity below
	_ = neg

	q, _ := valueobjects.QuantityFromString("5")
	q = quantityNeg(q)

	if _, err := NewInventoryMovement(NewInventoryMovementOptions{
		CompanyID:    uuid.New(),
		ProductID:    uuid.New(),
		WarehouseID:  uuid.New(),
		Type:         enums.MovementTypePurchase,
		Quantity:     q,
		UnitCost:     mustMoneyT(t, "1.00"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		OccurredAt:   at,
	}); err == nil {
		t.Error("inbound with negative quantity should fail")
	}
}

func TestMovementZeroQuantity(t *testing.T) {
	at := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	if _, err := NewInventoryMovement(NewInventoryMovementOptions{
		CompanyID:    uuid.New(),
		ProductID:    uuid.New(),
		WarehouseID:  uuid.New(),
		Type:         enums.MovementTypePurchase,
		Quantity:     mustQuantityT(t, "0"),
		UnitCost:     mustMoneyT(t, "1.00"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		OccurredAt:   at,
	}); err == nil {
		t.Error("zero quantity should fail")
	}
}

func TestMovementInvalidType(t *testing.T) {
	at := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	if _, err := NewInventoryMovement(NewInventoryMovementOptions{
		CompanyID:    uuid.New(),
		ProductID:    uuid.New(),
		WarehouseID:  uuid.New(),
		Type:         enums.InventoryMovementType("bogus"),
		Quantity:     mustQuantityT(t, "1"),
		UnitCost:     mustMoneyT(t, "1.00"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		OccurredAt:   at,
	}); err == nil {
		t.Error("invalid type should fail")
	}
}

func TestMovementTotalCost(t *testing.T) {
	at := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	m, _ := NewInventoryMovement(NewInventoryMovementOptions{
		CompanyID:    uuid.New(),
		ProductID:    uuid.New(),
		WarehouseID:  uuid.New(),
		Type:         enums.MovementTypePurchase,
		Quantity:     mustQuantityT(t, "5"),
		UnitCost:     mustMoneyT(t, "3.50"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		OccurredAt:   at,
	})
	if m.TotalCost().String() != "17.50" {
		t.Errorf("total: %s", m.TotalCost())
	}
}

func mustQuantityT(t *testing.T, s string) valueobjects.Quantity {
	t.Helper()
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		t.Fatalf("qty: %v", err)
	}
	return q
}

func mustMoneyT(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	return m
}

// quantityNeg returns a negative copy of q, used to construct
// negative-signed test inputs.
func quantityNeg(q valueobjects.Quantity) valueobjects.Quantity {
	d := q.Decimal().Neg()
	out, _ := valueobjects.QuantityFromDecimal(d)
	return out
}
