package masterdata

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

func newSupplier(t *testing.T) *Supplier {
	t.Helper()
	s, err := NewSupplier(time.Now(), NewSupplierOptions{
		CompanyID:       uuid.New(),
		Document:        mustDocument(t, enums.DocumentTypeRUC, "20123456789"),
		BusinessName:    valueobjects.MustFullName("Acme Supplies"),
		TaxID:            "20123456789",
		DefaultCurrency: valueobjects.MustCurrencyCode("USD"),
		PaymentTermDays: 45,
		Email:           mustEmail(t),
		Address:         mustAddress(t),
	})
	if err != nil {
		t.Fatalf("NewSupplier: %v", err)
	}
	return s
}

func TestSupplierNew(t *testing.T) {
	if _, err := NewSupplier(time.Now(), NewSupplierOptions{
		CompanyID:       uuid.Nil,
		Document:        mustDocument(t, enums.DocumentTypeRUC, "20123456789"),
		BusinessName:    valueobjects.MustFullName("X"),
		TaxID:            "1",
		DefaultCurrency: valueobjects.MustCurrencyCode("USD"),
		PaymentTermDays: 0,
		Email:           mustEmail(t),
		Address:         mustAddress(t),
	}); err == nil {
		t.Error("empty company id should fail")
	}
}

func TestSupplierRecordPurchaseAndPayment(t *testing.T) {
	s := newSupplier(t)
	debt := s.RecordPurchase(mustMoney(t, "500.00"))
	if debt.String() != "500.00" {
		t.Errorf("debt: %s", debt)
	}
	debt, _ = s.RecordPayment(mustMoney(t, "200"))
	if debt.String() != "300.00" {
		t.Errorf("debt after payment: %s", debt)
	}
}

func TestSupplierCanPlacePurchase(t *testing.T) {
	s := newSupplier(t)
	if err := s.CanPlacePurchase(mustMoney(t, "100")); err != nil {
		t.Errorf("active supplier: %v", err)
	}
	s.Deactivate()
	if err := s.CanPlacePurchase(mustMoney(t, "100")); err == nil {
		t.Error("inactive supplier should fail")
	}
}

func TestSupplierChangePaymentTerms(t *testing.T) {
	s := newSupplier(t)
	if err := s.ChangePaymentTerms(-1); err == nil {
		t.Error("negative terms should fail")
	}
	if err := s.ChangePaymentTerms(60); err != nil {
		t.Errorf("err: %v", err)
	}
}

func TestSupplierRecordPaymentInvalid(t *testing.T) {
	s := newSupplier(t)
	s.RecordPurchase(mustMoney(t, "100.00"))
	if _, err := s.RecordPayment(mustMoney(t, "0")); err == nil {
		t.Error("zero payment should fail")
	}
	if _, err := s.RecordPayment(mustMoney(t, "-1")); err == nil {
		t.Error("negative payment should fail")
	}
}
