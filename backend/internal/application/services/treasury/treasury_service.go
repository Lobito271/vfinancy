// Package treasury implements the business logic for the treasury
// module: bank accounts, credit cards, transactions, reconciliation,
// and exchange rates.
package treasury

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/treasury"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// TreasuryService owns the bank / card / exchange workflows.
type TreasuryService struct {
	accounts     repositories.BankAccountRepository
	cards        repositories.CreditCardRepository
	transactions repositories.BankTransactionRepository
	rates        repositories.ExchangeRateRepository
	txm          services.TxManager
	log          *common.Logger
}

// New returns a TreasuryService ready for use.
func New(
	accounts repositories.BankAccountRepository,
	cards repositories.CreditCardRepository,
	transactions repositories.BankTransactionRepository,
	rates repositories.ExchangeRateRepository,
	txm services.TxManager,
	log *common.Logger,
) *TreasuryService {
	return &TreasuryService{
		accounts:     accounts,
		cards:        cards,
		transactions: transactions,
		rates:        rates,
		txm:          txm,
		log:          log,
	}
}

// OpenAccountInput creates a new bank account.
type OpenAccountInput struct {
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	BankName      string
	AccountNumber string
	AccountType   string // "checking" | "savings"
	CurrencyCode  valueobjects.CurrencyCode
	GLAccountID   uuid.UUID
	IsDefault     bool
}

// OpenAccount persists a new bank account with zero opening balance.
func (s *TreasuryService) OpenAccount(ctx context.Context, in OpenAccountInput) (*treasury.BankAccount, error) {
	var out *treasury.BankAccount
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		acc, err := treasury.NewBankAccount(time.Now().UTC(), treasury.NewBankAccountOptions{
			CompanyID:     in.CompanyID,
			BranchID:      in.BranchID,
			BankName:      in.BankName,
			AccountNumber: in.AccountNumber,
			AccountType:   in.AccountType,
			CurrencyCode:  in.CurrencyCode,
			GLAccountID:   in.GLAccountID,
			IsDefault:     in.IsDefault,
		})
		if err != nil {
			return err
		}
		if err := uow.BankAccounts().Create(ctx, acc); err != nil {
			return err
		}
		out = acc
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("bank account opened", "account_id", out.ID, "bank", out.BankName)
	return out, nil
}

// RegisterTransactionInput records a movement on a bank account.
type RegisterTransactionInput struct {
	BankAccountID  uuid.UUID
	TransactionDate time.Time
	ValueDate      time.Time
	Description    string
	Amount         valueobjects.Money
	Type           string
	Reference      string
}

// RegisterTransaction appends a movement to the account and updates
// the running balance. Reconciliation is a separate operation.
func (s *TreasuryService) RegisterTransaction(ctx context.Context, in RegisterTransactionInput) (*treasury.BankTransaction, error) {
	var out *treasury.BankTransaction
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		acc, err := uow.BankAccounts().GetByID(ctx, in.BankAccountID)
		if err != nil {
			return err
		}
		acc.ApplyDelta(in.Amount)
		if err := uow.BankAccounts().Update(ctx, acc); err != nil {
			return err
		}
		// The bank_transactions table does not exist in the current
		// migration set; the stub returns ErrNotFound. We surface that
		// to the caller rather than silently dropping the transaction.
		return fmt.Errorf("repositories: %w", repositories.ErrNotFound)
	})
	if err != nil {
		return nil, err
	}
	_ = out
	return nil, nil
}

// ReconcileTransaction marks a transaction as reconciled.
func (s *TreasuryService) ReconcileTransaction(ctx context.Context, id uuid.UUID, by uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		// Stub: BankTransactionRepository.Reconcile isn't implemented
		// in Phase 1.3. We expose the intent here so the service
		// surface is complete.
		_ = uow
		return repositories.ErrNotFound
	})
	if err != nil {
		return err
	}
	s.log.Info("bank transaction reconciled", "transaction_id", id, "by", by)
	return nil
}

// IssueCardInput creates a company-issued credit card.
type IssueCardInput struct {
	CompanyID       uuid.UUID
	BranchID        *uuid.UUID
	Issuer          string
	LastFour        string
	CardHolder      string
	ExpirationMonth int
	ExpirationYear  int
	CreditLimit     valueobjects.Money
	CutOffDay       int
	PaymentDueDay   int
	CurrencyCode    valueobjects.CurrencyCode
	GLAccountID     uuid.UUID
}

// IssueCard creates a new card with zero opening balance.
func (s *TreasuryService) IssueCard(ctx context.Context, in IssueCardInput) (*treasury.CreditCard, error) {
	var out *treasury.CreditCard
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		card, err := treasury.NewCreditCard(time.Now().UTC(), treasury.NewCreditCardOptions{
			CompanyID:       in.CompanyID,
			BranchID:        in.BranchID,
			Issuer:          in.Issuer,
			LastFour:        in.LastFour,
			CardHolder:      in.CardHolder,
			ExpirationMonth: in.ExpirationMonth,
			ExpirationYear:  in.ExpirationYear,
			CreditLimit:     in.CreditLimit,
			CutOffDay:       in.CutOffDay,
			PaymentDueDay:   in.PaymentDueDay,
			CurrencyCode:    in.CurrencyCode,
			GLAccountID:     in.GLAccountID,
		})
		if err != nil {
			return err
		}
		if err := uow.CreditCards().Create(ctx, card); err != nil {
			return err
		}
		out = card
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("credit card issued", "card_id", out.ID, "issuer", out.Issuer)
	return out, nil
}

// ChargeCardInput records a purchase on a credit card. The amount is
// added to the card's outstanding balance; the application use case is
// expected to record the corresponding journal entry.
func (s *TreasuryService) ChargeCard(ctx context.Context, cardID uuid.UUID, amount valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		card, err := uow.CreditCards().GetByID(ctx, cardID)
		if err != nil {
			return err
		}
		if err := card.Charge(amount); err != nil {
			return err
		}
		return uow.CreditCards().Update(ctx, card)
	})
	if err != nil {
		return err
	}
	s.log.Info("credit card charged", "card_id", cardID, "amount", amount)
	return nil
}

// PayCard records a payment against a credit card.
func (s *TreasuryService) PayCard(ctx context.Context, cardID uuid.UUID, amount valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		card, err := uow.CreditCards().GetByID(ctx, cardID)
		if err != nil {
			return err
		}
		if err := card.Pay(amount); err != nil {
			return err
		}
		return uow.CreditCards().Update(ctx, card)
	})
	if err != nil {
		return err
	}
	s.log.Info("credit card payment", "card_id", cardID, "amount", amount)
	return nil
}

// UpsertExchangeRateInput creates or updates the (from, to, date) rate.
type UpsertExchangeRateInput struct {
	From          valueobjects.CurrencyCode
	To            valueobjects.CurrencyCode
	Rate          valueobjects.Money
	EffectiveDate time.Time
	Source        string
}

// UpsertExchangeRate records a rate snapshot. The application use case
// is expected to source rates from BCRP / SUNAT; this is the manual
// override path.
func (s *TreasuryService) UpsertExchangeRate(ctx context.Context, in UpsertExchangeRateInput) error {
	date := in.EffectiveDate.Format("2006-01-02")
	err := s.rates.Upsert(ctx, in.From.String(), in.To.String(), in.Rate.String(), date, in.Source)
	if err != nil {
		return err
	}
	s.log.Info("exchange rate upserted",
		"from", in.From, "to", in.To,
		"date", date, "rate", in.Rate,
	)
	return nil
}

// LatestExchangeRate returns the most recent rate for a currency pair.
// Used by the application's use case to convert transactional
// amounts to the company's functional currency.
func (s *TreasuryService) LatestExchangeRate(ctx context.Context, from, to valueobjects.CurrencyCode) (valueobjects.Money, error) {
	rate, err := s.rates.GetLatest(ctx, from.String(), to.String())
	if err != nil {
		return valueobjects.Money{}, err
	}
	return valueobjects.MoneyFromString(rate)
}
