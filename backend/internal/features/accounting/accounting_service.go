// Package accounting implements the business logic for the
// double-entry bookkeeping module: journal entries, posting, ledger
// queries, trial balance, and financial statements.
package accounting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/shared/logger"
)

// AccountingService owns the journal entry workflow plus the report
// queries (trial balance, etc.).
type AccountingService struct {
	entries JournalRepository
	chart   ChartOfAccountsRepository
	ledger  LedgerRepository
	periods FiscalPeriodRepository
	txm     repositories.TransactionManager
	log     *logger.Logger
}

// New returns an AccountingService ready for use.
func New(
	entries JournalRepository,
	chart ChartOfAccountsRepository,
	ledger LedgerRepository,
	periods FiscalPeriodRepository,
	txm repositories.TransactionManager,
	log *logger.Logger,
) *AccountingService {
	return &AccountingService{
		entries: entries,
		chart:   chart,
		ledger:  ledger,
		periods: periods,
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
func (s *AccountingService) CreateEntry(ctx context.Context, in EntryInput) (*JournalEntry, error) {
	var out *JournalEntry
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		number := in.Number
		if number == "" {
			n, err := s.entries.GetNextNumber(ctx, in.CompanyID)
			if err != nil {
				return err
			}
			number = n
		}
		entry, err := NewJournalEntry(time.Now().UTC(), NewJournalEntryOptions{
			CompanyID:      in.CompanyID,
			FiscalPeriodID: in.FiscalPeriodID,
			Number:         number,
			EntryDate:      in.EntryDate,
			Description:    in.Description,
			Source:         in.Source,
			SourceID:       in.SourceID,
		})
		if err != nil {
			return err
		}
		for _, l := range in.Lines {
			line, err := NewJournalEntryLine(NewJournalEntryLineOptions{
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
		if err := s.entries.Create(ctx, entry); err != nil {
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
func (s *AccountingService) Post(ctx context.Context, id uuid.UUID, by uuid.UUID) (*JournalEntry, error) {
	var out *JournalEntry
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		entry, err := s.entries.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := entry.Post(time.Now().UTC(), by); err != nil {
			return err
		}
		if err := s.entries.Update(ctx, entry); err != nil {
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
		original, err := s.entries.GetByID(ctx, originalID)
		if err != nil {
			return err
		}
		reversing, err := s.entries.GetByID(ctx, reversingID)
		if err != nil {
			return err
		}
		original.MarkAsReversed(reversingID)
		if err := s.entries.Update(ctx, original); err != nil {
			return err
		}
		if err := s.entries.Update(ctx, reversing); err != nil {
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
func (s *AccountingService) TrialBalance(ctx context.Context, fiscalPeriodID uuid.UUID) ([]TrialBalanceRow, error) {
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
// and path are computed by the caller from the parent.
func (s *AccountingService) CreateChartOfAccounts(ctx context.Context, in CreateChartOfAccountsInput) (*ChartOfAccount, error) {
	var out *ChartOfAccount
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		acc, err := NewChartOfAccount(time.Now().UTC(), NewChartOfAccountOptions{
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
		if err := s.chart.Create(ctx, acc); err != nil {
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
func (s *AccountingService) ListChartOfAccounts(ctx context.Context, companyID uuid.UUID) ([]*ChartOfAccount, error) {
	page, err := s.chart.List(ctx, ChartOfAccountsFilter{
		CompanyID:   &companyID,
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// GetChartAccount returns a single chart-of-accounts entry.
func (s *AccountingService) GetChartAccount(ctx context.Context, id uuid.UUID) (*ChartOfAccount, error) {
	return s.chart.GetByID(ctx, id)
}

// UpdateChartOfAccountInput describes the editable fields of an account.
type UpdateChartOfAccountInput struct {
	ID             uuid.UUID
	Name           string
	Type           enums.AccountType
	ParentID       *uuid.UUID
	Code           string
	Path           string
	Depth          int
	AllowsMovement bool
	IsActive       bool
	Description    string
}

// UpdateChartOfAccount persists changes to a chart-of-accounts entry.
func (s *AccountingService) UpdateChartOfAccount(ctx context.Context, in UpdateChartOfAccountInput) (*ChartOfAccount, error) {
	var out *ChartOfAccount
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		acc, err := s.chart.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if in.Name == "" {
			return derrors.Wrap(derrors.ErrRequired, errField("account name is required"))
		}
		if !in.Type.Valid() {
			return derrors.Wrap(derrors.ErrInvalidEnum, errField("account type is invalid"))
		}
		code, err := valueobjects.NewChartOfAccountsCode(in.Code)
		if err != nil {
			return err
		}
		acc.Code = code
		acc.Name = in.Name
		acc.Type = in.Type
		acc.ParentID = in.ParentID
		acc.Path = in.Path
		acc.Depth = in.Depth
		acc.AllowsMovement = in.AllowsMovement
		acc.IsActive = in.IsActive
		acc.Description = in.Description
		acc.UpdatedAt = time.Now().UTC()
		if err := s.chart.Update(ctx, acc); err != nil {
			return err
		}
		out = acc
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("chart of accounts updated", "account_id", out.ID, "code", out.Code)
	return out, nil
}

// DeleteChartOfAccount deactivates a chart-of-accounts entry.
func (s *AccountingService) DeleteChartOfAccount(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.chart.Delete(ctx, id)
	})
	if err != nil {
		return err
	}
	s.log.Info("chart of accounts deleted", "account_id", id)
	return nil
}

// NextEntryNumber returns the next sequential journal entry number.
func (s *AccountingService) NextEntryNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	return s.entries.GetNextNumber(ctx, companyID)
}

// ListFiscalPeriods returns the fiscal periods of a company.
func (s *AccountingService) ListFiscalPeriods(ctx context.Context, companyID uuid.UUID) ([]*FiscalPeriod, error) {
	page, err := s.periods.List(ctx, FiscalPeriodFilter{
		CompanyID:   &companyID,
		PageRequest: repositories.PageRequest{Limit: 100},
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// EnsureOpenFiscalPeriod returns the open fiscal period covering date,
// creating an open calendar-year period when none exists yet.
func (s *AccountingService) EnsureOpenFiscalPeriod(ctx context.Context, companyID uuid.UUID, date time.Time) (*FiscalPeriod, error) {
	var out *FiscalPeriod
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.periods.GetOpenForDate(ctx, companyID, date)
		if err == nil {
			out = p
			return nil
		}
		if !errors.Is(err, repositories.ErrNotFound) {
			return err
		}
		start := time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(date.Year(), time.December, 31, 23, 59, 59, 0, time.UTC)
		p, err = NewFiscalPeriod(time.Now().UTC(), NewFiscalPeriodOptions{
			CompanyID:   companyID,
			Name:        fmt.Sprintf("Ejercicio %d", date.Year()),
			PeriodStart: start,
			PeriodEnd:   end,
			Status:      "open",
		})
		if err != nil {
			return err
		}
		if err := s.periods.Create(ctx, p); err != nil {
			return err
		}
		out = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListEntries returns journal entries matching the filter.
func (s *AccountingService) ListEntries(ctx context.Context, filter JournalEntryFilter) (repositories.Page[*JournalEntry], error) {
	return s.entries.List(ctx, filter)
}
