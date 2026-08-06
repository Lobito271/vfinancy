package valueobjects

import (
	"testing"
)

func TestQuantity(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"0", "0.0000", false},
		{"1", "1.0000", false},
		{"12.5", "12.5000", false},
		{"0.0001", "0.0001", false},
		{"abc", "", true},
		{"", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			q, err := QuantityFromString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if err == nil && q.String() != tc.want {
				t.Errorf("got %s, want %s", q.String(), tc.want)
			}
		})
	}
}

func TestQuantityArithmetic(t *testing.T) {
	a, _ := QuantityFromString("10.5")
	b, _ := QuantityFromString("2.25")
	if a.Add(b).String() != "12.7500" {
		t.Errorf("add: %s", a.Add(b))
	}
	if a.Sub(b).String() != "8.2500" {
		t.Errorf("sub: %s", a.Sub(b))
	}
	if a.Neg().String() != "-10.5000" {
		t.Errorf("neg: %s", a.Neg())
	}
	if a.GreaterThan(b) {
		t.Log("ok")
	} else {
		t.Error("a > b expected")
	}
}
