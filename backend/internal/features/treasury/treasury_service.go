// Package treasury implements the business logic for the treasury
// module: credit cards, their billing cycles, and exchange rates.
package treasury

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// StatusSettled marks a billing cycle whose cut-off has passed.
const StatusSettled = "settled"

// StatusOpen marks a billing cycle still collecting charges.
const StatusOpen = "open"

// TreasuryService owns the card / exchange-rate workflows.
type TreasuryService struct {
	cards    CreditCardRepository
	rates    ExchangeRateRepository
	txm      repositories.TransactionManager
	log      *logger.Logger
	live     *liveRateProvider
	fallback func(context.Context) float64
}

// New returns a TreasuryService ready for use.
func New(
	cards CreditCardRepository,
	rates ExchangeRateRepository,
	txm repositories.TransactionManager,
	log *logger.Logger,
) *TreasuryService {
	return &TreasuryService{
		cards: cards,
		rates: rates,
		txm:   txm,
		log:   log,
		live:  newLiveRateProvider(),
	}
}

// SetFallbackRate registers the function that yields the preferred
// PEN/USD rate from workspace preferences. Without it the built-in
// DefaultFallbackUSDPEN applies.
func (s *TreasuryService) SetFallbackRate(fn func(context.Context) float64) {
	if fn != nil {
		s.fallback = fn
	}
}

// IssueCardInput creates a company-issued credit card.
type IssueCardInput struct {
	Issuer        string
	LastFour      string
	CreditLimit   valueobjects.Money
	CutOffDay     int
	PaymentDueDay int
}

// IssueCard creates a new card with zero opening balance.
func (s *TreasuryService) IssueCard(ctx context.Context, in IssueCardInput) (*CreditCard, error) {
	card, err := NewCreditCard(time.Now().UTC(), NewCreditCardOptions{
		Issuer:        in.Issuer,
		LastFour:      in.LastFour,
		CreditLimit:   in.CreditLimit,
		CutOffDay:     in.CutOffDay,
		PaymentDueDay: in.PaymentDueDay,
	})
	if err != nil {
		return nil, err
	}
	if err := s.cards.Create(ctx, card); err != nil {
		return nil, err
	}
	s.log.Info("credit card issued", "card_id", card.ID, "issuer", card.Issuer)
	return card, nil
}

// UpdateCardInput describes editable fields of a credit card; nil
// fields are left unchanged.
type UpdateCardInput struct {
	ID            uuid.UUID
	Issuer        *string
	LastFour      *string
	CreditLimit   *valueobjects.Money
	CutOffDay     *int
	PaymentDueDay *int
}

// UpdateCard persists changes to an existing credit card.
func (s *TreasuryService) UpdateCard(ctx context.Context, in UpdateCardInput) (*CreditCard, error) {
	var out *CreditCard
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		card, err := s.cards.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if in.Issuer != nil {
			card.Issuer = *in.Issuer
		}
		if in.LastFour != nil {
			card.LastFour = *in.LastFour
		}
		if in.CreditLimit != nil {
			card.CreditLimit = *in.CreditLimit
		}
		if in.CutOffDay != nil {
			card.CutOffDay = *in.CutOffDay
		}
		if in.PaymentDueDay != nil {
			card.PaymentDueDay = *in.PaymentDueDay
		}
		card.UpdatedAt = time.Now().UTC()
		if err := card.Validate(); err != nil {
			return err
		}
		if err := s.cards.Update(ctx, card); err != nil {
			return err
		}
		out = card
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("credit card updated", "card_id", in.ID)
	return out, nil
}

// DeleteCard soft-deletes a credit card, deactivating it in the same
// step.
func (s *TreasuryService) DeleteCard(ctx context.Context, id uuid.UUID) error {
	if err := s.cards.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.log.Info("credit card deleted", "card_id", id)
	return nil
}

// GetCard returns a single credit card by ID.
func (s *TreasuryService) GetCard(ctx context.Context, cardID uuid.UUID) (*CreditCard, error) {
	return s.cards.GetByID(ctx, cardID)
}

// ListCards returns every credit card that has not been soft-deleted,
// including deactivated ones (they keep history for projections).
func (s *TreasuryService) ListCards(ctx context.Context) ([]*CreditCard, error) {
	return s.cards.List(ctx)
}

// ChargeCard records a purchase on a credit card. The amount is added
// to the card's outstanding balance; the caller is expected to record
// the corresponding journal entry.
func (s *TreasuryService) ChargeCard(ctx context.Context, cardID uuid.UUID, amount valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		card, err := s.cards.GetByIDForUpdate(ctx, cardID)
		if err != nil {
			return err
		}
		if err := card.Charge(amount); err != nil {
			return err
		}
		card.UpdatedAt = time.Now().UTC()
		return s.cards.Update(ctx, card)
	})
	if err != nil {
		return err
	}
	s.log.Info("credit card charged", "card_id", cardID, "amount", amount)
	return nil
}

// ReleaseCardCharge reverses a charge previously made against a card,
// flooring the balance at zero.
func (s *TreasuryService) ReleaseCardCharge(ctx context.Context, cardID uuid.UUID, amount valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		card, err := s.cards.GetByIDForUpdate(ctx, cardID)
		if err != nil {
			return err
		}
		if err := card.Release(amount); err != nil {
			return err
		}
		card.UpdatedAt = time.Now().UTC()
		return s.cards.Update(ctx, card)
	})
	if err != nil {
		return err
	}
	s.log.Info("credit card charge released", "card_id", cardID, "amount", amount)
	return nil
}

// PayCard records a payment that liquidates card debt. Overpayments
// settle the balance to zero instead of failing.
func (s *TreasuryService) PayCard(ctx context.Context, cardID uuid.UUID, amount valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		card, err := s.cards.GetByIDForUpdate(ctx, cardID)
		if err != nil {
			return err
		}
		if err := card.Pay(amount); err != nil {
			return err
		}
		card.UpdatedAt = time.Now().UTC()
		return s.cards.Update(ctx, card)
	})
	if err != nil {
		return err
	}
	s.log.Info("credit card payment", "card_id", cardID, "amount", amount)
	return nil
}

// CurrentCycleStart returns the start of the card's current billing
// cycle: the day after the last past cut-off.
func (s *TreasuryService) CurrentCycleStart(ctx context.Context, cardID uuid.UUID) (time.Time, error) {
	card, err := s.cards.GetByID(ctx, cardID)
	if err != nil {
		return time.Time{}, err
	}
	return card.CurrentCycleStart(time.Now().UTC()), nil
}

// ProjectPayments returns one projection per non-defaulted active
// credit card for its current billing cycle. TotalUSD sums the USD
// costs charged inside the cycle; RefundsUSD the cancelled-order
// refunds.
func (s *TreasuryService) ProjectPayments(ctx context.Context) ([]CardPaymentProjection, error) {
	cards, err := s.cards.List(ctx)
	if err != nil {
		return nil, err
	}
	projections := make([]CardPaymentProjection, 0, len(cards))
	now := time.Now().UTC()
	for _, card := range cards {
		if !card.IsActive {
			continue
		}
		start := card.CurrentCycleStart(now)
		end := card.cycleEnd(start)
		due := monthDay(addMonth(end), card.PaymentDueDay)
		total, refunds, err := s.cards.CycleTotals(ctx, card.ID, start, end)
		if err != nil {
			return nil, err
		}
		status := StatusOpen
		if dateAfter(now, end) {
			status = StatusSettled
		}
		projections = append(projections, CardPaymentProjection{
			Card:       card,
			CycleStart: start,
			CycleEnd:   end,
			PaymentDue: due,
			TotalUSD:   total,
			RefundsUSD: refunds,
			Status:     status,
		})
	}
	return projections, nil
}

// ExchangeRateInfo is a resolved exchange rate plus its provenance.
type ExchangeRateInfo struct {
	Rate valueobjects.Money
	// Source is the provider that produced the rate ("manual",
	// "sunat", "apis.net.pe", "open.er-api.com", "fallback", ...).
	Source string
	// IsFallback reports hard OFFLINE degradation to the configured
	// default rate.
	IsFallback bool
}

// UpsertExchangeRateSnapshot persists a rate snapshot for today. This
// is the cache write used by the provider resolution; manual
// overrides stored in the exchange_rates table are also served from
// there.
func (s *TreasuryService) UpsertExchangeRateSnapshot(ctx context.Context, from, to valueobjects.CurrencyCode, rate valueobjects.Money, source string) error {
	if err := s.rates.Upsert(ctx, from, to, time.Now().UTC(), rate, source); err != nil {
		return err
	}
	s.log.Info("exchange rate snapshot", "from", from.String(), "to", to.String(), "rate", rate, "source", source)
	return nil
}

// LatestExchangeRate returns the most recent rate for a currency pair.
// Resolution order: a stored snapshot dated today, then the public
// APIs (SUNAT first for USD→PEN), then — for USD→PEN only — the
// configured fallback rate. Non-USD→PEN pairs that cannot be resolved
// live error out.
func (s *TreasuryService) LatestExchangeRate(ctx context.Context, from, to valueobjects.CurrencyCode) (ExchangeRateInfo, error) {
	snap, dbErr := s.rates.Latest(ctx, from, to)
	if dbErr == nil && !strings.EqualFold(snap.Source, SourceFallback) && sameDay(snap.RateDate, time.Now().UTC()) {
		return ExchangeRateInfo{Rate: snap.Rate, Source: snap.Source}, nil
	}
	if dbErr != nil && !errors.Is(dbErr, repositories.ErrNotFound) {
		s.log.Warn("exchange rate lookup failed", "from", from.String(), "to", to.String(), "error", dbErr)
	}

	rateStr, source, liveErr := s.live.Fetch(ctx, from.String(), to.String())
	if liveErr == nil {
		parsed, perr := valueobjects.MoneyFromString(rateStr)
		if perr == nil {
			if uerr := s.rates.Upsert(ctx, from, to, time.Now().UTC(), parsed, source); uerr != nil {
				s.log.Warn("could not cache live exchange rate", "error", uerr)
			}
			return ExchangeRateInfo{Rate: parsed, Source: source}, nil
		}
		liveErr = perr
	}

	isUSDPEN := strings.EqualFold(from.String(), "USD") && strings.EqualFold(to.String(), "PEN")
	if !isUSDPEN {
		return ExchangeRateInfo{}, fmt.Errorf("%w: no exchange rate available for %s->%s: %v",
			repositories.ErrNotFound, from.String(), to.String(), liveErr)
	}
	rate := s.fallbackRate(ctx)
	info, ierr := valueobjects.MoneyFromString(strconv.FormatFloat(rate, 'f', 2, 64))
	if ierr != nil {
		return ExchangeRateInfo{}, ierr
	}
	s.log.Warn("live exchange rate fetch failed, using fallback", "from", from.String(), "to", to.String(), "error", liveErr)
	return ExchangeRateInfo{Rate: info, Source: SourceFallback, IsFallback: true}, nil
}

// fallbackRate resolves the preference-configured rate, defaulting to
// DefaultFallbackUSDPEN on any failure.
func (s *TreasuryService) fallbackRate(ctx context.Context) float64 {
	if s.fallback != nil {
		return s.fallback(ctx)
	}
	return DefaultFallbackUSDPEN
}

// sameDay compares two instants by UTC calendar date.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}
