package accounting

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"
	"time"

	"github.com/google/uuid"

)

// ChartOfAccountsFilter is the input to
// ChartOfAccountsRepository.List.
type ChartOfAccountsFilter struct {
	CompanyID   *uuid.UUID
	AccountType string
	ParentID    *uuid.UUID
	ActiveOnly  bool
	repositories.PageRequest
}

// ChartOfAccountsRepository persists the chart of accounts. The
// repository also exposes path-based lookups for the UI (e.g. "show
// the account tree under 1.1").
type ChartOfAccountsRepository interface {
	Create(ctx context.Context, a *ChartOfAccount) error
	Update(ctx context.Context, a *ChartOfAccount) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*ChartOfAccount, error)
	GetByCode(ctx context.Context, companyID uuid.UUID, code string) (*ChartOfAccount, error)
	List(ctx context.Context, filter ChartOfAccountsFilter) (repositories.Page[*ChartOfAccount], error)

	// ListChildren returns the direct children of a parent (or root
	// accounts if parent is nil). Used to render the account tree.
	ListChildren(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID) ([]*ChartOfAccount, error)
}

// JournalEntryFilter is the input to JournalRepository.List.
type JournalEntryFilter struct {
	CompanyID      *uuid.UUID
	FiscalPeriodID *uuid.UUID
	Source         string
	Status         string
	EntryRange     repositories.TimeRange
	repositories.PageRequest
}

// JournalRepository persists journal entries and their lines.
type JournalRepository interface {
	Create(ctx context.Context, e *JournalEntry) error
	Update(ctx context.Context, e *JournalEntry) error

	GetByID(ctx context.Context, id uuid.UUID) (*JournalEntry, error)
	GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*JournalEntry, error)
	List(ctx context.Context, filter JournalEntryFilter) (repositories.Page[*JournalEntry], error)

	// GetNextNumber returns the next sequential number for the
	// company's journal series.
	GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error)
}

// LedgerRepository exposes the account-balance queries used by the
// general ledger, trial balance and financial statements. Balances
// are derived from posted journal entries; the implementation may use
// the account_balances summary table for performance but must
// re-compute on demand when the summary is stale.
type LedgerRepository interface {
	// GetAccountBalance returns the running balance (debit - credit
	// for asset/expense accounts, credit - debit for the rest) for
	// a single account as of `at`.
	GetAccountBalance(ctx context.Context, accountID uuid.UUID, at time.Time) (string, error)

	// GetTrialBalance returns one row per account with the
	// opening, period and closing balance for the given period.
	GetTrialBalance(ctx context.Context, fiscalPeriodID uuid.UUID) ([]TrialBalanceRow, error)
}

// TrialBalanceRow is a single line of the trial balance report.
type TrialBalanceRow struct {
	AccountID       uuid.UUID
	AccountCode     string
	AccountName     string
	AccountType     string
	OpeningBalance  string
	DebitTotal      string
	CreditTotal     string
	ClosingBalance  string
}
