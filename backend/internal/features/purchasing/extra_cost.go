package purchasing

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// ExtraCost is an ad-hoc surcharge recorded against a purchase
// (freight, customs, handling). It is summed into the purchase's
// extra_cost_usd and total_cost_usd. The exchange rate snapshot is the
// USD->PEN rate at the moment the cost was assigned or last updated.
type ExtraCost struct {
	ID               uuid.UUID
	PurchaseID       uuid.UUID
	Concept          string
	Amount           valueobjects.Money
	CurrencyCode     valueobjects.CurrencyCode
	ExchangeRate     valueobjects.ExchangeRate
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate checks the extra-cost invariants: a non-blank concept, a
// non-negative amount, an ISO 4217 currency code and a positive rate.
func (e *ExtraCost) Validate() error {
	if strings.TrimSpace(e.Concept) == "" {
		return derrors.Wrap(derrors.ErrRequired, errField("concept is required"))
	}
	if e.Amount.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("amount cannot be negative"))
	}
	if e.CurrencyCode.IsZero() {
		return derrors.Wrap(derrors.ErrRequired, errField("currency is required"))
	}
	if !e.ExchangeRate.Decimal().IsPositive() {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("exchange rate must be positive"))
	}
	return nil
}

// AmountUSD normalises the surcharge into USD, the purchase's cost
// currency. A cost already in USD is taken as-is; any other currency is
// divided by the USD->PEN rate snapshot. This mirrors how real_cost_pen
// carries a PEN projection back to USD.
func (e *ExtraCost) AmountUSD() valueobjects.Money {
	if e.CurrencyCode.String() == USD {
		return e.Amount
	}
	rate := e.ExchangeRate.Decimal()
	if !rate.IsPositive() {
		return valueobjects.Zero()
	}
	usd, err := valueobjects.MoneyFromDecimal(e.Amount.Decimal().Div(rate))
	if err != nil {
		return valueobjects.Zero()
	}
	return usd.RoundToCurrencyPrecision()
}

// SumExtraCostUSD adds up the normalised USD value of every surcharge of
// a purchase.
func SumExtraCostUSD(costs []*ExtraCost) valueobjects.Money {
	total := decimal.Zero
	for _, c := range costs {
		total = total.Add(c.AmountUSD().Decimal())
	}
	sum, err := valueobjects.MoneyFromDecimal(total)
	if err != nil {
		return valueobjects.Zero()
	}
	return sum.RoundToCurrencyPrecision()
}
