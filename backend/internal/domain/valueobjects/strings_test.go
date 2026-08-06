package valueobjects

import (
	"testing"

	"vfinancy/backend/internal/domain/enums"
)

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"user@example.com", false},
		{"USER@EXAMPLE.COM", false}, // lowercased
		{"  user@example.com  ", false},
		{"first.last+tag@sub.example.co", false},
		{"", true},
		{"no-at-sign", true},
		{"user@", true},
		{"@example.com", true},
		{"user@example", true},
		{"user@.com", true},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, err := NewEmail(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestPhoneValidation(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"987654321", false},
		{"987-654-321", false},  // dashes stripped
		{"  987654321  ", false}, // whitespace trimmed
		{"", true},               // NewPhone rejects empty (use OptionalPhone for the optional case)
		{"12345678", true},       // must start with 9
		{"98765432", true},       // 8 digits
		{"9876543210", true},     // 10 digits
		{"abcdefghi", true},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, err := NewPhone(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestSKUNormalization(t *testing.T) {
	s, err := NewSKU("abc-123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s.String() != "ABC-123" {
		t.Errorf("expected uppercased, got %s", s.String())
	}

	if _, err := NewSKU(""); err == nil {
		t.Error("empty SKU should fail")
	}
	if _, err := NewSKU("has space"); err == nil {
		t.Error("SKU with space should fail")
	}
}

func TestBarcodeEAN13(t *testing.T) {
	// 4006381333931 is a real valid EAN-13.
	valid := "4006381333931"
	b, err := NewBarcode(valid)
	if err != nil {
		t.Fatalf("valid barcode should pass: %v", err)
	}
	if b.String() != valid {
		t.Errorf("got %s, want %s", b.String(), valid)
	}

	// Tampered last digit -> invalid checksum
	bad := "4006381333932"
	if _, err := NewBarcode(bad); err == nil {
		t.Error("bad checksum should fail")
	}
	if _, err := NewBarcode("12345"); err == nil {
		t.Error("short barcode should fail")
	}
	if _, err := NewBarcode("12345678901234"); err == nil {
		t.Error("14-digit barcode should fail")
	}
}

func TestDocumentNumber(t *testing.T) {
	dn, err := NewDocumentNumber(enums.DocumentTypeDNI, "12345678")
	if err != nil || dn.Type() != enums.DocumentTypeDNI || dn.Number() != "12345678" {
		t.Errorf("DNI: %+v, %v", dn, err)
	}
	if _, err := NewDocumentNumber(enums.DocumentTypeDNI, "1234567"); err == nil {
		t.Error("7-digit DNI should fail")
	}
	r, err := NewDocumentNumber(enums.DocumentTypeRUC, "20123456789")
	if err != nil || r.Type() != enums.DocumentTypeRUC {
		t.Errorf("RUC: %+v, %v", r, err)
	}
	if _, err := NewDocumentNumber(enums.DocumentTypeRUC, "1012345678"); err == nil {
		t.Error("10-digit RUC should fail")
	}
	if _, err := NewDocumentNumber(enums.DocumentType("FAKE"), "12345678"); err == nil {
		t.Error("unknown doc type should fail")
	}
}

func TestCurrencyCode(t *testing.T) {
	c, err := NewCurrencyCode("pen")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c.String() != "PEN" {
		t.Errorf("got %s, want PEN", c.String())
	}
	if _, err := NewCurrencyCode("US"); err == nil {
		t.Error("2-letter code should fail")
	}
	if _, err := NewCurrencyCode("US1"); err == nil {
		t.Error("non-letter code should fail")
	}
}

func TestExchangeRateConvert(t *testing.T) {
	r, _ := ExchangeRateFromString("3.75")
	m, _ := MoneyFromString("100.00")
	got := r.Convert(m)
	if got.String() != "375.00" {
		t.Errorf("convert: %s", got)
	}
	if _, err := ExchangeRateFromString("0"); err == nil {
		t.Error("zero rate should fail")
	}
	if _, err := ExchangeRateFromString("-1"); err == nil {
		t.Error("negative rate should fail")
	}
}

func TestAddress(t *testing.T) {
	a, err := NewAddress("Calle 1", "Lima 01", "Peru")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if a.String() != "Calle 1, Lima 01, Peru" {
		t.Errorf("got %s", a.String())
	}
	if _, err := NewAddress("", "   "); err == nil {
		t.Error("all-empty address should fail")
	}
}

func TestFullName(t *testing.T) {
	n, err := NewFullName("  María García  ")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n.String() != "María García" {
		t.Errorf("got %s", n.String())
	}
	if _, err := NewFullName(""); err == nil {
		t.Error("empty name should fail")
	}
}

func TestShortCode(t *testing.T) {
	c, err := NewShortCode("vfi")
	if err != nil || c.String() != "VFI" {
		t.Errorf("got %s, %v", c, err)
	}
	if _, err := NewShortCode("vfi-01"); err != nil {
		t.Errorf("vfi-01 should pass: %v", err)
	}
	if _, err := NewShortCode("has space"); err == nil {
		t.Error("spaces should fail")
	}
}

func TestChartOfAccountsCode(t *testing.T) {
	c, err := NewChartOfAccountsCode("1.1.01")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c.Depth() != 3 {
		t.Errorf("depth: %d", c.Depth())
	}
	if c.Parent().String() != "1.1" {
		t.Errorf("parent: %s", c.Parent())
	}
	if _, err := NewChartOfAccountsCode(""); err == nil {
		t.Error("empty should fail")
	}
	if _, err := NewChartOfAccountsCode("1.1.1.1.1.1"); err == nil {
		t.Error("6-level depth should fail")
	}
	if _, err := NewChartOfAccountsCode("1.1.0a"); err == nil {
		t.Error("non-numeric segment should fail")
	}
}
