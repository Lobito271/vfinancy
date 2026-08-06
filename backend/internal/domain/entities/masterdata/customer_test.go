package masterdata

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

func mustMoney(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("mustMoney(%q): %v", s, err)
	}
	return m
}

func mustPercent(t *testing.T, s string) valueobjects.Percentage {
	t.Helper()
	p, err := valueobjects.PercentageFromString(s)
	if err != nil {
		t.Fatalf("mustPercent(%q): %v", s, err)
	}
	return p
}

func mustDocument(t *testing.T, dt enums.DocumentType, num string) valueobjects.DocumentNumber {
	t.Helper()
	d, err := valueobjects.NewDocumentNumber(dt, num)
	if err != nil {
		t.Fatalf("mustDocument: %v", err)
	}
	return d
}

func newCustomer(t *testing.T, creditLimit string) *Customer {
	t.Helper()
	email, _ := valueobjects.NewEmail("buyer@example.com")
	addr, _ := valueobjects.NewAddress("Calle 1")
	c, err := NewCustomer(time.Now(), NewCustomerOptions{
		CompanyID:       uuid.New(),
		Document:        mustDocument(t, enums.DocumentTypeRUC, "20123456789"),
		BusinessName:    valueobjects.MustFullName("Distribuidora García"),
		TaxCategory:     enums.TaxCategoryTaxed,
		CreditLimit:     mustMoney(t, creditLimit),
		PaymentTermDays: 30,
		Email:           email,
		Address:         addr,
	})
	if err != nil {
		t.Fatalf("NewCustomer: %v", err)
	}
	return c
}

func TestCustomerNew(t *testing.T) {
	c := newCustomer(t, "5000.00")
	if c.Status != enums.CustomerStatusActive {
		t.Errorf("status: %s", c.Status)
	}
	if !c.CurrentDebt.IsZero() {
		t.Errorf("debt should start at zero, got %s", c.CurrentDebt)
	}

	if _, err := NewCustomer(time.Now(), NewCustomerOptions{
		CompanyID:    uuid.Nil,
		Document:     mustDocument(t, enums.DocumentTypeRUC, "20123456789"),
		BusinessName: valueobjects.MustFullName("X"),
		TaxCategory:  enums.TaxCategoryTaxed,
		CreditLimit:  mustMoney(t, "0"),
		Email:        mustEmail(t),
		Address:      mustAddress(t),
	}); err == nil {
		t.Error("empty company id should fail")
	}
}

func TestCustomerActivateDeactivateBlock(t *testing.T) {
	c := newCustomer(t, "1000.00")
	c.Deactivate()
	if c.Status != enums.CustomerStatusInactive {
		t.Errorf("status: %s", c.Status)
	}
	c.Activate()
	if c.Status != enums.CustomerStatusActive {
		t.Errorf("status: %s", c.Status)
	}
	c.Block("credit overdue")
	if c.Status != enums.CustomerStatusBlocked {
		t.Errorf("status: %s", c.Status)
	}
	if c.BlockedReason != "credit overdue" {
		t.Errorf("reason: %s", c.BlockedReason)
	}
	c.Unblock()
	if c.Status != enums.CustomerStatusActive {
		t.Errorf("status: %s", c.Status)
	}
}

func TestCustomerRecordSaleAndPayment(t *testing.T) {
	c := newCustomer(t, "1000.00")
	debt := c.RecordSale(mustMoney(t, "300.00"))
	if debt.String() != "300.00" {
		t.Errorf("debt after sale: %s", debt)
	}
	debt, _ = c.RecordPayment(mustMoney(t, "100.00"))
	if debt.String() != "200.00" {
		t.Errorf("debt after payment: %s", debt)
	}
	// overpayment clamps to zero
	debt, _ = c.RecordPayment(mustMoney(t, "999.00"))
	if debt.String() != "0.00" {
		t.Errorf("debt after overpayment: %s", debt)
	}
}

func TestCustomerAvailableCredit(t *testing.T) {
	c := newCustomer(t, "1000.00")
	if c.AvailableCredit().String() != "1000.00" {
		t.Errorf("avail: %s", c.AvailableCredit())
	}
	c.RecordSale(mustMoney(t, "300.00"))
	if c.AvailableCredit().String() != "700.00" {
		t.Errorf("avail after sale: %s", c.AvailableCredit())
	}
	c.RecordSale(mustMoney(t, "800.00")) // now 1100, over limit
	if !c.IsOverLimit() {
		t.Error("should be over limit")
	}
	if c.AvailableCredit().String() != "0.00" {
		t.Errorf("avail over limit: %s", c.AvailableCredit())
	}
}

func TestCustomerCanPlaceSale(t *testing.T) {
	c := newCustomer(t, "1000.00")
	if err := c.CanPlaceSale(mustMoney(t, "500.00")); err != nil {
		t.Errorf("should allow: %v", err)
	}
	if err := c.CanPlaceSale(mustMoney(t, "1500.00")); err == nil {
		t.Error("should reject over-limit sale")
	}
	c.Block("test")
	if err := c.CanPlaceSale(mustMoney(t, "100.00")); err == nil {
		t.Error("blocked customer should not be able to place sale")
	}
}

func TestCustomerUpdateCreditLimit(t *testing.T) {
	c := newCustomer(t, "1000.00")
	if err := c.UpdateCreditLimit(mustMoney(t, "-1")); err == nil {
		t.Error("negative limit should fail")
	}
	if err := c.UpdateCreditLimit(mustMoney(t, "2000")); err != nil {
		t.Errorf("valid update failed: %v", err)
	}
	if c.CreditLimit.String() != "2000.00" {
		t.Errorf("limit: %s", c.CreditLimit)
	}
}

func TestCustomerRecordPaymentInvalid(t *testing.T) {
	c := newCustomer(t, "1000.00")
	if _, err := c.RecordPayment(mustMoney(t, "0")); err == nil {
		t.Error("zero payment should fail")
	}
	if _, err := c.RecordPayment(mustMoney(t, "-1")); err == nil {
		t.Error("negative payment should fail")
	}
}

func TestCustomerErrors(t *testing.T) {
	if !errors.IsCode(errors.ErrCustomerInactive, "CUSTOMER_INACTIVE") {
		t.Error("error code mismatch")
	}
	if errors.IsCode(errors.ErrCustomerInactive, "OTHER") {
		t.Error("unexpected match")
	}
}

func mustEmail(t *testing.T) valueobjects.Email {
	t.Helper()
	e, _ := valueobjects.NewEmail("user@example.com")
	return e
}

func mustAddress(t *testing.T) valueobjects.Address {
	t.Helper()
	a, _ := valueobjects.NewAddress("Calle 1")
	return a
}
