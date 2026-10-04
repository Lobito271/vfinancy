package purchasing

import (
	"sort"
	"strings"

	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/valueobjects"
)

// ProfitMonth is one calendar month of the purchase-side cost
// breakdown. Purchase costs are attributed to the month of
// Purchase.OrderDate.
type ProfitMonth struct {
	Year  int
	Month int

	// Cost amounts in USD, the purchase cost currency.
	BaseCostUSD  valueobjects.Money
	ExtraCostUSD valueobjects.Money
	TotalCostUSD valueobjects.Money

	// PEN amounts of the same costs. BaseCostPEN and ExtraCostPEN always
	// add up to TotalCostPEN: the import factor carried by the landed
	// cost is part of the base, never an unaccounted remainder.
	BaseCostPEN  valueobjects.Money
	ExtraCostPEN valueobjects.Money
	TotalCostPEN valueobjects.Money
}

// MonthlyCost is one row of the grouped purchase-cost aggregate, in
// both USD and PEN so the caller never has to re-derive the rate.
type MonthlyCost struct {
	Year         int
	Month        int
	BaseCostUSD  string
	ExtraCostUSD string
	ExtraCostPEN string
	TotalCostPEN string
}

// MonthlyExtraCost is one surcharge of the period, tagged with the month
// of its purchase and with both exchange rates involved: the snapshot of
// the surcharge itself and the rate of the purchase that carries it.
type MonthlyExtraCost struct {
	Year         int
	Month        int
	Concept      string
	Amount       string
	CurrencyCode string
	CostRate     string
	PurchaseRate string
}

// ConceptExtraCost is one extra-cost concept of a period, normalised
// into USD the same way ExtraCost.AmountUSD does and projected into PEN
// at the rate of the purchases that carry it.
type ConceptExtraCost struct {
	Concept string
	USD     valueobjects.Money
	PEN     valueobjects.Money
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
	// ExtraCosts breaks the month's surcharges down by concept, largest
	// first.
	ExtraCosts []ConceptExtraCost
}

// MonthlyProfit joins grouped purchase costs with grouped sales revenue
// into a continuous month-by-month series. Months missing from either
// side are reported with zeroes rather than skipped, and the series is
// ordered oldest first.
func MonthlyProfit(costs []MonthlyCost, extraCosts []MonthlyExtraCost, sales map[int]valueobjects.Money, months []int) []ProfitBreakdown {
	// Month keys are 0-based (y*12 + m-1); SQL returns a 1-based
	// month, so normalise here rather than at each call site.
	byKey := make(map[int]MonthlyCost, len(costs))
	for _, c := range costs {
		byKey[c.Year*12+c.Month-1] = c
	}
	byConcept := groupExtraCostsByMonth(extraCosts)
	out := make([]ProfitBreakdown, 0, len(months))
	for _, key := range months {
		year, month := key/12, key%12
		c, ok := byKey[key]
		var (
			baseUSD, extraUSD, totalPEN, extraPEN valueobjects.Money
		)
		if ok {
			baseUSD = moneyOrZero(c.BaseCostUSD)
			extraUSD = moneyOrZero(c.ExtraCostUSD)
			totalPEN = moneyOrZero(c.TotalCostPEN)
			extraPEN = moneyOrZero(c.ExtraCostPEN)
		}
		// The landed cost is the recorded one, so the base takes whatever
		// the import factor added on top of the surcharges.
		basePEN := totalPEN.Sub(extraPEN)
		salesPEN := sales[key]
		profit, err := valueobjects.MoneyFromDecimal(salesPEN.Decimal().Sub(totalPEN.Decimal()))
		if err != nil {
			profit = valueobjects.Zero()
		}
		out = append(out, ProfitBreakdown{
			Month: ProfitMonth{
				Year:         year,
				Month:        month,
				BaseCostUSD:  baseUSD,
				ExtraCostUSD: extraUSD,
				TotalCostUSD: addMoney(baseUSD, extraUSD),
				BaseCostPEN:  basePEN,
				ExtraCostPEN: extraPEN,
				TotalCostPEN: totalPEN,
			},
			SalesPEN:   salesPEN,
			ProfitPEN:  profit,
			ExtraCosts: byConcept[key],
		})
	}
	return out
}

// groupExtraCostsByMonth normalises every surcharge into USD exactly the
// way the purchase recorded it and projects it into PEN at the rate of
// its purchase, then sums the period by concept.
func groupExtraCostsByMonth(rows []MonthlyExtraCost) map[int][]ConceptExtraCost {
	type sum struct{ usd, pen decimal.Decimal }
	grouped := make(map[int]map[string]*sum)
	for _, r := range rows {
		amount := moneyOrZero(r.Amount)
		usd := amount
		if r.CurrencyCode != USD {
			rate := moneyOrZero(r.CostRate)
			if !rate.IsPositive() {
				continue
			}
			converted, err := valueobjects.MoneyFromDecimal(amount.Decimal().Div(rate.Decimal()))
			if err != nil {
				continue
			}
			usd = converted.RoundToCurrencyPrecision()
		}
		key := r.Year*12 + r.Month - 1
		if grouped[key] == nil {
			grouped[key] = make(map[string]*sum)
		}
		concept := strings.TrimSpace(r.Concept)
		if concept == "" {
			concept = "Otros"
		}
		s := grouped[key][concept]
		if s == nil {
			s = &sum{}
			grouped[key][concept] = s
		}
		s.usd = s.usd.Add(usd.Decimal())
		s.pen = s.pen.Add(usd.Decimal().Mul(moneyOrZero(r.PurchaseRate).Decimal()))
	}
	out := make(map[int][]ConceptExtraCost, len(grouped))
	for key, byConcept := range grouped {
		list := make([]ConceptExtraCost, 0, len(byConcept))
		for concept, s := range byConcept {
			list = append(list, ConceptExtraCost{
				Concept: concept,
				USD:     moneyOrDecimal(s.usd),
				PEN:     moneyOrDecimal(s.pen),
			})
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].PEN.Cmp(list[j].PEN) > 0
		})
		out[key] = list
	}
	return out
}

// MonthKeys returns the month keys of a calendar month, or of a whole
// year when month is 0.
func MonthKeys(year, month int) []int {
	if month > 0 {
		return []int{year*12 + month - 1}
	}
	keys := make([]int, 0, 12)
	for m := 1; m <= 12; m++ {
		keys = append(keys, year*12+m-1)
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

func moneyOrDecimal(d decimal.Decimal) valueobjects.Money {
	m, err := valueobjects.MoneyFromDecimal(d)
	if err != nil {
		return valueobjects.Zero()
	}
	return m.RoundToCurrencyPrecision()
}

func addMoney(a, b valueobjects.Money) valueobjects.Money {
	sum, err := valueobjects.MoneyFromDecimal(a.Decimal().Add(b.Decimal()))
	if err != nil {
		return valueobjects.Zero()
	}
	return sum
}
