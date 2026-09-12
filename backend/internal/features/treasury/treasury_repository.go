package treasury

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/valueobjects"
)

// CreditCardRepository persists company-issued credit cards.
type CreditCardRepository interface {
	Create(ctx context.Context, c *CreditCard) error
	Update(ctx context.Context, c *CreditCard) error
	// SoftDelete sets deleted_at and is_active=false in one step.
	SoftDelete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*CreditCard, error)
	// GetByIDForUpdate locks the row (SELECT ... FOR UPDATE) for the
	// duration of the surrounding write transaction.
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*CreditCard, error)
	List(ctx context.Context) ([]*CreditCard, error)

	// CycleTotals sums, for purchase orders charged to the card with
	// order_date in [from, to): totalUSD is the summed cost_usd of
	// non-cancelled orders; refundsUSD the summed refund_amount of
	// cancelled orders.
	CycleTotals(ctx context.Context, cardID uuid.UUID, from, to time.Time) (totalUSD, refundsUSD valueobjects.Money, err error)
}

// ExchangeRateSnapshot is a stored daily rate for a currency pair.
type ExchangeRateSnapshot struct {
	From     valueobjects.CurrencyCode
	To       valueobjects.CurrencyCode
	RateDate time.Time
	Rate     valueobjects.Money
	Source   string
}

// ExchangeRateRepository persists daily exchange rates (one row per
// (from, to, rate_date)).
type ExchangeRateRepository interface {
	// Latest returns the most recent snapshot by rate_date for the
	// pair. Returns repositories.ErrNotFound when nothing is stored.
	Latest(ctx context.Context, from, to valueobjects.CurrencyCode) (*ExchangeRateSnapshot, error)
	// Upsert inserts or overwrites the rate for the given date.
	Upsert(ctx context.Context, from, to valueobjects.CurrencyCode, rateDate time.Time, rate valueobjects.Money, source string) error
}
