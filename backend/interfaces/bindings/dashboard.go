package bindings

import (
	"time"

	"vfinancy/backend/internal/features/purchasing"
)

// ProfitExtraCostDTO is one extra-cost concept of a period, in both
// currencies: USD is what the purchases recorded, PEN is that amount
// converted at the rate of each purchase that carries it.
type ProfitExtraCostDTO struct {
	Concept   string  `json:"concept"`
	AmountUSD float64 `json:"amountUsd"`
	AmountPEN float64 `json:"amountPen"`
}

// ProfitBreakdownDTO is one calendar month of the profit breakdown.
// Every amount is in PEN so the profit column is a real subtraction, and
// base + extra always equals total.
type ProfitBreakdownDTO struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	// Label is the preformatted month name for the chart axis.
	Label string `json:"label"`
	// BaseCostUSD is the sum of the base costs of the month's purchases.
	BaseCostUSD float64 `json:"baseCostUsd"`
	// ExtraCostUSD is the sum of their extra costs.
	ExtraCostUSD float64 `json:"extraCostUsd"`
	// TotalCostUSD is BaseCostUSD + ExtraCostUSD.
	TotalCostUSD float64 `json:"totalCostUsd"`
	// BaseCostPEN and ExtraCostPEN split TotalCostPEN; the base carries
	// the import factor of the landed cost.
	BaseCostPEN  float64 `json:"baseCostPen"`
	ExtraCostPEN float64 `json:"extraCostPen"`
	// TotalCostPEN is the landed cost, converted at each purchase's own
	// exchange-rate snapshot.
	TotalCostPEN float64 `json:"totalCostPen"`
	// SalesPEN is the sum of the month's non-cancelled sale totals.
	SalesPEN float64 `json:"salesPen"`
	// ProfitPEN is SalesPEN - TotalCostPEN.
	ProfitPEN float64 `json:"profitPen"`
	// ExtraCosts breaks the month's surcharges down by concept.
	ExtraCosts []ProfitExtraCostDTO `json:"extraCosts"`
}

// ListProfitBreakdown returns the profit breakdown of a calendar month,
// or of a whole year when month is 0, oldest first. Purchase costs are
// attributed to the month of the purchase date; revenue is the
// non-cancelled sales of the same calendar month. A year below 1 falls
// back to the current one and a month out of range to the current month.
//
// Utilidad = Suma de precios de venta - (Suma de costos base + Suma de costos extras).
func (a *App) ListProfitBreakdown(year int, month int) ([]ProfitBreakdownDTO, error) {
	now := time.Now().UTC()
	if year < 1 {
		year = now.Year()
	}
	if month < 0 || month > 12 {
		month = int(now.Month())
	}
	keys := purchasing.MonthKeys(year, month)
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	if month == 0 {
		first = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	until := first.AddDate(0, 1, 0)

	revenue, err := a.salesSvc.MonthlyRevenue(a.Context(), first, until)
	if err != nil {
		return nil, err
	}
	rows := a.purchasingSvc.ProfitBreakdown(a.Context(), keys, first, until, revenue)

	out := make([]ProfitBreakdownDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, profitBreakdownDTO(r))
	}
	return out, nil
}

func profitBreakdownDTO(r purchasing.ProfitBreakdown) ProfitBreakdownDTO {
	extraCosts := make([]ProfitExtraCostDTO, 0, len(r.ExtraCosts))
	for _, c := range r.ExtraCosts {
		extraCosts = append(extraCosts, ProfitExtraCostDTO{
			Concept:   c.Concept,
			AmountUSD: moneyFloat(c.USD),
			AmountPEN: moneyFloat(c.PEN),
		})
	}
	return ProfitBreakdownDTO{
		Year:         r.Month.Year,
		Month:        r.Month.Month,
		Label:        monthLabel(r.Month.Year, r.Month.Month),
		BaseCostUSD:  moneyFloat(r.Month.BaseCostUSD),
		ExtraCostUSD: moneyFloat(r.Month.ExtraCostUSD),
		TotalCostUSD: moneyFloat(r.Month.TotalCostUSD),
		BaseCostPEN:  moneyFloat(r.Month.BaseCostPEN),
		ExtraCostPEN: moneyFloat(r.Month.ExtraCostPEN),
		TotalCostPEN: moneyFloat(r.Month.TotalCostPEN),
		SalesPEN:     moneyFloat(r.SalesPEN),
		ProfitPEN:    moneyFloat(r.ProfitPEN),
		ExtraCosts:   extraCosts,
	}
}

// monthLabel renders "sept 2026" for the chart axis.
func monthLabel(year, month int) string {
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return t.Format("jan 2006")
}
