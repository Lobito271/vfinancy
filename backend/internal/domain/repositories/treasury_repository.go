package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/treasury"
)

// BankAccountFilter is the input to BankAccountRepository.List.
type BankAccountFilter struct {
	CompanyID *uuid.UUID
	BranchID  *uuid.UUID
	IsActive  *bool
	PageRequest
}

// BankAccountRepository persists bank accounts and their
// transactions.
type BankAccountRepository interface {
	Create(ctx context.Context, a *treasury.BankAccount) error
	Update(ctx context.Context, a *treasury.BankAccount) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*treasury.BankAccount, error)
	List(ctx context.Context, filter BankAccountFilter) (Page[*treasury.BankAccount], error)
}

// BankTransactionFilter is the input to
// BankTransactionRepository.List.
type BankTransactionFilter struct {
	BankAccountID *uuid.UUID
	Reconciled     *bool
	OccurredRange  TimeRange
	PageRequest
}

// BankTransactionRepository persists bank transactions (movements
// against a bank account). Reconciliation is reflected by the
// is_reconciled flag.
type BankTransactionRepository interface {
	Create(ctx context.Context, t *treasury.BankTransaction) error
	Update(ctx context.Context, t *treasury.BankTransaction) error

	GetByID(ctx context.Context, id uuid.UUID) (*treasury.BankTransaction, error)
	List(ctx context.Context, filter BankTransactionFilter) (Page[*treasury.BankTransaction], error)
}

// CreditCardRepository persists company-issued credit cards.
type CreditCardRepository interface {
	Create(ctx context.Context, c *treasury.CreditCard) error
	Update(ctx context.Context, c *treasury.CreditCard) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*treasury.CreditCard, error)
	List(ctx context.Context, companyID uuid.UUID) ([]*treasury.CreditCard, error)
}

// ExchangeRateRepository persists daily exchange rates (one row per
// (from, to, effective_date)).
type ExchangeRateRepository interface {
	Upsert(ctx context.Context, from, to string, rate string, effectiveDate string, source string) error
	// GetForDate returns the rate effective at (or before) the given
	// date for the (from, to) pair. Returns ErrNotFound if no rate
	// has been recorded yet.
	GetForDate(ctx context.Context, from, to string, date string) (string, error)
	// GetLatest returns the most recent rate for the (from, to) pair.
	GetLatest(ctx context.Context, from, to string) (string, error)
}
