package accounting

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// JournalEntry is the root aggregate for a double-entry bookkeeping
// posting. A posted entry is immutable; corrections require a
// reversing entry (ReversesEntryID is set on the reversing entry).
type JournalEntry struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	FiscalPeriodID uuid.UUID
	Number        string
	EntryDate     valueobjects.Date
	PostingDate   *valueobjects.Date
	Description   string
	Source        enums.JournalType
	SourceID      *uuid.UUID
	Status        enums.JournalStatus
	Lines         []*JournalEntryLine
	ReversesEntryID *uuid.UUID
	ReversedByEntryID *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PostedAt      *time.Time
	CreatedBy     *uuid.UUID
	PostedBy      *uuid.UUID
}

// NewJournalEntryOptions is the input to NewJournalEntry.
type NewJournalEntryOptions struct {
	CompanyID      uuid.UUID
	FiscalPeriodID uuid.UUID
	Number         string
	EntryDate      valueobjects.Date
	Description    string
	Source         enums.JournalType
	SourceID       *uuid.UUID
}

// NewJournalEntry creates a new draft entry with no lines. The
// application layer adds lines via AddLine and then either calls Post
// (after Validate) or saves the draft for later editing.
func NewJournalEntry(now time.Time, opts NewJournalEntryOptions) (*JournalEntry, error) {
	if opts.CompanyID == uuid.Nil || opts.FiscalPeriodID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company and fiscal period are required"))
	}
	if opts.Number == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("entry number is required"))
	}
	if !opts.Source.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("source is invalid"))
	}
	return &JournalEntry{
		ID:            uuid.New(),
		CompanyID:     opts.CompanyID,
		FiscalPeriodID: opts.FiscalPeriodID,
		Number:        opts.Number,
		EntryDate:     opts.EntryDate,
		Description:   opts.Description,
		Source:        opts.Source,
		SourceID:      opts.SourceID,
		Status:        enums.JournalStatusDraft,
		Lines:         []*JournalEntryLine{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// AddLine appends a line. Only allowed in Draft status.
func (e *JournalEntry) AddLine(line *JournalEntryLine) error {
	if !e.Status.IsEditable() {
		return derrors.Wrap(derrors.ErrJournalEntryAlreadyPosted, errField("cannot modify a posted or reversed entry"))
	}
	line.JournalEntryID = e.ID
	line.LineNumber = len(e.Lines) + 1
	e.Lines = append(e.Lines, line)
	return nil
}

// RemoveLine drops a line by ID. Only allowed in Draft status.
func (e *JournalEntry) RemoveLine(lineID uuid.UUID) error {
	if !e.Status.IsEditable() {
		return derrors.Wrap(derrors.ErrJournalEntryAlreadyPosted, errField("cannot modify a posted or reversed entry"))
	}
	for i, l := range e.Lines {
		if l.ID == lineID {
			e.Lines = append(e.Lines[:i], e.Lines[i+1:]...)
			for j, ln := range e.Lines {
				ln.LineNumber = j + 1
			}
			return nil
		}
	}
	return derrors.Wrap(derrors.ErrRequired, errField("line not found"))
}

// TotalDebit returns the sum of debits across all lines.
func (e *JournalEntry) TotalDebit() valueobjects.Money {
	sum := valueobjects.Zero()
	for _, l := range e.Lines {
		sum = sum.Add(l.Debit)
	}
	return sum
}

// TotalCredit returns the sum of credits across all lines.
func (e *JournalEntry) TotalCredit() valueobjects.Money {
	sum := valueobjects.Zero()
	for _, l := range e.Lines {
		sum = sum.Add(l.Credit)
	}
	return sum
}

// IsBalanced reports whether the entry satisfies
// SUM(debit) == SUM(credit).
func (e *JournalEntry) IsBalanced() bool {
	if len(e.Lines) == 0 {
		return false
	}
	return e.TotalDebit().Equals(e.TotalCredit())
}

// Validate runs the full set of business rules that must hold for an
// entry to be postable. The application layer calls this before
// Post().
func (e *JournalEntry) Validate() error {
	if len(e.Lines) < 2 {
		return derrors.Wrap(derrors.ErrUnbalancedJournalEntry, errField("entry must have at least two lines"))
	}
	if !e.IsBalanced() {
		return derrors.Wrap(derrors.ErrUnbalancedJournalEntry, errField("debits and credits do not balance"))
	}
	if e.TotalDebit().IsZero() {
		return derrors.Wrap(derrors.ErrUnbalancedJournalEntry, errField("entry has zero total"))
	}
	seen := make(map[uuid.UUID]struct{}, len(e.Lines))
	for _, l := range e.Lines {
		if _, dup := seen[l.AccountID]; dup {
			return derrors.Wrap(derrors.ErrDuplicateLine, errField("entry has duplicate account lines"))
		}
		seen[l.AccountID] = struct{}{}
	}
	return nil
}

// Post transitions the entry to Posted status. Once posted, the entry
// is immutable. The application layer is responsible for also marking
// the fiscal period (no-op if already open) and emitting the related
// domain events.
func (e *JournalEntry) Post(at time.Time, by uuid.UUID) error {
	if e.Status == enums.JournalStatusPosted {
		return derrors.Wrap(derrors.ErrJournalEntryAlreadyPosted, errField("entry is already posted"))
	}
	if e.Status == enums.JournalStatusReversed {
		return derrors.Wrap(derrors.ErrJournalEntryReversed, errField("entry has been reversed"))
	}
	if err := e.Validate(); err != nil {
		return err
	}
	e.Status = enums.JournalStatusPosted
	e.PostedAt = &at
	e.PostedBy = &by
	return nil
}

// MarkAsReversed flags this entry as having a reversing entry. The
// reversing entry itself sets ReversesEntryID to this entry's ID
// (so the link is bidirectional).
func (e *JournalEntry) MarkAsReversed(reversingEntryID uuid.UUID) {
	e.Status = enums.JournalStatusReversed
	e.ReversedByEntryID = &reversingEntryID
}

// IsBalanced is exposed as Validate, kept here for backwards-compatible
// naming. Callers should prefer Validate for the full set of checks.
func (e *JournalEntry) ValidateBalanceOnly() bool { return e.IsBalanced() }
