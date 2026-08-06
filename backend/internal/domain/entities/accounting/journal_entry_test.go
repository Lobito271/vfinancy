package accounting

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

func mustMoney(t *testing.T, s string) valueobjects.Money {
	t.Helper()
	m, err := valueobjects.MoneyFromString(s)
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	return m
}

func mustDate(t *testing.T, y int, m time.Month, d int) valueobjects.Date {
	t.Helper()
	date, err := valueobjects.NewDate(y, m, d)
	if err != nil {
		t.Fatalf("date: %v", err)
	}
	return date
}

func newEntry(t *testing.T) *JournalEntry {
	t.Helper()
	e, err := NewJournalEntry(time.Now(), NewJournalEntryOptions{
		CompanyID:      uuid.New(),
		FiscalPeriodID: uuid.New(),
		Number:         "JE-1",
		EntryDate:      mustDate(t, 2026, 1, 15),
		Description:    "test",
		Source:         enums.JournalTypeManual,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return e
}

func mustLine(t *testing.T, accountID uuid.UUID, debit, credit string) *JournalEntryLine {
	t.Helper()
	d := mustMoney(t, debit)
	c := mustMoney(t, credit)
	if !d.IsZero() && !c.IsZero() {
		t.Fatalf("mustLine: only one of debit/credit may be non-zero")
	}
	li, err := NewJournalEntryLine(NewJournalEntryLineOptions{
		AccountID:    accountID,
		Debit:        d,
		Credit:       c,
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchange(t),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return li
}

func mustExchange(t *testing.T) valueobjects.ExchangeRate {
	t.Helper()
	r, err := valueobjects.ExchangeRateFromString("1")
	if err != nil {
		t.Fatalf("rate: %v", err)
	}
	return r
}

func TestJournalEntryNew(t *testing.T) {
	e := newEntry(t)
	if e.Status != enums.JournalStatusDraft {
		t.Errorf("status: %s", e.Status)
	}
	if !e.TotalDebit().IsZero() || !e.TotalCredit().IsZero() {
		t.Error("totals should be zero")
	}
	if e.IsBalanced() {
		t.Error("empty entry should not be balanced")
	}
}

func TestJournalEntryAddRemoveLine(t *testing.T) {
	e := newEntry(t)
	cash, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	rev, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	if err := e.AddLine(mustLine(t, cash, "100", "0")); err != nil {
		t.Errorf("add: %v", err)
	}
	if err := e.AddLine(mustLine(t, rev, "0", "100")); err != nil {
		t.Errorf("add: %v", err)
	}
	if !e.IsBalanced() {
		t.Error("entry should be balanced")
	}

	if err := e.RemoveLine(e.Lines[0].ID); err != nil {
		t.Errorf("remove: %v", err)
	}
	if e.IsBalanced() {
		t.Error("entry should not be balanced after remove")
	}
}

func TestJournalEntryBothDebitAndCreditRejected(t *testing.T) {
	acc, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	if _, err := NewJournalEntryLine(NewJournalEntryLineOptions{
		AccountID:    acc,
		Debit:        mustMoney(t, "100"),
		Credit:       mustMoney(t, "100"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchange(t),
	}); err == nil {
		t.Error("both debit and credit should fail")
	}
}

func TestJournalEntryNeitherDebitNorCreditRejected(t *testing.T) {
	acc, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	if _, err := NewJournalEntryLine(NewJournalEntryLineOptions{
		AccountID:    acc,
		Debit:        mustMoney(t, "0"),
		Credit:       mustMoney(t, "0"),
		CurrencyCode: valueobjects.MustCurrencyCode("PEN"),
		ExchangeRate: mustExchange(t),
	}); err == nil {
		t.Error("both zero should fail")
	}
}

func TestJournalEntryPostRequiresBalance(t *testing.T) {
	e := newEntry(t)
	cash, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	if err := e.AddLine(mustLine(t, cash, "100", "0")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.Post(time.Now(), uuid.New()); err == nil {
		t.Error("unbalanced entry should not be postable")
	}
}

func TestJournalEntryPostRequiresMinTwoLines(t *testing.T) {
	e := newEntry(t)
	// Empty entry: posting should fail because there are fewer than 2 lines.
	if err := e.Post(time.Now(), uuid.New()); err == nil {
		t.Error("empty entry should not be postable")
	}
}

func TestJournalEntryPostRequiresNoDuplicateAccounts(t *testing.T) {
	e := newEntry(t)
	cash, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	// Two lines on the same account: balanced (debit 100, credit 100) but
	// the duplicate-account rule rejects posting.
	if err := e.AddLine(mustLine(t, cash, "100", "0")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.AddLine(mustLine(t, cash, "0", "100")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.Post(time.Now(), uuid.New()); err == nil {
		t.Error("duplicate account should fail post")
	}
}

func TestJournalEntryPostHappyPath(t *testing.T) {
	e := newEntry(t)
	cash, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	rev, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	if err := e.AddLine(mustLine(t, cash, "100", "0")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.AddLine(mustLine(t, rev, "0", "100")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.Post(time.Now(), uuid.New()); err != nil {
		t.Errorf("post: %v", err)
	}
	if e.Status != enums.JournalStatusPosted {
		t.Errorf("status: %s", e.Status)
	}
	if e.PostedAt == nil {
		t.Error("posted_at should be set")
	}
	// Re-post should fail
	if err := e.Post(time.Now(), uuid.New()); err == nil {
		t.Error("double post should fail")
	}
}

func TestJournalEntryPostImmutability(t *testing.T) {
	e := newEntry(t)
	cash, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	rev, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	if err := e.AddLine(mustLine(t, cash, "100", "0")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.AddLine(mustLine(t, rev, "0", "100")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.Post(time.Now(), uuid.New()); err != nil {
		t.Fatalf("post: %v", err)
	}
	// Attempt to add a valid new line to a posted entry.
	other, _ := uuid.Parse("00000000-0000-0000-0000-000000000003")
	if err := e.AddLine(mustLine(t, other, "10", "0")); err == nil {
		t.Error("cannot add to posted entry")
	}
	if err := e.RemoveLine(e.Lines[0].ID); err == nil {
		t.Error("cannot remove from posted entry")
	}
}

func TestJournalEntryReverse(t *testing.T) {
	e := newEntry(t)
	cash, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	rev, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	if err := e.AddLine(mustLine(t, cash, "100", "0")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.AddLine(mustLine(t, rev, "0", "100")); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := e.Post(time.Now(), uuid.New()); err != nil {
		t.Fatalf("post: %v", err)
	}
	reversingID := uuid.New()
	e.MarkAsReversed(reversingID)
	if e.Status != enums.JournalStatusReversed {
		t.Errorf("status: %s", e.Status)
	}
	if e.ReversedByEntryID == nil || *e.ReversedByEntryID != reversingID {
		t.Error("reversing id not set")
	}
}

func TestChartOfAccountNormalBalance(t *testing.T) {
	cases := []struct {
		typ  enums.AccountType
		want enums.NormalBalance
	}{
		{enums.AccountTypeAsset, enums.DebitNormal},
		{enums.AccountTypeExpense, enums.DebitNormal},
		{enums.AccountTypeLiability, enums.CreditNormal},
		{enums.AccountTypeEquity, enums.CreditNormal},
		{enums.AccountTypeIncome, enums.CreditNormal},
	}
	for _, c := range cases {
		got := c.typ.NormalBalance()
		if got != c.want {
			t.Errorf("%s: got %s, want %s", c.typ, got, c.want)
		}
	}
}
