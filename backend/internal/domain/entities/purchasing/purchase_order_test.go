package purchasing

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

func mustMoney(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, _ := valueobjects.MoneyFromString(s)
	return m
}

func mustQuantity(t *testing.T, s string) valueobjects.Quantity {
	t.Helper()
	q, _ := valueobjects.QuantityFromString(s)
	return q
}

func mustExchange(t *testing.T) valueobjects.ExchangeRate {
	t.Helper()
	r, _ := valueobjects.ExchangeRateFromString("1")
	return r
}

func newPO(t *testing.T) *PurchaseOrder {
	t.Helper()
	o, err := NewPurchaseOrder(time.Now(), NewPurchaseOrderOptions{
		CompanyID:    uuid.New(),
		Number:       "C001-1",
		SupplierID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchange(t),
		OrderDate:    time.Now(),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return o
}

func TestPurchaseOrderNewAndTotals(t *testing.T) {
	o := newPO(t)
	if o.Status != enums.PurchaseStatusPending {
		t.Errorf("status: %s", o.Status)
	}

	// 3 * 100 = 300
	li, _ := NewPurchaseOrderItem(NewPurchaseOrderItemOptions{
		ProductID: uuid.New(),
		Quantity:  mustQuantity(t, "3"),
		UnitPrice: mustMoney(t, "100"),
	})
	if err := o.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	if got := o.CalculateTotal().String(); got != "300.00" {
		t.Errorf("total: %s", got)
	}
}

func TestPurchaseOrderApproveRequiresItems(t *testing.T) {
	o := newPO(t)
	if err := o.Approve(); err == nil {
		t.Error("empty order should not be approvable")
	}
}

func TestPurchaseOrderApproveAndMarkAsReceived(t *testing.T) {
	o := newPO(t)
	li, _ := NewPurchaseOrderItem(NewPurchaseOrderItemOptions{
		ProductID: uuid.New(),
		Quantity:  mustQuantity(t, "5"),
		UnitPrice: mustMoney(t, "10"),
	})
	if err := o.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := o.Approve(); err != nil {
		t.Errorf("approve: %v", err)
	}
	if !o.IsReceived() {
		t.Error("should be received")
	}
}

func TestPurchaseOrderMarkAsReceivedTwice(t *testing.T) {
	o := newPO(t)
	li, _ := NewPurchaseOrderItem(NewPurchaseOrderItemOptions{
		ProductID: uuid.New(),
		Quantity:  mustQuantity(t, "1"),
		UnitPrice: mustMoney(t, "1"),
	})
	if err := o.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := o.Approve(); err != nil {
		t.Fatalf("approve: %v", err)
	}
	at := time.Now()
	if err := o.MarkAsReceived(at); err != nil {
		t.Errorf("first mark: %v", err)
	}
	if err := o.MarkAsReceived(at); err == nil {
		t.Error("double mark should fail")
	}
}

func TestPurchaseOrderPayment(t *testing.T) {
	o := newPO(t)
	li, _ := NewPurchaseOrderItem(NewPurchaseOrderItemOptions{
		ProductID: uuid.New(),
		Quantity:  mustQuantity(t, "2"),
		UnitPrice: mustMoney(t, "100"),
	})
	if err := o.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	// total = 200
	if got := o.Balance().String(); got != "200.00" {
		t.Errorf("balance: %s", got)
	}
	if _, err := o.ApplyPayment(mustMoney(t, "80")); err != nil {
		t.Errorf("apply: %v", err)
	}
	if o.Status != enums.PurchaseStatusPending {
		t.Errorf("status after partial: %s", o.Status)
	}
	if _, err := o.ApplyPayment(mustMoney(t, "120")); err != nil {
		t.Errorf("apply: %v", err)
	}
	if !o.IsPaid() {
		t.Error("should be paid")
	}
}

func TestPurchaseOrderOverpaymentRejected(t *testing.T) {
	o := newPO(t)
	li, _ := NewPurchaseOrderItem(NewPurchaseOrderItemOptions{
		ProductID: uuid.New(),
		Quantity:  mustQuantity(t, "1"),
		UnitPrice: mustMoney(t, "100"),
	})
	if err := o.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := o.ApplyPayment(mustMoney(t, "150")); err == nil {
		t.Error("overpayment should fail")
	}
}

func TestPurchaseOrderCancel(t *testing.T) {
	o := newPO(t)
	if err := o.Cancel("test"); err != nil {
		t.Errorf("cancel: %v", err)
	}
	if !o.IsCancelled() {
		t.Error("should be cancelled")
	}
	if err := o.Cancel("again"); err == nil {
		t.Error("double cancel should fail")
	}
}

func TestPurchaseOrderReconcile(t *testing.T) {
	o := newPO(t)
	li, _ := NewPurchaseOrderItem(NewPurchaseOrderItemOptions{
		ProductID: uuid.New(),
		Quantity:  mustQuantity(t, "1"),
		UnitPrice: mustMoney(t, "100"),
	})
	if err := o.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := o.MarkAsPaid(); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if err := o.Reconcile(); err != nil {
		t.Errorf("reconcile: %v", err)
	}
	if !o.IsReconciled() {
		t.Error("should be reconciled")
	}
	if err := o.Reconcile(); err == nil {
		t.Error("double reconcile should fail")
	}
}

func TestSupplierPaymentAllocation(t *testing.T) {
	sp, err := NewSupplierPayment(time.Now(), NewSupplierPaymentOptions{
		CompanyID:   uuid.New(),
		SupplierID:  uuid.New(),
		Number:      "SP-1",
		PaymentDate: time.Now(),
		Amount:      mustMoney(t, "500"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchange(t),
		Method:      enums.PaymentMethodBankTransfer,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	p1 := uuid.New()
	p2 := uuid.New()
	if err := sp.ApplyToPurchase(p1, mustMoney(t, "200")); err != nil {
		t.Errorf("apply 200: %v", err)
	}
	if err := sp.ApplyToPurchase(p2, mustMoney(t, "300")); err != nil {
		t.Errorf("apply 300: %v", err)
	}
	if sp.AllocatedAmount().String() != "500.00" {
		t.Errorf("allocated: %s", sp.AllocatedAmount())
	}
	// over-allocation rejected
	if err := sp.ApplyToPurchase(p1, mustMoney(t, "1")); err == nil {
		t.Error("over-allocation should fail")
	}
}
