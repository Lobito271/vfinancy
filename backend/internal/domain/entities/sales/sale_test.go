package sales

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

func mustMoney(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	return m
}

func mustPercent(t *testing.T, s string) valueobjects.Percentage {
	t.Helper()
	p, err := valueobjects.PercentageFromString(s)
	if err != nil {
		t.Fatalf("percent: %v", err)
	}
	return p
}

func mustQuantity(t *testing.T, s string) valueobjects.Quantity {
	t.Helper()
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		t.Fatalf("qty: %v", err)
	}
	return q
}

func mustExchangeRate(t *testing.T, s string) valueobjects.ExchangeRate {
	t.Helper()
	r, err := valueobjects.ExchangeRateFromString(s)
	if err != nil {
		t.Fatalf("rate: %v", err)
	}
	return r
}

func newSaleItem(t *testing.T, qty, price, cost string) *SaleItem {
	t.Helper()
	li, err := NewSaleItem(NewSaleItemOptions{
		ProductID:    uuid.New(),
		Quantity:     mustQuantity(t, qty),
		UnitPrice:    mustMoney(t, price),
		TaxRate:      mustPercent(t, "18"),
		TaxAmount:    mustMoney(t, "0"),
		CostSnapshot: mustMoney(t, cost),
	})
	if err != nil {
		t.Fatalf("NewSaleItem: %v", err)
	}
	return li
}

func TestSaleItemLineSubtotalAndTotal(t *testing.T) {
	li := newSaleItem(t, "3", "100", "60")
	if got := li.LineSubtotal().String(); got != "300.00" {
		t.Errorf("subtotal: %s", got)
	}
	if got := li.LineTotal().String(); got != "300.00" {
		t.Errorf("total (no tax/disc): %s", got)
	}
	if got := li.LineCost().String(); got != "180.00" {
		t.Errorf("cost: %s", got)
	}
	if got := li.LineProfit().String(); got != "120.00" {
		t.Errorf("profit: %s", got)
	}
}

func TestSaleItemLineTotalWithDiscountAndTax(t *testing.T) {
	qty := mustQuantity(t, "2")
	price := mustMoney(t, "100")
	discount := mustMoney(t, "20") // 2 * 100 - 20 = 180
	tax := mustMoney(t, "32.40")   // 18% of 180
	li, err := NewSaleItem(NewSaleItemOptions{
		ProductID:      uuid.New(),
		Quantity:       qty,
		UnitPrice:      price,
		DiscountAmount: discount,
		TaxRate:        mustPercent(t, "18"),
		TaxAmount:      tax,
		CostSnapshot:   mustMoney(t, "50"),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got := li.LineTotal().String(); got != "212.40" {
		t.Errorf("total: %s", got)
	}
	// profit = total - tax - cost = 212.40 - 32.40 - 100 = 80
	if got := li.LineProfit().String(); got != "80.00" {
		t.Errorf("profit: %s", got)
	}
}

func TestSaleAddItemAndTotals(t *testing.T) {
	s, err := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-1",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	if err != nil {
		t.Fatalf("NewSale: %v", err)
	}
	if !s.CalculateSubtotal().IsZero() {
		t.Error("subtotal should start at zero")
	}

	// 3 * 100 = 300
	if err := s.AddItem(newSaleItem(t, "3", "100", "60")); err != nil {
		t.Errorf("add: %v", err)
	}
	if got := s.CalculateSubtotal().String(); got != "300.00" {
		t.Errorf("subtotal: %s", got)
	}

	// 2 * 50 = 100 -- different product (newSaleItem creates a new ProductID)
	li2, _ := NewSaleItem(NewSaleItemOptions{
		ProductID:    uuid.New(),
		Quantity:     mustQuantity(t, "2"),
		UnitPrice:    mustMoney(t, "50"),
		CostSnapshot: mustMoney(t, "30"),
	})
	if err := s.AddItem(li2); err != nil {
		t.Errorf("add: %v", err)
	}
	if got := s.CalculateSubtotal().String(); got != "400.00" {
		t.Errorf("subtotal: %s", got)
	}

	// Cost total = 3*60 + 2*30 = 240
	if got := s.CostTotal.String(); got != "240.00" {
		t.Errorf("cost total: %s", got)
	}

	// Profit = total - tax - cost = 400 - 0 - 240 = 160
	if got := s.CalculateProfit().String(); got != "160.00" {
		t.Errorf("profit: %s", got)
	}
}

func TestSaleDuplicateItemRejected(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-2",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	pid := uuid.New()
	mk := func() *SaleItem {
		li, _ := NewSaleItem(NewSaleItemOptions{
			ProductID:    pid,
			Quantity:     mustQuantity(t, "1"),
			UnitPrice:    mustMoney(t, "10"),
			CostSnapshot: mustMoney(t, "5"),
		})
		return li
	}
	if err := s.AddItem(mk()); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := s.AddItem(mk()); err == nil {
		t.Error("duplicate product should fail")
	}
}

func TestSaleRemoveItem(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-3",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	li := newSaleItem(t, "3", "100", "60")
	if err := s.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := s.RemoveItem(li.ID); err != nil {
		t.Errorf("remove: %v", err)
	}
	if len(s.Items) != 0 {
		t.Errorf("items after remove: %d", len(s.Items))
	}
	if !s.CalculateSubtotal().IsZero() {
		t.Error("subtotal should be zero after remove")
	}
}

func TestSalePayments(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-4",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	if err := s.AddItem(newSaleItem(t, "3", "100", "60")); err != nil {
		t.Fatalf("add: %v", err)
	}
	// Total = 300, balance = 300
	if got := s.Balance().String(); got != "300.00" {
		t.Errorf("balance: %s", got)
	}

	// Apply 100 -> status = partial
	if _, err := s.ApplyPayment(mustMoney(t, "100")); err != nil {
		t.Errorf("apply 100: %v", err)
	}
	if s.Status != enums.SaleStatusPartial {
		t.Errorf("status: %s", s.Status)
	}
	if got := s.Balance().String(); got != "200.00" {
		t.Errorf("balance: %s", got)
	}

	// Overpayment rejected
	if _, err := s.ApplyPayment(mustMoney(t, "999")); err == nil {
		t.Error("overpayment should fail")
	}

	// Apply 200 -> status = paid
	if _, err := s.ApplyPayment(mustMoney(t, "200")); err != nil {
		t.Errorf("apply 200: %v", err)
	}
	if !s.IsPaid() {
		t.Error("should be paid")
	}
	if got := s.Balance().String(); got != "0.00" {
		t.Errorf("balance after paid: %s", got)
	}

	// Further payment rejected
	if _, err := s.ApplyPayment(mustMoney(t, "1")); err == nil {
		t.Error("further payment should fail on paid sale")
	}
}

func TestSaleMarkAsPaid(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-5",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	if err := s.AddItem(newSaleItem(t, "2", "50", "30")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := s.MarkAsPaid(); err != nil {
		t.Errorf("mark paid: %v", err)
	}
	if !s.IsPaid() {
		t.Error("should be paid")
	}
}

func TestSaleCancel(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-6",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	if err := s.AddItem(newSaleItem(t, "1", "10", "5")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := s.Cancel("test"); err != nil {
		t.Errorf("cancel: %v", err)
	}
	if !s.IsCancelled() {
		t.Error("should be cancelled")
	}
	if err := s.Cancel("again"); err == nil {
		t.Error("double cancel should fail")
	}
}

func TestSaleCannotAddToPaidSale(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-7",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	if err := s.AddItem(newSaleItem(t, "1", "10", "5")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := s.MarkAsPaid(); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := s.AddItem(newSaleItem(t, "1", "10", "5")); err == nil {
		t.Error("cannot add to paid sale")
	}
}

func TestSaleEmptyDocumentRejection(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-8",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	if err := s.MarkAsPaid(); err == nil {
		t.Error("mark paid on empty sale should fail")
	}
}

func TestSalePartiallyPaidOverride(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-9",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	if err := s.AddItem(newSaleItem(t, "1", "100", "60")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := s.MarkAsPartiallyPaid(mustMoney(t, "40")); err != nil {
		t.Errorf("err: %v", err)
	}
	if !s.IsPartiallyPaid() {
		t.Error("should be partial")
	}
	// invalid: amount >= total
	if err := s.MarkAsPartiallyPaid(mustMoney(t, "100")); err == nil {
		t.Error("amount >= total should fail")
	}
}

func TestSaleProfitCalculation(t *testing.T) {
	s, _ := NewSale(time.Now(), NewSaleOptions{
		CompanyID:    uuid.New(),
		Number:       "F001-10",
		CustomerID:   uuid.New(),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchangeRate(t, "1"),
	})
	// line 1: 10 * 100, cost 60, tax 0 -> total 1000, profit 1000 - 0 - 600 = 400
	li, _ := NewSaleItem(NewSaleItemOptions{
		ProductID:    uuid.New(),
		Quantity:     mustQuantity(t, "10"),
		UnitPrice:    mustMoney(t, "100"),
		TaxRate:      mustPercent(t, "0"),
		CostSnapshot: mustMoney(t, "60"),
	})
	if err := s.AddItem(li); err != nil {
		t.Fatalf("add: %v", err)
	}
	if got := s.CalculateProfit().String(); got != "400.00" {
		t.Errorf("profit: %s", got)
	}
}
