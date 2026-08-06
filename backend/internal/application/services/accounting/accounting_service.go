// Package accounting implements the business logic for the
// double-entry bookkeeping module: journal entries, posting, ledger
// queries, trial balance, and financial statements.
package accounting

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/accounting"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// AccountingService owns the journal entry workflow plus the report
// queries (trial balance, etc.).
type AccountingService struct {
	entries repositories.JournalRepository
	chart   repositories.ChartOfAccountsRepository
	ledger  repositories.LedgerRepository
	txm     services.TxManager
	log     *common.Logger
}

// New returns an AccountingService ready for use.
func New(
	entries repositories.JournalRepository,
	chart repositories.ChartOfAccountsRepository,
	ledger repositories.LedgerRepository,
	txm services.TxManager,
	log *common.Logger,
) *AccountingService {
	return &AccountingService{
		entries: entries,
		chart:   chart,
		ledger:  ledger,
		txm:     txm,
		log:     log,
	}
}

// EntryInput is the payload for CreateEntry.
type EntryInput struct {
	CompanyID      uuid.UUID
	FiscalPeriodID uuid.UUID
	Number         string
	EntryDate      valueobjects.Date
	PostingDate    *valueobjects.Date
	Description    string
	Source         enums.JournalType
	SourceID       *uuid.UUID
	Lines          []EntryLineInput
}

// EntryLineInput is one line of the journal entry.
type EntryLineInput struct {
	AccountID           uuid.UUID
	Description         string
	Debit               valueobjects.Money
	Credit              valueobjects.Money
	CurrencyCode        valueobjects.CurrencyCode
	ExchangeRate        valueobjects.ExchangeRate
	AmountInTxnCurrency valueobjects.Money
}

// CreateEntry constructs a journal entry from the input, validates it,
// and persists it as a draft. To convert it to a posted entry, call
// Post.
func (s *AccountingService) CreateEntry(ctx context.Context, in EntryInput) (*accounting.JournalEntry, error) {
	var out *accounting.JournalEntry
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		entry, err := accounting.NewJournalEntry(time.Now().UTC(), accounting.NewJournalEntryOptions{
			CompanyID:      in.CompanyID,
			FiscalPeriodID: in.FiscalPeriodID,
			Number:         in.Number,
			EntryDate:      in.EntryDate,
			Description:    in.Description,
			Source:         in.Source,
			SourceID:       in.SourceID,
		})
		if err != nil {
			return err
		}
		for _, l := range in.Lines {
			line, err := accounting.NewJournalEntryLine(accounting.NewJournalEntryLineOptions{
				AccountID:           l.AccountID,
				Description:         l.Description,
				Debit:               l.Debit,
				Credit:              l.Credit,
				CurrencyCode:        l.CurrencyCode,
				ExchangeRate:        l.ExchangeRate,
				AmountInTxnCurrency: l.AmountInTxnCurrency,
			})
			if err != nil {
				return err
			}
			if err := entry.AddLine(line); err != nil {
				return err
			}
		}
		if err := entry.Validate(); err != nil {
			return err
		}
		if err := uow.JournalEntries().Create(ctx, entry); err != nil {
			return err
		}
		out = entry
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Post transitions a draft entry to posted. After this call the entry
// is immutable; corrections require a reversing entry.
func (s *AccountingService) Post(ctx context.Context, id uuid.UUID, by uuid.UUID) (*accounting.JournalEntry, error) {
	var out *accounting.JournalEntry
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		entry, err := uow.JournalEntries().GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := entry.Post(time.Now().UTC(), by); err != nil {
			return err
		}
		if err := uow.JournalEntries().Update(ctx, entry); err != nil {
			return err
		}
		out = entry
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("journal entry posted", "entry_id", id, "by", by)
	return out, nil
}

// Reverse creates a reversing entry that flips debits/credits and
// marks the original as reversed. Both entries are persisted in the
// same transaction. The reversing entry's date and number come from
// the application layer; this method only handles the side-effects.
func (s *AccountingService) Reverse(ctx context.Context, originalID, reversingID uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		original, err := uow.JournalEntries().GetByID(ctx, originalID)
		if err != nil {
			return err
		}
		reversing, err := uow.JournalEntries().GetByID(ctx, reversingID)
		if err != nil {
			return err
		}
		original.MarkAsReversed(reversingID)
		if err := uow.JournalEntries().Update(ctx, original); err != nil {
			return err
		}
		if err := uow.JournalEntries().Update(ctx, reversing); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.log.Info("journal entry reversed", "original_id", originalID, "reversing_id", reversingID)
	return nil
}

// AccountBalance returns the running balance of a chart-of-accounts
// entry as of `at`.
func (s *AccountingService) AccountBalance(ctx context.Context, accountID uuid.UUID, at time.Time) (string, error) {
	return s.ledger.GetAccountBalance(ctx, accountID, at)
}

// TrialBalance returns one row per account for the given period.
func (s *AccountingService) TrialBalance(ctx context.Context, fiscalPeriodID uuid.UUID) ([]repositories.TrialBalanceRow, error) {
	return s.ledger.GetTrialBalance(ctx, fiscalPeriodID)
}

// CreateChartOfAccountsInput seeds a new account.
type CreateChartOfAccountsInput struct {
	CompanyID      uuid.UUID
	Code           valueobjects.ChartOfAccountsCode
	Name           string
	Type           enums.AccountType
	ParentID       *uuid.UUID
	Path           string
	Depth          int
	AllowsMovement bool
	Description    string
}

// CreateChartOfAccounts persists a chart-of-accounts row. The depth
// and path are typically computed by the application use case from
// the parent.
func (s *AccountingService) CreateChartOfAccounts(ctx context.Context, in CreateChartOfAccountsInput) (*accounting.ChartOfAccount, error) {
	var out *accounting.ChartOfAccount
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		acc, err := accounting.NewChartOfAccount(time.Now().UTC(), accounting.NewChartOfAccountOptions{
			CompanyID:      in.CompanyID,
			Code:           in.Code,
			Name:           in.Name,
			Type:           in.Type,
			ParentID:       in.ParentID,
			Path:           in.Path,
			Depth:          in.Depth,
			AllowsMovement: in.AllowsMovement,
			Description:    in.Description,
		})
		if err != nil {
			return err
		}
		if err := uow.ChartOfAccounts().Create(ctx, acc); err != nil {
			return err
		}
		out = acc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListChartOfAccounts returns the chart for a company.
func (s *AccountingService) ListChartOfAccounts(ctx context.Context, companyID uuid.UUID) ([]*accounting.ChartOfAccount, error) {
	page, err := s.chart.List(ctx, repositories.ChartOfAccountsFilter{
		CompanyID: &companyID,
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// Validate is a thin wrapper around the entity's Validate method. It
// is exposed so the application use case (which constructs the entry)
// can sanity-check before persistence.
func Validate(e *accounting.JournalEntry) error {
	return e.Validate()
}

// silence unused-import lint
var _ = repositories.ErrNotFound
var _ = enums.JournalTypeManual
