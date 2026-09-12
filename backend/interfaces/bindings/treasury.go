package bindings

import (
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/treasury"
)

type CreditCardDTO struct {
	ID             string  `json:"id"`
	Issuer         string  `json:"issuer"`
	LastFour       string  `json:"lastFour"`
	CreditLimit    float64 `json:"creditLimit"`
	CurrentBalance float64 `json:"currentBalance"`
	CutOffDay      int     `json:"cutOffDay"`
	PaymentDueDay  int     `json:"paymentDueDay"`
	IsActive       bool    `json:"isActive"`
}

func cardDTO(c *treasury.CreditCard) CreditCardDTO {
	return CreditCardDTO{
		ID:             c.ID.String(),
		Issuer:         c.Issuer,
		LastFour:       c.LastFour,
		CreditLimit:    moneyFloat(c.CreditLimit),
		CurrentBalance: moneyFloat(c.CurrentBalance),
		CutOffDay:      c.CutOffDay,
		PaymentDueDay:  c.PaymentDueDay,
		IsActive:       c.IsActive,
	}
}

// ListCreditCards returns the treasury card catalog.
func (a *App) ListCreditCards() ([]CreditCardDTO, error) {
	cards, err := a.treasurySvc.ListCards(a.Context())
	if err != nil {
		return nil, err
	}
	items := make([]CreditCardDTO, 0, len(cards))
	for _, c := range cards {
		items = append(items, cardDTO(c))
	}
	return items, nil
}

type SaveCreditCardRequest struct {
	ID            string  `json:"id"`
	Issuer        string  `json:"issuer"`
	LastFour      string  `json:"lastFour"`
	CreditLimit   float64 `json:"creditLimit"`
	CutOffDay     int     `json:"cutOffDay"`
	PaymentDueDay int     `json:"paymentDueDay"`
}

// IssueCreditCard registers a new international card (USD).
func (a *App) IssueCreditCard(req SaveCreditCardRequest) (CreditCardDTO, error) {
	limit, err := moneyFromFloat(req.CreditLimit)
	if err != nil {
		return CreditCardDTO{}, err
	}
	c, err := a.treasurySvc.IssueCard(a.Context(), treasury.IssueCardInput{
		Issuer:        req.Issuer,
		LastFour:      req.LastFour,
		CreditLimit:   limit,
		CutOffDay:     req.CutOffDay,
		PaymentDueDay: req.PaymentDueDay,
	})
	if err != nil {
		return CreditCardDTO{}, err
	}
	return cardDTO(c), nil
}

// UpdateCreditCard edits the card parameters.
func (a *App) UpdateCreditCard(req SaveCreditCardRequest) (CreditCardDTO, error) {
	id, err := parseUUID(req.ID)
	if err != nil {
		return CreditCardDTO{}, err
	}
	limit, err := moneyPtrFromFloat(req.CreditLimit)
	if err != nil {
		return CreditCardDTO{}, err
	}
	c, err := a.treasurySvc.UpdateCard(a.Context(), treasury.UpdateCardInput{
		ID:            id,
		Issuer:        &req.Issuer,
		LastFour:      &req.LastFour,
		CreditLimit:   limit,
		CutOffDay:     &req.CutOffDay,
		PaymentDueDay: &req.PaymentDueDay,
	})
	if err != nil {
		return CreditCardDTO{}, err
	}
	return cardDTO(c), nil
}

// DeleteCreditCard soft-deletes a card preserving the historical
// projections (edge case 4.4).
func (a *App) DeleteCreditCard(id string) error {
	cid, err := parseUUID(id)
	if err != nil {
		return err
	}
	return a.treasurySvc.DeleteCard(a.Context(), cid)
}

type CardPaymentRequest struct {
	CardID string  `json:"cardId"`
	Amount float64 `json:"amount"`
}

// PayCreditCard liquidates the active cycle (full amount resets the
// balance to zero).
func (a *App) PayCreditCard(req CardPaymentRequest) (CreditCardDTO, error) {
	cid, err := parseUUID(req.CardID)
	if err != nil {
		return CreditCardDTO{}, err
	}
	amount, err := moneyFromFloat(req.Amount)
	if err != nil {
		return CreditCardDTO{}, err
	}
	if err := a.treasurySvc.PayCard(a.Context(), cid, amount); err != nil {
		return CreditCardDTO{}, err
	}
	c, err := a.treasurySvc.GetCard(a.Context(), cid)
	if err != nil {
		return CreditCardDTO{}, err
	}
	return cardDTO(c), nil
}

type CardProjectionDTO struct {
	CardID       string  `json:"cardId"`
	Issuer       string  `json:"issuer"`
	LastFour     string  `json:"lastFour"`
	CycleStart   string  `json:"cycleStart"`
	CycleEnd     string  `json:"cycleEnd"`
	PaymentDue   string  `json:"paymentDue"`
	TotalUSD     float64 `json:"totalUsd"`
	RefundsUSD   float64 `json:"refundsUsd"`
	Status       string  `json:"status"`
	CurrentBalance float64 `json:"currentBalance"`
}

// GetCardProjections returns the active billing cycle per card with the
// month-end clamping applied (edge case 4.4).
func (a *App) GetCardProjections() ([]CardProjectionDTO, error) {
	projections, err := a.treasurySvc.ProjectPayments(a.Context())
	if err != nil {
		return nil, err
	}
	items := make([]CardProjectionDTO, 0, len(projections))
	for _, p := range projections {
		items = append(items, CardProjectionDTO{
			CardID:         p.Card.ID.String(),
			Issuer:         p.Card.Issuer,
			LastFour:       p.Card.LastFour,
			CycleStart:     dayStr(p.CycleStart),
			CycleEnd:       dayStr(p.CycleEnd),
			PaymentDue:     dayStr(p.PaymentDue),
			TotalUSD:       moneyFloat(p.TotalUSD),
			RefundsUSD:     moneyFloat(p.RefundsUSD),
			Status:         p.Status,
			CurrentBalance: moneyFloat(p.Card.CurrentBalance),
		})
	}
	return items, nil
}

type ExchangeRateDTO struct {
	Rate       float64 `json:"rate"`
	Source     string  `json:"source"`
	IsFallback bool    `json:"isFallback"`
}

// LatestExchangeRate returns the USD/PEN rate with its source so the UI
// can surface the "Modo Contingencia" indicator when the API failed
// (edge case 4.1).
func (a *App) LatestExchangeRate() (ExchangeRateDTO, error) {
	usd, err := valueobjects.NewCurrencyCode("USD")
	if err != nil {
		return ExchangeRateDTO{}, err
	}
	pen, err := valueobjects.NewCurrencyCode("PEN")
	if err != nil {
		return ExchangeRateDTO{}, err
	}
	info, err := a.treasurySvc.LatestExchangeRate(a.Context(), usd, pen)
	if err != nil {
		return ExchangeRateDTO{}, err
	}
	return ExchangeRateDTO{Rate: moneyFloat(info.Rate), Source: info.Source, IsFallback: info.IsFallback}, nil
}
