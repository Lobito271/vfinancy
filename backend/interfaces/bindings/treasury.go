package bindings

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/accounting"
	"vfinancy/backend/internal/features/treasury"
)

// BankAccountDTO is the serializable view of a bank account.
type BankAccountDTO struct {
	ID             string `json:"id"`
	BankName       string `json:"bankName"`
	AccountNumber  string `json:"accountNumber"`
	AccountType    string `json:"accountType"`
	CurrencyCode   string `json:"currencyCode"`
	CurrentBalance string `json:"currentBalance"`
	IsDefault      bool   `json:"isDefault"`
	IsActive       bool   `json:"isActive"`
}

// BankTransactionDTO is the serializable view of a bank transaction.
type BankTransactionDTO struct {
	ID           string `json:"id"`
	AccountID    string `json:"accountId"`
	Date         string `json:"date"`
	Description  string `json:"description"`
	Amount       string `json:"amount"`
	Type         string `json:"type"`
	BalanceAfter string `json:"balanceAfter"`
	Reference    string `json:"reference"`
	IsReconciled bool   `json:"isReconciled"`
}

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

func toBankAccountDTO(a *treasury.BankAccount) *BankAccountDTO {
	return &BankAccountDTO{
		ID:             a.ID.String(),
		BankName:       a.BankName,
		AccountNumber:  a.AccountNumber,
		AccountType:    a.AccountType,
		CurrencyCode:   a.CurrencyCode.String(),
		CurrentBalance: a.CurrentBalance.String(),
		IsDefault:      a.IsDefault,
		IsActive:       a.IsActive,
	}
}

func toBankTransactionDTO(t *treasury.BankTransaction) *BankTransactionDTO {
	return &BankTransactionDTO{
		ID:           t.ID.String(),
		AccountID:    t.BankAccountID.String(),
		Date:         t.TransactionDate.Format(time.RFC3339),
		Description:  t.Description,
		Amount:       t.Amount.String(),
		Type:         t.Type.String(),
		BalanceAfter: t.BalanceAfter.String(),
		Reference:    t.Reference,
		IsReconciled: t.IsReconciled,
	}
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
		return nil, err
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
		return nil, err
	}
	currency, err := valueobjects.NewCurrencyCode(req.CurrencyCode)
	if err != nil {
		return nil, err
	}
	gl, err := a.ensureCardGLAccount(a.Context())
	if err != nil {
		return nil, err
	}
	card, err := a.treasurySvc.IssueCard(a.Context(), treasury.IssueCardInput{
		CompanyID: a.companyID(), Issuer: req.Issuer, LastFour: req.LastFour, CardHolder: req.CardHolder,
		ExpirationMonth: req.ExpirationMonth, ExpirationYear: req.ExpirationYear, CreditLimit: limit,
		CutOffDay: req.CutOffDay, PaymentDueDay: req.PaymentDueDay, CurrencyCode: currency, GLAccountID: gl,
	})
	if err != nil {
		return nil, err
	}
	return toCreditCardDTO(card), nil
}

type CreditCardAmountRequest struct {
	ID     string `json:"id"`
	Amount string `json:"amount"`
}

func (a *App) ChargeCreditCard(req CreditCardAmountRequest) error {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return err
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return err
	}
	return a.treasurySvc.ChargeCard(a.Context(), id, amount)
}

// ListBankAccounts returns all active bank accounts.
func (a *App) ListBankAccounts() ([]*BankAccountDTO, error) {
	ctx := a.Context()
	filter := treasury.BankAccountFilter{
		CompanyID: a.companyIDPtr(),
	}
	page, err := a.treasurySvc.ListAccounts(ctx, filter)
	if err != nil {
		return nil, err
	}
	items := make([]*BankAccountDTO, 0, len(page.Items))
	for _, acc := range page.Items {
		items = append(items, toBankAccountDTO(acc))
	}
	return items, nil
}

// GetBankAccount returns a single bank account.
func (a *App) GetBankAccount(id string) (*BankAccountDTO, error) {
	aid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	acc, err := a.treasurySvc.GetAccount(a.Context(), aid)
	if err != nil {
		return nil, err
	}
	return toBankAccountDTO(acc), nil
}

// CreateBankAccountRequest creates a bank account.
type CreateBankAccountRequest struct {
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountType   string `json:"accountType"`
	CurrencyCode  string `json:"currencyCode"`
	IsDefault     bool   `json:"isDefault"`
}

// CreateBankAccount opens a new bank account, resolving the linked GL
// account automatically (the standard 104.01 bank account is created
// on first use when the chart of accounts is empty).
func (a *App) CreateBankAccount(req CreateBankAccountRequest) (*BankAccountDTO, error) {
	ctx := a.Context()
	glAccountID, err := a.ensureBankGLAccount(ctx)
	if err != nil {
		return nil, err
	}
	currencyCode, err := valueobjects.NewCurrencyCode(req.CurrencyCode)
	if err != nil {
		return nil, err
	}
	acc, err := a.treasurySvc.OpenAccount(ctx, treasury.OpenAccountInput{
		CompanyID:     a.companyID(),
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountType:   req.AccountType,
		CurrencyCode:  currencyCode,
		GLAccountID:   glAccountID,
		IsDefault:     req.IsDefault,
	})
	if err != nil {
		return nil, err
	}
	return toBankAccountDTO(acc), nil
}

// UpdateBankAccountRequest updates a bank account.
type UpdateBankAccountRequest struct {
	ID            string `json:"id"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountType   string `json:"accountType"`
	CurrencyCode  string `json:"currencyCode"`
	IsDefault     bool   `json:"isDefault"`
	IsActive      bool   `json:"isActive"`
}

// UpdateBankAccount persists changes to a bank account.
func (a *App) UpdateBankAccount(req UpdateBankAccountRequest) (*BankAccountDTO, error) {
	ctx := a.Context()
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	currencyCode, err := valueobjects.NewCurrencyCode(req.CurrencyCode)
	if err != nil {
		return nil, err
	}
	acc, err := a.treasurySvc.UpdateAccount(ctx, treasury.UpdateAccountInput{
		ID:            id,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountType:   req.AccountType,
		CurrencyCode:  currencyCode,
		IsDefault:     req.IsDefault,
		IsActive:      req.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return toBankAccountDTO(acc), nil
}

// DeleteBankAccount soft-deletes a bank account.
func (a *App) DeleteBankAccount(id string) error {
	aid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return a.treasurySvc.DeleteAccount(a.Context(), aid)
}

// ensureBankGLAccount returns the GL account used for bank movements,
// creating the standard 104.01 current account when the chart of
// accounts is empty during initial setup.
func (a *App) ensureBankGLAccount(ctx context.Context) (uuid.UUID, error) {
	accounts, err := a.accountingSvc.ListChartOfAccounts(ctx, a.companyID())
	if err != nil {
		return uuid.Nil, err
	}
	for _, acc := range accounts {
		if strings.HasPrefix(acc.Code.String(), "104") {
			return acc.ID, nil
		}
	}
	code, err := valueobjects.NewChartOfAccountsCode("104.01")
	if err != nil {
		return uuid.Nil, err
	}
	acc, err := a.accountingSvc.CreateChartOfAccounts(ctx, accounting.CreateChartOfAccountsInput{
		CompanyID:      a.companyID(),
		Code:           code,
		Name:           "Cuentas corrientes en instituciones financieras",
		Type:           enums.AccountTypeAsset,
		Depth:          2,
		AllowsMovement: true,
		Description:    "Banco y dinero en cuentas corrientes (creada automáticamente por Tesorería)",
	})
	if err != nil {
		return uuid.Nil, err
	}
	return acc.ID, nil
}

// ListBankTransactionsRequest lists transactions for an account.
type ListBankTransactionsRequest struct {
	AccountID  string `json:"accountId"`
	Reconciled *bool  `json:"reconciled"`
	PaginationRequest
}

// ListBankTransactions returns paged transactions for an account.
func (a *App) ListBankTransactions(req ListBankTransactionsRequest) (PageResult, error) {
	ctx := a.Context()
	filter := treasury.BankTransactionFilter{
		PageRequest: req.toPageRequest(),
	}
	accountID, err := parseOptionalUUID(req.AccountID)
	if err != nil {
		return PageResult{}, err
	}
	filter.BankAccountID = accountID
	filter.Reconciled = req.Reconciled
	page, err := a.treasurySvc.ListTransactions(ctx, filter)
	if err != nil {
		return PageResult{}, err
	}
	items := make([]*BankTransactionDTO, 0, len(page.Items))
	for _, t := range page.Items {
		items = append(items, toBankTransactionDTO(t))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// ReconcileBankTransaction flags a transaction as reconciled.
func (a *App) ReconcileBankTransaction(id string) (*BankTransactionDTO, error) {
	tid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	t, err := a.treasurySvc.MarkTransactionReconciled(a.Context(), tid)
	if err != nil {
		return nil, err
	}
	return toBankTransactionDTO(t), nil
}

// CreateBankTransactionRequest records a movement against an account.
type CreateBankTransactionRequest struct {
	AccountID   string `json:"accountId"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Type        string `json:"type"`
	Reference   string `json:"reference"`
}

// CreateBankTransaction registers a bank transaction and updates the
// account balance in the same transaction.
func (a *App) CreateBankTransaction(req CreateBankTransactionRequest) (*BankTransactionDTO, error) {
	ctx := a.Context()
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, err
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return nil, err
	}
	t, err := a.treasurySvc.RegisterTransaction(ctx, treasury.RegisterTransactionInput{
		BankAccountID: accountID,
		Date:          date,
		Description:   req.Description,
		Amount:        amount,
		Type:          enums.BankTransactionType(req.Type),
		Reference:     req.Reference,
	})
	if err != nil {
		return nil, err
	}
	return toBankTransactionDTO(t), nil
}

// UpsertExchangeRateRequest creates or updates a (from, to, date) rate.
type UpsertExchangeRateRequest struct {
	From          string `json:"from"`
	To            string `json:"to"`
	Rate          string `json:"rate"`
	EffectiveDate string `json:"effectiveDate"`
	Source        string `json:"source"`
}

// UpsertExchangeRate records a manual exchange rate snapshot.
func (a *App) UpsertExchangeRate(req UpsertExchangeRateRequest) error {
	from, err := valueobjects.NewCurrencyCode(req.From)
	if err != nil {
		return err
	}
	to, err := valueobjects.NewCurrencyCode(req.To)
	if err != nil {
		return err
	}
	rate, err := valueobjects.MoneyFromString(req.Rate)
	if err != nil {
		return err
	}
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return err
	}
	source := req.Source
	if source == "" {
		source = "manual"
	}
	return a.treasurySvc.UpsertExchangeRate(a.Context(), treasury.UpsertExchangeRateInput{
		From:          from,
		To:            to,
		Rate:          rate,
		EffectiveDate: effectiveDate,
		Source:        source,
	})
}

// LatestExchangeRate returns the most recent rate for a currency pair.
func (a *App) LatestExchangeRate(from, to string) (string, error) {
	f, err := valueobjects.NewCurrencyCode(from)
	if err != nil {
		return "", err
	}
	t, err := valueobjects.NewCurrencyCode(to)
	if err != nil {
		return "", err
	}
	rate, err := a.treasurySvc.LatestExchangeRate(a.Context(), f, t)
	if err != nil {
		return "", err
	}
	return rate.String(), nil
}

// ensureCardGLAccount returns the GL account used for card payables,
// creating the standard 165.01 when the chart of accounts is empty.
func (a *App) ensureCardGLAccount(ctx context.Context) (uuid.UUID, error) {
	accounts, err := a.accountingSvc.ListChartOfAccounts(ctx, a.companyID())
	if err != nil {
		return uuid.Nil, err
	}
	for _, acc := range accounts {
		if strings.HasPrefix(acc.Code.String(), "165") {
			return acc.ID, nil
		}
	}
	code, err := valueobjects.NewChartOfAccountsCode("165.01")
	if err != nil {
		return uuid.Nil, err
	}
	acc, err := a.accountingSvc.CreateChartOfAccounts(ctx, accounting.CreateChartOfAccountsInput{
		CompanyID:      a.companyID(),
		Code:           code,
		Name:           "Tarjetas de crédito por pagar",
		Type:           enums.AccountTypeLiability,
		Depth:          2,
		AllowsMovement: true,
		Description:    "Cuentas por pagar a tarjetas de crédito corporativas (creada automáticamente por Tesorería)",
	})
	if err != nil {
		return uuid.Nil, err
	}
	return acc.ID, nil
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
		return nil, err
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
		return err
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return err
	}
	return a.treasurySvc.PayCard(a.Context(), cardID, amount)
}
