package valueobjects

import (
	"testing"
)

func TestPercentageFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"whole", "18", "18.0000", false},
		{"fractional", "12.5", "12.5000", false},
		{"small", "0.5", "0.5000", false},
		{"zero", "0", "0.0000", false},
		{"hundred", "100", "100.0000", false},
		{"over", "101", "", true},
		{"negative", "-1", "", true},
		{"empty", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := PercentageFromString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if err == nil && p.String() != tc.want {
				t.Errorf("got %s, want %s", p.String(), tc.want)
			}
		})
	}
}

func TestPercentageAsDecimal(t *testing.T) {
	p, _ := PercentageFromString("18")
	// 18 / 100 = 0.18
	if p.AsDecimal().String() != "0.18" {
		t.Errorf("AsDecimal: %s", p.AsDecimal())
	}
}

func mustPercent(t *testing.T, s string) Percentage {
	t.Helper()
	p, err := PercentageFromString(s)
	if err != nil {
		t.Fatalf("mustPercent(%q): %v", s, err)
	}
	return p
}
