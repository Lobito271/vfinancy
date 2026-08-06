// Package reporting implements read-only reporting queries. Each method
// aggregates over one or more repositories and returns a typed result.
// The service never writes — that is the application's job.
package reporting

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/shared/logger"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/purchasing"
	"vfinancy/backend/internal/features/sales"
)

// ReportingService exposes the queries used by dashboards and reports.
// It is read-only.
type ReportingService struct {
	ar     sales.AccountsReceivableRepository
	ap     purchasing.AccountsPayableRepository
	sales  sales.SalesRepository
	log    *logger.Logger
}

// New returns a ReportingService ready for use.
func New(
	ar sales.AccountsReceivableRepository,
	ap purchasing.AccountsPayableRepository,
	sales sales.SalesRepository,
	log *logger.Logger,
) *ReportingService {
	return &ReportingService{ar: ar, ap: ap, sales: sales, log: log}
}

// ReceivableSummary is the customer-level AR roll-up.
type ReceivableSummary struct {
	Open           valueobjects.Money
	Overdue0to30   valueobjects.Money
	Overdue31to60  valueobjects.Money
	Overdue61to90  valueobjects.Money
	Overdue90plus  valueobjects.Money
}

// ReceivableSummaryByCustomer returns the AR breakdown for a customer.
func (s *ReportingService) ReceivableSummaryByCustomer(ctx context.Context, customerID uuid.UUID) (ReceivableSummary, error) {
	open, err := s.ar.GetOpenBalanceForCustomer(ctx, customerID)
	if err != nil {
		return ReceivableSummary{}, err
	}
	buckets, err := s.ar.ListAgingBucket(ctx, customerID)
	if err != nil {
		return ReceivableSummary{}, err
	}
	return ReceivableSummary{
		Open:          mustMoney(open),
		Overdue0to30:  mustMoney(buckets["0-30"]),
		Overdue31to60: mustMoney(buckets["31-60"]),
		Overdue61to90: mustMoney(buckets["61-90"]),
		Overdue90plus: mustMoney(buckets["90+"]),
	}, nil
}

// PayableSummary is the supplier-level AP roll-up.
type PayableSummary struct {
	Open           valueobjects.Money
	Overdue0to30   valueobjects.Money
	Overdue31to60  valueobjects.Money
	Overdue61to90  valueobjects.Money
	Overdue90plus  valueobjects.Money
}

// PayableSummaryBySupplier returns the AP breakdown for a supplier.
func (s *ReportingService) PayableSummaryBySupplier(ctx context.Context, supplierID uuid.UUID) (PayableSummary, error) {
	open, err := s.ap.GetOpenBalanceForSupplier(ctx, supplierID)
	if err != nil {
		return PayableSummary{}, err
	}
	buckets, err := s.ap.ListAgingBucket(ctx, supplierID)
	if err != nil {
		return PayableSummary{}, err
	}
	return PayableSummary{
		Open:          mustMoney(open),
		Overdue0to30:  mustMoney(buckets["0-30"]),
		Overdue31to60: mustMoney(buckets["31-60"]),
		Overdue61to90: mustMoney(buckets["61-90"]),
		Overdue90plus: mustMoney(buckets["90+"]),
	}, nil
}

// ProfitSummary aggregates sales revenue minus cost over a date range.
type ProfitSummary struct {
	Revenue  valueobjects.Money
	Cost     valueobjects.Money
	Tax      valueobjects.Money
	Discount valueobjects.Money
	Profit   valueobjects.Money
	Count    int
}

// ProfitInRange sums profits of all sales with issue_date in [from, to].
// The current implementation lists all sales and sums in memory; a
// future revision pushes the aggregation into SQL.
func (s *ReportingService) ProfitInRange(ctx context.Context, companyID uuid.UUID, from, to time.Time) (ProfitSummary, error) {
	page, err := s.sales.List(ctx, sales.SaleFilter{
		CompanyID: &companyID,
		IssueRange: repositories.TimeRange{From: from, To: to},
		PageRequest: repositories.PageRequest{Limit: 10000},
	})
	if err != nil {
		return ProfitSummary{}, err
	}
	out := ProfitSummary{}
	for _, s := range page.Items {
		out.Count++
		out.Revenue = out.Revenue.Add(s.CalculateTotal())
		out.Cost = out.Cost.Add(s.CostTotal)
		out.Tax = out.Tax.Add(s.TaxAmount)
		out.Discount = out.Discount.Add(s.DiscountAmount)
		out.Profit = out.Profit.Add(s.CalculateProfit())
	}
	return out, nil
}

func mustMoney(s string) valueobjects.Money {
	m, _ := valueobjects.MoneyFromString(s)
	return m
}
