package accounting

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// JournalEntryLine is a single debit or credit line in a journal entry.
// Exactly one of Debit/Credit must be non-zero.
type JournalEntryLine struct {
	ID                  uuid.UUID
	JournalEntryID      uuid.UUID
	LineNumber          int
	AccountID           uuid.UUID
	Description         string
	Debit               valueobjects.Money
	Credit              valueobjects.Money
	CurrencyCode        valueobjects.CurrencyCode
	ExchangeRate        valueobjects.ExchangeRate
	AmountInTxnCurrency valueobjects.Money
	CreatedAt           time.Time
}

// NewJournalEntryLineOptions is the input to NewJournalEntryLine.
type NewJournalEntryLineOptions struct {
	LineNumber          int
	AccountID           uuid.UUID
	Description         string
	Debit               valueobjects.Money
	Credit              valueobjects.Money
	CurrencyCode        valueobjects.CurrencyCode
	ExchangeRate        valueobjects.ExchangeRate
	AmountInTxnCurrency valueobjects.Money
}

// NewJournalEntryLine validates and constructs a line. Exactly one of
// Debit / Credit must be positive.
func NewJournalEntryLine(opts NewJournalEntryLineOptions) (*JournalEntryLine, error) {
	if opts.AccountID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("account id is required"))
	}
	debitZero := opts.Debit.IsZero()
	creditZero := opts.Credit.IsZero()
	if debitZero && creditZero {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("debit and credit cannot both be zero"))
	}
	if !debitZero && !creditZero {
		return nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("debit and credit cannot both be non-zero"))
	}
	if opts.Debit.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("debit cannot be negative"))
	}
	if opts.Credit.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("credit cannot be negative"))
	}
	return &JournalEntryLine{
		LineNumber:          opts.LineNumber,
		AccountID:           opts.AccountID,
		Description:         opts.Description,
		Debit:               opts.Debit,
		Credit:              opts.Credit,
		CurrencyCode:        opts.CurrencyCode,
		ExchangeRate:        opts.ExchangeRate,
		AmountInTxnCurrency: opts.AmountInTxnCurrency,
	}, nil
}

// Amount returns the absolute value posted to the account, in
// functional currency. Use Direction to know whether it is a debit or
// a credit.
func (l *JournalEntryLine) Amount() valueobjects.Money {
	if l.Debit.IsZero() {
		return l.Credit
	}
	return l.Debit
}

// IsDebit reports whether the line posts a debit.
func (l *JournalEntryLine) IsDebit() bool { return !l.Debit.IsZero() }

// IsCredit reports whether the line posts a credit.
func (l *JournalEntryLine) IsCredit() bool { return !l.Credit.IsZero() }
