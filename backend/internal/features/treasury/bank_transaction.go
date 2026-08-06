package treasury

import (
	"time"

	"github.com/google/uuid"
)

// BankTransaction is a single movement against a bank account
// (deposit, withdrawal, fee, interest, transfer). The repository
// reconciles transactions against the bank statement.
//
// This is intentionally a thin data type. Reconciliation logic and
// matching to payments live in the application layer.
type BankTransaction struct {
	ID            uuid.UUID
	BankAccountID uuid.UUID
	TransactionDate time.Time
	ValueDate      time.Time
	Description    string
	Amount         string // NUMERIC(18,2) serialized
	Type           string // deposit | withdrawal | fee | interest | transfer | other
	Reference      string
	BalanceAfter   string
	IsReconciled   bool
	ReconciledAt   *time.Time
	ReconciledBy   *uuid.UUID
	JournalEntryID *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
