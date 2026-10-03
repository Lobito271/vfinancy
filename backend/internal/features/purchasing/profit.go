package purchasing

import (
	"time"

	"vfinancy/backend/internal/domain/valueobjects"
)

// ProfitMonth is one calendar month of the purchase-side cost
// breakdown. Purchase costs are attributed to the month of
// Purchase.OrderDate.
type ProfitMonth struct {
	Year  int
	Month int

	// Cost amounts in USD, the purchase cost currency.
	BaseCostUSD   valueobjects.Money
	ExtraCostUSD  valueobjects.Money
	TotalCostUSD  valueobjects.Money

	// TotalCostPEN is the landed cost in PEN: every purchase is
	// converted at its own exchange-rate snapshot, so the sum honours
	// the rate that was actually paid.
	TotalCostPEN valueobjects.Money
}

// MonthKey is the 0-based identity of a calendar month, matching the
// keys produced by MonthRange.
func (m ProfitMonth) MonthKey() int { return m.Year*12 + m.Month - 1 }

// MonthStart returns the first instant of the month in UTC.
func (m ProfitMonth) MonthStart() time.Time {
	return time.Date(m.Year, time.Month(m.Month), 1, 0, 0, 0, 0, time.UTC)
}

// MonthlyCost is one row of the grouped purchase-cost aggregate, in
// both USD and PEN so the caller never has to re-derive the rate.
type MonthlyCost struct {
	Year         int
	Month        int
	BaseCostUSD  string
	ExtraCostUSD string
	TotalCostPEN string
}

// ProfitBreakdown is a purchase month joined with the sales revenue of
// the same calendar month, plus the resulting profit. Everything is in
// PEN so the subtraction is meaningful.
type ProfitBreakdown struct {
	Month ProfitMonth
	// SalesPEN is the sum of the non-cancelled sales of the month.
	SalesPEN valueobjects.Money
	// ProfitPEN = SalesPEN - TotalCostPEN.
	ProfitPEN valueobjects.Money
}

// MonthlyProfit joins grouped purchase costs with grouped sales revenue
// into a continuous month-by-month series. Months missing from either
// side are reported with zeroes rather than skipped, and the series is
// ordered oldest first.
func MonthlyProfit(costs []MonthlyCost, sales map[int]valueobjects.Money, months []int) []ProfitBreakdown {
	// MonthRange keys are 0-based (y*12 + m-1); SQL returns a 1-based
	// month, so normalise here rather than at each call site.
	byKey := make(map[int]MonthlyCost, len(costs))
	for _, c := range costs {
		byKey[c.Year*12+c.Month-1] = c
	}
	out := make([]ProfitBreakdown, 0, len(months))
	for _, key := range months {
		year, month := key/12, key%12
		c, ok := byKey[key]
		var (
			baseUSD, extraUSD, totalPEN valueobjects.Money
		)
		if ok {
			baseUSD = moneyOrZero(c.BaseCostUSD)
			extraUSD = moneyOrZero(c.ExtraCostUSD)
			totalPEN = moneyOrZero(c.TotalCostPEN)
		}
		salesPEN := sales[key]
		profit, err := valueobjects.MoneyFromDecimal(salesPEN.Decimal().Sub(totalPEN.Decimal()))
		if err != nil {
			profit = valueobjects.Zero()
		}
		out = append(out, ProfitBreakdown{
			Month: ProfitMonth{
				Year:          year,
				Month:         month,
				BaseCostUSD:   baseUSD,
				ExtraCostUSD:  extraUSD,
				TotalCostUSD:  addMoney(baseUSD, extraUSD),
				TotalCostPEN:  totalPEN,
			},
			SalesPEN:  salesPEN,
			ProfitPEN: profit,
		})
	}
	return out
}

// MonthRange returns the contiguous month keys covering the last n
// months up to and including the month of at, oldest first.
func MonthRange(at time.Time, n int) []int {
	if n <= 0 {
		return nil
	}
	first := time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -(n - 1), 0)
	keys := make([]int, 0, n)
	for i := 0; i < n; i++ {
		m := first.AddDate(0, i, 0)
		keys = append(keys, m.Year()*12+int(m.Month())-1)
	}
	return keys
}

func moneyOrZero(s string) valueobjects.Money {
	if s == "" {
		return valueobjects.Zero()
	}
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		return valueobjects.Zero()
	}
	return m
}

func addMoney(a, b valueobjects.Money) valueobjects.Money {
	sum, err := valueobjects.MoneyFromDecimal(a.Decimal().Add(b.Decimal()))
	if err != nil {
		return valueobjects.Zero()
	}
	return sum
}