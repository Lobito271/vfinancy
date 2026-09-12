package treasury

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

type fakeCardRepo struct{}

func (fakeCardRepo) Create(context.Context, *CreditCard) error   { return nil }
func (fakeCardRepo) Update(context.Context, *CreditCard) error   { return nil }
func (fakeCardRepo) SoftDelete(context.Context, uuid.UUID) error { return nil }
func (fakeCardRepo) GetByID(context.Context, uuid.UUID) (*CreditCard, error) {
	return nil, repositories.ErrNotFound
}
func (fakeCardRepo) GetByIDForUpdate(context.Context, uuid.UUID) (*CreditCard, error) {
	return nil, repositories.ErrNotFound
}
func (fakeCardRepo) List(context.Context) ([]*CreditCard, error) { return nil, nil }

func (fakeCardRepo) CycleTotals(context.Context, uuid.UUID, time.Time, time.Time) (valueobjects.Money, valueobjects.Money, error) {
	return valueobjects.Zero(), valueobjects.Zero(), nil
}

type fakeRateRepo struct {
	snap    *ExchangeRateSnapshot
	upserts int
}

func (r *fakeRateRepo) Latest(context.Context, valueobjects.CurrencyCode, valueobjects.CurrencyCode) (*ExchangeRateSnapshot, error) {
	if r.snap == nil {
		return nil, repositories.ErrNotFound
	}
	return r.snap, nil
}

func (r *fakeRateRepo) Upsert(context.Context, valueobjects.CurrencyCode, valueobjects.CurrencyCode, time.Time, valueobjects.Money, string) error {
	r.upserts++
	return nil
}

func testService(erURL string) (*TreasuryService, *fakeRateRepo) {
	rates := &fakeRateRepo{}
	s := New(fakeCardRepo{}, rates, nil, logger.New("error", "text", ""))
	s.live = &liveRateProvider{
		client:   &http.Client{},
		sunatURL: "http://127.0.0.1:1/unreachable",
		erAPIURL: erURL,
	}
	return s, rates
}

func TestNextBillingCycleClamp(t *testing.T) {
	card := &CreditCard{CutOffDay: 31, PaymentDueDay: 31}

	start, end, due := NextBillingCycle(card, time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC))
	if got := end.Format("2006-01-02"); got != "2025-02-28" {
		t.Fatalf("cut-off 31 in February: end = %s, want 2025-02-28", got)
	}
	if got := start.Format("2006-01-02"); got != "2025-02-01" {
		t.Fatalf("start = %s, want 2025-02-01", got)
	}
	if got := due.Format("2006-01-02"); got != "2025-03-31" {
		t.Fatalf("due = %s, want 2025-03-31", got)
	}

	start, end, due = NextBillingCycle(card, time.Date(2025, time.April, 15, 0, 0, 0, 0, time.UTC))
	if got := end.Format("2006-01-02"); got != "2025-04-30" {
		t.Fatalf("cut-off 31 in April: end = %s, want 2025-04-30", got)
	}
	if got := start.Format("2006-01-02"); got != "2025-04-01" {
		t.Fatalf("start = %s, want 2025-04-01", got)
	}
	if got := due.Format("2006-01-02"); got != "2025-05-31" {
		t.Fatalf("due = %s, want 2025-05-31", got)
	}
}

func TestLatestExchangeRateFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	s, rates := testService(srv.URL + "/latest/%s")
	s.SetFallbackRate(func(context.Context) float64 { return 3.9 })
	usd := valueobjects.MustCurrencyCode("USD")
	pen := valueobjects.MustCurrencyCode("PEN")

	info, err := s.LatestExchangeRate(context.Background(), usd, pen)
	if err != nil {
		t.Fatalf("LatestExchangeRate: %v", err)
	}
	if !info.IsFallback {
		t.Fatal("IsFallback = false, want true")
	}
	if info.Source != SourceFallback {
		t.Fatalf("source = %q, want fallback", info.Source)
	}
	if got := info.Rate.String(); got != "3.90" {
		t.Fatalf("rate = %s, want 3.90", got)
	}
	if rates.upserts != 0 {
		t.Fatalf("fallback must not be cached, got %d upserts", rates.upserts)
	}

	_, err = s.LatestExchangeRate(context.Background(), valueobjects.MustCurrencyCode("EUR"), pen)
	if !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("non-USD/PEN pair error = %v, want ErrNotFound", err)
	}
}
