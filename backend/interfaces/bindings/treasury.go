package bindings

import (
	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/treasury"
	"vfinancy/backend/internal/utils"
)

// CreditCardDTO is the serializable view of a credit card.
type CreditCardDTO struct {
	ID              string `json:"id"`
	Issuer          string `json:"issuer"`
	LastFour        string `json:"lastFour"`
	CardHolder      string `json:"cardHolder"`
	ExpirationMonth int    `json:"expirationMonth"`
	ExpirationYear  int    `json:"expirationYear"`
	CreditLimit     string `json:"creditLimit"`
	CurrentBalance  string `json:"currentBalance"`
	AvailableCredit string `json:"availableCredit"`
	CutOffDay       int    `json:"cutOffDay"`
	PaymentDueDay   int    `json:"paymentDueDay"`
	CurrencyCode    string `json:"currencyCode"`
	IsActive        bool   `json:"isActive"`
}

func toCreditCardDTO(c *treasury.CreditCard) *CreditCardDTO {
	return &CreditCardDTO{
		ID: c.ID.String(), Issuer: c.Issuer, LastFour: c.LastFour, CardHolder: c.CardHolder,
		ExpirationMonth: c.ExpirationMonth, ExpirationYear: c.ExpirationYear,
		CreditLimit: c.CreditLimit.String(), CurrentBalance: c.CurrentBalance.String(),
		AvailableCredit: c.AvailableCredit().String(), CutOffDay: c.CutOffDay,
		PaymentDueDay: c.PaymentDueDay, CurrencyCode: c.CurrencyCode.String(), IsActive: c.IsActive,
	}
}

func (a *App) ListCreditCards() ([]*CreditCardDTO, error) {
	cards, err := a.treasurySvc.ListCards(a.Context(), a.companyID())
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	result := make([]*CreditCardDTO, 0, len(cards))
	for _, card := range cards {
		result = append(result, toCreditCardDTO(card))
	}
	return result, nil
}

type IssueCreditCardRequest struct {
	Issuer          string `json:"issuer"`
	LastFour        string `json:"lastFour"`
	CardHolder      string `json:"cardHolder"`
	ExpirationMonth int    `json:"expirationMonth"`
	ExpirationYear  int    `json:"expirationYear"`
	CreditLimit     string `json:"creditLimit"`
	CutOffDay       int    `json:"cutOffDay"`
	PaymentDueDay   int    `json:"paymentDueDay"`
	CurrencyCode    string `json:"currencyCode"`
}

func (a *App) IssueCreditCard(req IssueCreditCardRequest) (*CreditCardDTO, error) {
	limit, err := valueobjects.MoneyFromString(req.CreditLimit)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	currency, err := valueobjects.NewCurrencyCode(req.CurrencyCode)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	card, err := a.treasurySvc.IssueCard(a.Context(), treasury.IssueCardInput{
		CompanyID: a.companyID(), Issuer: req.Issuer, LastFour: req.LastFour, CardHolder: req.CardHolder,
		ExpirationMonth: req.ExpirationMonth, ExpirationYear: req.ExpirationYear, CreditLimit: limit,
		CutOffDay: req.CutOffDay, PaymentDueDay: req.PaymentDueDay, CurrencyCode: currency,
	})
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toCreditCardDTO(card), nil
}

type UpdateCreditCardRequest struct {
	ID            string `json:"id"`
	Issuer        string `json:"issuer"`
	LastFour      string `json:"lastFour"`
	CardHolder    string `json:"cardHolder"`
	CreditLimit   string `json:"creditLimit"`
	CutOffDay     int    `json:"cutOffDay"`
	PaymentDueDay int    `json:"paymentDueDay"`
	IsActive      bool   `json:"isActive"`
}

func (a *App) UpdateCreditCard(req UpdateCreditCardRequest) (*CreditCardDTO, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	limit, err := valueobjects.MoneyFromString(req.CreditLimit)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	card, err := a.treasurySvc.UpdateCard(a.Context(), treasury.UpdateCardInput{
		ID:            id,
		Issuer:        req.Issuer,
		LastFour:      req.LastFour,
		CardHolder:    req.CardHolder,
		CreditLimit:   limit,
		CutOffDay:     req.CutOffDay,
		PaymentDueDay: req.PaymentDueDay,
		IsActive:      req.IsActive,
	})
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toCreditCardDTO(card), nil
}

func (a *App) DeleteCreditCard(id string) error {
	cardID, err := uuid.Parse(id)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.treasurySvc.DeleteCard(a.Context(), cardID))
}

type CreditCardAmountRequest struct {
	ID     string `json:"id"`
	Amount string `json:"amount"`
}

func (a *App) ChargeCreditCard(req CreditCardAmountRequest) error {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return utils.ProcessError(err)
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.treasurySvc.ChargeCard(a.Context(), id, amount))
}

// LatestExchangeRate returns the most recent rate for a currency pair.
func (a *App) LatestExchangeRate(from, to string) (string, error) {
	f, err := valueobjects.NewCurrencyCode(from)
	if err != nil {
		return "", utils.ProcessError(err)
	}
	t, err := valueobjects.NewCurrencyCode(to)
	if err != nil {
		return "", utils.ProcessError(err)
	}
	rate, err := a.treasurySvc.LatestExchangeRate(a.Context(), f, t)
	if err != nil {
		return "", utils.ProcessError(err)
	}
	return rate.String(), nil
}

// CardProjectionDTO is the serializable view of a credit card payment projection.
type CardProjectionDTO struct {
	CardID          string  `json:"cardId"`
	Issuer          string  `json:"issuer"`
	LastFour        string  `json:"lastFour"`
	CardHolder      string  `json:"cardHolder"`
	ProjectedUSD    float64 `json:"projectedUSD"`
	CycleStart      string  `json:"cycleStart"`
	NextCutOffDate  string  `json:"nextCutOffDate"`
	NextPaymentDate string  `json:"nextPaymentDate"`
}

func toCardProjectionDTO(p treasury.CardPaymentProjection) *CardProjectionDTO {
	return &CardProjectionDTO{
		CardID:          p.CardID,
		Issuer:          p.Issuer,
		LastFour:        p.LastFour,
		CardHolder:      p.CardHolder,
		ProjectedUSD:    p.ProjectedUSD,
		CycleStart:      p.CycleStart,
		NextCutOffDate:  p.NextCutOffDate,
		NextPaymentDate: p.NextPaymentDate,
	}
}

// GetCardProjections returns the projected USD debt for each credit card.
func (a *App) GetCardProjections() ([]*CardProjectionDTO, error) {
	projections, err := a.treasurySvc.ProjectPayments(a.Context(), a.companyID())
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	items := make([]*CardProjectionDTO, 0, len(projections))
	for _, p := range projections {
		items = append(items, toCardProjectionDTO(p))
	}
	return items, nil
}

// PayCreditCardRequest records a payment against a credit card.
type PayCreditCardRequest struct {
	CardID string `json:"cardId"`
	Amount string `json:"amount"`
}

// PayCreditCard pays the given amount against a credit card, reducing
// its outstanding balance.
func (a *App) PayCreditCard(req PayCreditCardRequest) error {
	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		return utils.ProcessError(err)
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.treasurySvc.PayCard(a.Context(), cardID, amount))
}
