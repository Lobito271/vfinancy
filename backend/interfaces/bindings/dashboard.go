package bindings

import (
	"time"

	"vfinancy/backend/internal/features/purchasing"
)

// MonthlyProfitDTO is one calendar month of the profit breakdown.
// Every amount is in PEN so the profit column is a real subtraction.
type MonthlyProfitDTO struct {
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
	// TotalCostPEN is the landed cost, converted at each purchase's own
	// exchange-rate snapshot.
	TotalCostPEN float64 `json:"totalCostPen"`
	// SalesPEN is the sum of the month's non-cancelled sale totals.
	SalesPEN float64 `json:"salesPen"`
	// ProfitPEN is SalesPEN - TotalCostPEN.
	ProfitPEN float64 `json:"profitPen"`
}

// DefaultProfitMonths is how many months ListMonthlyProfit returns when
// the caller does not ask for a specific window.
const DefaultProfitMonths = 12

// ListMonthlyProfit returns the profit breakdown of the last `months`
// calendar months, oldest first. Purchase costs are attributed to the
// month of the purchase date; revenue is the non-cancelled sales of the
// same month.
//
// Utilidad = Suma de precios de venta - (Suma de costos base + Suma de costos extras).
func (a *App) ListMonthlyProfit(months int) ([]MonthlyProfitDTO, error) {
	if months <= 0 || months > 36 {
		months = DefaultProfitMonths
	}
	now := time.Now().UTC()
	keys := purchasing.MonthRange(now, months)
	if len(keys) == 0 {
		return []MonthlyProfitDTO{}, nil
	}
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -(months - 1), 0)
	until := first.AddDate(0, months, 0)

	revenue, err := a.salesSvc.MonthlyRevenue(a.Context(), first, until)
	if err != nil {
		return nil, err
	}
	rows := a.purchasingSvc.MonthlyProfit(a.Context(), now, months, revenue)

	out := make([]MonthlyProfitDTO, 0, len(keys))
	for _, r := range rows {
		out = append(out, monthlyProfitDTO(r))
	}
	return out, nil
}

func monthlyProfitDTO(r purchasing.ProfitBreakdown) MonthlyProfitDTO {
	return MonthlyProfitDTO{
		Year:         r.Month.Year,
		Month:        r.Month.Month,
		Label:        monthLabel(r.Month.Year, r.Month.Month),
		BaseCostUSD:  moneyFloat(r.Month.BaseCostUSD),
		ExtraCostUSD: moneyFloat(r.Month.ExtraCostUSD),
		TotalCostUSD: moneyFloat(r.Month.TotalCostUSD),
		TotalCostPEN: moneyFloat(r.Month.TotalCostPEN),
		SalesPEN:     moneyFloat(r.SalesPEN),
		ProfitPEN:    moneyFloat(r.ProfitPEN),
	}
}

// monthLabel renders "sept 2026" for the chart axis.
func monthLabel(year, month int) string {
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return t.Format("jan 2006")
}
