// Package treasury implements the business logic for the treasury
// module: bank accounts, credit cards, transactions, reconciliation,
// and exchange rates.
package treasury

import (
	"context"
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/shared/logger"
)

// TreasuryService owns the bank / card / exchange workflows.
type TreasuryService struct {
	accounts     BankAccountRepository
	cards        CreditCardRepository
	transactions BankTransactionRepository
	rates        ExchangeRateRepository
	txm          repositories.TransactionManager
	log          *logger.Logger
}

// New returns a TreasuryService ready for use.
func New(
	accounts BankAccountRepository,
	cards CreditCardRepository,
	transactions BankTransactionRepository,
	rates ExchangeRateRepository,
	txm repositories.TransactionManager,
	log *logger.Logger,
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
func (s *TreasuryService) OpenAccount(ctx context.Context, in OpenAccountInput) (*BankAccount, error) {
	var out *BankAccount
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		acc, err := NewBankAccount(time.Now().UTC(), NewBankAccountOptions{
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
		if err := s.accounts.Create(ctx, acc); err != nil {
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

// UpdateAccountInput describes the editable fields of a bank account.
type UpdateAccountInput struct {
	ID            uuid.UUID
	BankName      string
	AccountNumber string
	AccountType   string
	CurrencyCode  valueobjects.CurrencyCode
	IsDefault     bool
	IsActive      bool
}

// UpdateAccount persists changes to a bank account.
func (s *TreasuryService) UpdateAccount(ctx context.Context, in UpdateAccountInput) (*BankAccount, error) {
	var out *BankAccount
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		acc, err := s.accounts.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if in.BankName == "" {
			return derrors.Wrap(derrors.ErrRequired, errField("bank name is required"))
		}
		if in.AccountNumber == "" {
			return derrors.Wrap(derrors.ErrRequired, errField("account number is required"))
		}
		if in.CurrencyCode.IsZero() {
			return derrors.Wrap(derrors.ErrRequired, errField("currency is required"))
		}
		acc.BankName = in.BankName
		acc.AccountNumber = in.AccountNumber
		acc.AccountType = in.AccountType
		acc.CurrencyCode = in.CurrencyCode
		acc.IsDefault = in.IsDefault
		if in.IsActive {
			acc.Activate()
		} else {
			acc.Deactivate()
		}
		if err := s.accounts.Update(ctx, acc); err != nil {
			return err
		}
		out = acc
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("bank account updated", "account_id", out.ID, "bank", out.BankName)
	return out, nil
}

// DeleteAccount soft-deletes a bank account.
func (s *TreasuryService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.accounts.Delete(ctx, id)
	})
	if err != nil {
		return err
	}
	s.log.Info("bank account deleted", "account_id", id)
	return nil
}

// RegisterTransactionInput creates a bank transaction.
type RegisterTransactionInput struct {
	BankAccountID uuid.UUID
	Date          time.Time
	Description   string
	Amount        valueobjects.Money
	Type          enums.BankTransactionType
	Reference     string
}

// RegisterTransaction records a movement against a bank account and
// updates the account balance in the same transaction.
func (s *TreasuryService) RegisterTransaction(ctx context.Context, in RegisterTransactionInput) (*BankTransaction, error) {
	if !in.Type.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("transaction type is invalid"))
	}
	if !in.Amount.IsPositive() {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("amount must be greater than zero"))
	}
	var out *BankTransaction
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		acc, err := s.accounts.GetByID(ctx, in.BankAccountID)
		if err != nil {
			return err
		}
		delta := in.Amount
		switch in.Type {
		case enums.BankTxTypeWithdrawal, enums.BankTxTypeFee, enums.BankTxTypeTransfer:
			delta = delta.Neg()
		}
		acc.ApplyDelta(delta)
		if err := s.accounts.Update(ctx, acc); err != nil {
			return err
		}
		now := time.Now().UTC()
		t := &BankTransaction{
			ID:              uuid.New(),
			BankAccountID:   in.BankAccountID,
			TransactionDate: in.Date,
			Description:     in.Description,
			Amount:          in.Amount,
			Type:            in.Type,
			Reference:       in.Reference,
			BalanceAfter:    acc.CurrentBalance,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.transactions.Create(ctx, t); err != nil {
			return err
		}
		out = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("bank transaction registered",
		"transaction_id", out.ID, "account_id", in.BankAccountID,
		"amount", in.Amount, "type", in.Type,
	)
	return out, nil
}

// ReconcileTransactionInput is a placeholder for the reconciliation
// workflow, which is scheduled for a later phase once the
// bank_transactions schema is in place.

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
func (s *TreasuryService) IssueCard(ctx context.Context, in IssueCardInput) (*CreditCard, error) {
	var out *CreditCard
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		card, err := NewCreditCard(time.Now().UTC(), NewCreditCardOptions{
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
		if err := s.cards.Create(ctx, card); err != nil {
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
// added to the card's outstanding balance; the caller is expected to
// record the corresponding journal entry.
func (s *TreasuryService) ChargeCard(ctx context.Context, cardID uuid.UUID, amount valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		card, err := s.cards.GetByID(ctx, cardID)
		if err != nil {
			return err
		}
		if err := card.Charge(amount); err != nil {
			return err
		}
		return s.cards.Update(ctx, card)
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
		card, err := s.cards.GetByID(ctx, cardID)
		if err != nil {
			return err
		}
		if err := card.Pay(amount); err != nil {
			return err
		}
		return s.cards.Update(ctx, card)
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

// UpsertExchangeRate records a rate snapshot. This is the manual
// override path; automated rates (BCRP / SUNAT) are sourced by the
// treasury slice.
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
// Used by callers to convert transactional amounts to the company's
// functional currency.
func (s *TreasuryService) LatestExchangeRate(ctx context.Context, from, to valueobjects.CurrencyCode) (valueobjects.Money, error) {
	rate, err := s.rates.GetLatest(ctx, from.String(), to.String())
	if err != nil {
		return valueobjects.Money{}, err
	}
	return valueobjects.MoneyFromString(rate)
}

// ListAccounts returns bank accounts matching the filter.
func (s *TreasuryService) ListAccounts(ctx context.Context, filter BankAccountFilter) (repositories.Page[*BankAccount], error) {
	return s.accounts.List(ctx, filter)
}

// GetAccount returns a single bank account by ID.
func (s *TreasuryService) GetAccount(ctx context.Context, id uuid.UUID) (*BankAccount, error) {
	return s.accounts.GetByID(ctx, id)
}

// ListTransactions returns bank transactions matching the filter.
func (s *TreasuryService) ListTransactions(ctx context.Context, filter BankTransactionFilter) (repositories.Page[*BankTransaction], error) {
	return s.transactions.List(ctx, filter)
}

// GetTransaction returns a single bank transaction by ID.
func (s *TreasuryService) GetTransaction(ctx context.Context, id uuid.UUID) (*BankTransaction, error) {
	return s.transactions.GetByID(ctx, id)
}

// MarkTransactionReconciled flags a bank transaction as reconciled.
func (s *TreasuryService) MarkTransactionReconciled(ctx context.Context, id uuid.UUID) (*BankTransaction, error) {
	var out *BankTransaction
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		t, err := s.transactions.GetByID(ctx, id)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		t.IsReconciled = true
		t.ReconciledAt = &now
		if err := s.transactions.Update(ctx, t); err != nil {
			return err
		}
		out = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("bank transaction reconciled", "transaction_id", id)
	return out, nil
}
