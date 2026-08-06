package masterdata

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

func newProduct(t *testing.T, cost, price string) *Product {
	t.Helper()
	p, err := NewProduct(time.Now(), NewProductOptions{
		CompanyID:    uuid.New(),
		SKU:          valueobjects.MustSKU("ABC001"),
		Barcode:      valueobjects.Barcode{}, // optional
		Description:  "Test product",
		UnitID:       uuid.New(),
		TaxID:        uuid.New(),
		Cost:         mustMoney(t, cost),
		SalePrice:    mustMoney(t, price),
		SaleCurrency: valueobjects.MustCurrencyCode("PEN"),
		MinStock:     mustQuantity(t, "0"),
		MaxStock:     mustQuantity(t, "100"),
	})
	if err != nil {
		t.Fatalf("NewProduct: %v", err)
	}
	return p
}

func TestProductNewValidations(t *testing.T) {
	if _, err := NewProduct(time.Now(), NewProductOptions{
		CompanyID:    uuid.New(),
		SKU:          valueobjects.MustSKU("X1"),
		Description:  "x",
		UnitID:       uuid.New(),
		TaxID:        uuid.New(),
		Cost:         mustMoney(t, "1"),
		SalePrice:    mustMoney(t, "1"),
		SaleCurrency: valueobjects.MustCurrencyCode("PEN"),
		MinStock:     mustQuantity(t, "10"),
		MaxStock:     mustQuantity(t, "5"), // max < min
	}); err == nil {
		t.Error("max < min should fail")
	}

	if _, err := NewProduct(time.Now(), NewProductOptions{
		CompanyID:    uuid.New(),
		SKU:          valueobjects.MustSKU("X1"),
		Description:  "",
		UnitID:       uuid.New(),
		TaxID:        uuid.New(),
		Cost:         mustMoney(t, "1"),
		SalePrice:    mustMoney(t, "1"),
		SaleCurrency: valueobjects.MustCurrencyCode("PEN"),
		MinStock:     mustQuantity(t, "0"),
		MaxStock:     mustQuantity(t, "10"),
	}); err == nil {
		t.Error("empty description should fail")
	}
}

func TestProductPriceAndCost(t *testing.T) {
	p := newProduct(t, "5.00", "10.00")
	if err := p.ChangeSalePrice(mustMoney(t, "-1")); err == nil {
		t.Error("negative price should fail")
	}
	if err := p.ChangeSalePrice(mustMoney(t, "12.50")); err != nil {
		t.Errorf("err: %v", err)
	}
	if p.SalePrice.String() != "12.50" {
		t.Errorf("price: %s", p.SalePrice)
	}
	if err := p.ChangeCost(mustMoney(t, "-1")); err == nil {
		t.Error("negative cost should fail")
	}
}

func TestProductMargin(t *testing.T) {
	p := newProduct(t, "5.00", "10.00")
	if p.CalculateMargin().String() != "5.00" {
		t.Errorf("margin: %s", p.CalculateMargin())
	}
	// 5/10 = 50%
	if got := p.MarginPercent(); got != 50 {
		t.Errorf("margin%%: %v", got)
	}

	// zero price => 0%
	z := newProduct(t, "0", "0")
	if got := z.MarginPercent(); got != 0 {
		t.Errorf("zero margin: %v", got)
	}
}

func TestProductStockLimits(t *testing.T) {
	p := newProduct(t, "1", "2")
	if err := p.ChangeStockLimits(mustQuantity(t, "100"), mustQuantity(t, "10")); err == nil {
		t.Error("max < min should fail")
	}
	if err := p.ChangeStockLimits(mustQuantity(t, "5"), mustQuantity(t, "20")); err != nil {
		t.Errorf("err: %v", err)
	}
	if p.MinStock.String() != "5.0000" || p.MaxStock.String() != "20.0000" {
		t.Errorf("limits: %s - %s", p.MinStock, p.MaxStock)
	}
}

func TestProductActivateDeactivate(t *testing.T) {
	p := newProduct(t, "1", "2")
	p.Deactivate()
	if p.IsActive {
		t.Error("should be inactive")
	}
	p.Activate()
	if !p.IsActive {
		t.Error("should be active")
	}
}

func mustQuantity(t *testing.T, s string) valueobjects.Quantity {
	t.Helper()
	q, err := valueobjects.QuantityFromString(s)
	if err != nil {
		t.Fatalf("mustQuantity(%q): %v", s, err)
	}
	return q
}

var _ = enums.TaxCategoryTaxed
