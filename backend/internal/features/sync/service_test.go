package sync

import (
	"testing"
	"time"
)

func TestResolveLWW(t *testing.T) {
	base := time.UnixMilli(1_700_000_000_000).UTC()
	cases := []struct {
		name          string
		local, remote time.Time
		want          bool
	}{
		{"local missing", time.Time{}, base, true},
		{"remote newer", base, base.Add(time.Second), true},
		{"local newer", base.Add(time.Second), base, false},
		{"equal re-applies", base, base, true},
	}
	for _, tc := range cases {
		if got := ResolveLWW(tc.local, tc.remote); got != tc.want {
			t.Errorf("%s: ResolveLWW = %v, want %v", tc.name, got, tc.want)
		}
	}
}
