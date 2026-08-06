package valueobjects

import "strings"

// ChartOfAccountsCode is a hierarchical chart-of-accounts code, e.g.
// "1.1.01" (Asset > Current > Cash). It must be a non-empty
// dot-separated sequence of integers, with up to 5 levels.
type ChartOfAccountsCode string

func NewChartOfAccountsCode(s string) (ChartOfAccountsCode, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ChartOfAccountsCode(""), wrapInvalid("chart-of-accounts code is empty")
	}
	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 5 {
		return ChartOfAccountsCode(""), wrapInvalid("chart-of-accounts code must be 1-5 dot-separated integers (e.g. '1.1.01')")
	}
	for _, p := range parts {
		if p == "" || len(p) > 2 {
			return ChartOfAccountsCode(""), wrapInvalid("chart-of-accounts code must be 1-5 dot-separated integers (e.g. '1.1.01')")
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return ChartOfAccountsCode(""), wrapInvalid("chart-of-accounts code must be 1-5 dot-separated integers (e.g. '1.1.01')")
			}
		}
	}
	return ChartOfAccountsCode(s), nil
}

func (c ChartOfAccountsCode) String() string            { return string(c) }
func (c ChartOfAccountsCode) Depth() int                { return strings.Count(string(c), ".") + 1 }
func (c ChartOfAccountsCode) Parent() ChartOfAccountsCode {
	s := string(c)
	if i := strings.LastIndex(s, "."); i > 0 {
		return ChartOfAccountsCode(s[:i])
	}
	return ChartOfAccountsCode("")
}
func (c ChartOfAccountsCode) IsZero() bool              { return string(c) == "" }
func (c ChartOfAccountsCode) Equals(other ChartOfAccountsCode) bool { return c == other }
