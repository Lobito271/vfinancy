package treasury

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

// BankTransaction is a single movement against a bank account
// (deposit, withdrawal, fee, interest, transfer). The repository
// reconciles transactions against the bank statement.
//
// This is intentionally a thin data type. Reconciliation logic and
// matching to payments live in the application layer.
type BankTransaction struct {
	ID              uuid.UUID
	BankAccountID   uuid.UUID
	TransactionDate time.Time
	ValueDate       time.Time
	Description     string
	Amount          valueobjects.Money
	Type            enums.BankTransactionType
	Reference       string
	BalanceAfter    valueobjects.Money
	IsReconciled    bool
	ReconciledAt    *time.Time
	ReconciledBy    *uuid.UUID
	JournalEntryID  *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
