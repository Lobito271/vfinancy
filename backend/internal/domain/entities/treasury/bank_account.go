package treasury

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// BankAccount is a bank account owned by the company. The current
// balance is maintained by the application layer as bank_transactions
// are recorded; the entity exposes a SetBalance for the persistence
// layer to use during rehydration.
type BankAccount struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	BankName      string
	AccountNumber string
	AccountType   string
	CurrencyCode  valueobjects.CurrencyCode
	GLAccountID   uuid.UUID
	CurrentBalance valueobjects.Money
	IsDefault     bool
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	CreatedBy     *uuid.UUID
	UpdatedBy     *uuid.UUID
}

// NewBankAccountOptions is the input to NewBankAccount.
type NewBankAccountOptions struct {
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	BankName      string
	AccountNumber string
	AccountType   string
	CurrencyCode  valueobjects.CurrencyCode
	GLAccountID   uuid.UUID
	IsDefault     bool
}

// NewBankAccount validates and constructs a bank account with a zero
// opening balance.
func NewBankAccount(now time.Time, opts NewBankAccountOptions) (*BankAccount, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company id is required"))
	}
	if opts.BankName == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("bank name is required"))
	}
	if opts.AccountNumber == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("account number is required"))
	}
	if opts.GLAccountID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("gl account is required"))
	}
	return &BankAccount{
		ID:             uuid.New(),
		CompanyID:      opts.CompanyID,
		BranchID:       opts.BranchID,
		BankName:       opts.BankName,
		AccountNumber:  opts.AccountNumber,
		AccountType:    opts.AccountType,
		CurrencyCode:   opts.CurrencyCode,
		GLAccountID:    opts.GLAccountID,
		CurrentBalance: valueobjects.Zero(),
		IsDefault:      opts.IsDefault,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// ApplyDelta adjusts the balance by a signed amount (deposits are
// positive, withdrawals negative). Used by the application layer after
// each bank_transaction is recorded.
func (a *BankAccount) ApplyDelta(delta valueobjects.Money) {
	a.CurrentBalance = a.CurrentBalance.Add(delta)
}

// Activate / Deactivate.
func (a *BankAccount) Activate()   { a.IsActive = true }
func (a *BankAccount) Deactivate() { a.IsActive = false }

// MarkAsDefault / ClearDefault.
func (a *BankAccount) MarkAsDefault() { a.IsDefault = true }
func (a *BankAccount) ClearDefault()  { a.IsDefault = false }
