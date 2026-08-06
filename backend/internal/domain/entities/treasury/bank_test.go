package treasury

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/valueobjects"
)

func mustMoney(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, _ := valueobjects.MoneyFromString(s)
	return m
}

func TestBankAccountNewAndApplyDelta(t *testing.T) {
	a, err := NewBankAccount(time.Now(), NewBankAccountOptions{
		CompanyID:     uuid.New(),
		BranchID:      nil,
		BankName:      "BCP",
		AccountNumber: "193-1234567-0-99",
		AccountType:   "checking",
		CurrencyCode:  valueobjects.MustCurrencyCode("PEN"),
		GLAccountID:   uuid.New(),
		IsDefault:     true,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !a.CurrentBalance.IsZero() {
		t.Errorf("initial balance: %s", a.CurrentBalance)
	}
	a.ApplyDelta(mustMoney(t, "1500.50"))
	if a.CurrentBalance.String() != "1500.50" {
		t.Errorf("balance: %s", a.CurrentBalance)
	}
	a.ApplyDelta(mustMoney(t, "-200"))
	if a.CurrentBalance.String() != "1300.50" {
		t.Errorf("balance: %s", a.CurrentBalance)
	}
}

func TestBankAccountNewValidations(t *testing.T) {
	if _, err := NewBankAccount(time.Now(), NewBankAccountOptions{
		CompanyID:     uuid.Nil,
		BankName:      "BCP",
		AccountNumber: "123",
		CurrencyCode:  valueobjects.MustCurrencyCode("PEN"),
		GLAccountID:   uuid.New(),
	}); err == nil {
		t.Error("empty company should fail")
	}
	if _, err := NewBankAccount(time.Now(), NewBankAccountOptions{
		CompanyID:     uuid.New(),
		BankName:      "",
		AccountNumber: "123",
		CurrencyCode:  valueobjects.MustCurrencyCode("PEN"),
		GLAccountID:   uuid.New(),
	}); err == nil {
		t.Error("empty bank name should fail")
	}
	if _, err := NewBankAccount(time.Now(), NewBankAccountOptions{
		CompanyID:     uuid.New(),
		BankName:      "BCP",
		AccountNumber: "",
		CurrencyCode:  valueobjects.MustCurrencyCode("PEN"),
		GLAccountID:   uuid.New(),
	}); err == nil {
		t.Error("empty account number should fail")
	}
}

func TestCreditCardChargeAndPay(t *testing.T) {
	c, err := NewCreditCard(time.Now(), NewCreditCardOptions{
		CompanyID:       uuid.New(),
		Issuer:          "Visa",
		LastFour:        "1234",
		CardHolder:      "Maria Garcia",
		ExpirationMonth: 12,
		ExpirationYear:  2030,
		CreditLimit:     mustMoney(t, "5000.00"),
		CutOffDay:       15,
		PaymentDueDay:   5,
		CurrencyCode:    valueobjects.MustCurrencyCode("PEN"),
		GLAccountID:     uuid.New(),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !c.AvailableCredit().Equals(mustMoney(t, "5000.00")) {
		t.Errorf("avail: %s", c.AvailableCredit())
	}
	if err := c.Charge(mustMoney(t, "1000")); err != nil {
		t.Errorf("charge: %v", err)
	}
	if c.AvailableCredit().String() != "4000.00" {
		t.Errorf("avail after charge: %s", c.AvailableCredit())
	}
	// Over-limit rejected
	if err := c.Charge(mustMoney(t, "5000")); err == nil {
		t.Error("over-limit should fail")
	}
	if err := c.Pay(mustMoney(t, "500")); err != nil {
		t.Errorf("pay: %v", err)
	}
	if c.CurrentBalance.String() != "500.00" {
		t.Errorf("balance: %s", c.CurrentBalance)
	}
}

func TestCreditCardValidations(t *testing.T) {
	if _, err := NewCreditCard(time.Now(), NewCreditCardOptions{
		CompanyID:       uuid.New(),
		Issuer:          "Visa",
		LastFour:        "12", // 2 digits
		ExpirationMonth: 12,
		ExpirationYear:  2030,
		CreditLimit:     mustMoney(t, "1000"),
		CutOffDay:       15,
		PaymentDueDay:   5,
		CurrencyCode:    valueobjects.MustCurrencyCode("PEN"),
		GLAccountID:     uuid.New(),
	}); err == nil {
		t.Error("4-digit last-four required")
	}
	if _, err := NewCreditCard(time.Now(), NewCreditCardOptions{
		CompanyID:       uuid.New(),
		Issuer:          "Visa",
		LastFour:        "1234",
		ExpirationMonth: 13,
		ExpirationYear:  2030,
		CreditLimit:     mustMoney(t, "1000"),
		CutOffDay:       15,
		PaymentDueDay:   5,
		CurrencyCode:    valueobjects.MustCurrencyCode("PEN"),
		GLAccountID:     uuid.New(),
	}); err == nil {
		t.Error("month 13 should fail")
	}
}
