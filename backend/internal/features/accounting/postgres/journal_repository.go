package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/accounting"
)

type journalRepository struct {
	q persistence.Querier
}

func NewJournalRepository(db *sql.DB) *journalRepository {
	return &journalRepository{q: persistence.FromDB(db)}
}

const journalEntryColumns = `
	id, company_id, fiscal_period_id, number, entry_date, posting_date,
	description, source, source_id, status, reverses_entry_id,
	reversed_by_entry_id, posted_at, created_at, updated_at, created_by, posted_by
`

const journalEntryLineColumns = `
	id, journal_entry_id, line_number, account_id, description,
	debit, credit, currency_code, exchange_rate, amount_in_txn_currency, created_at
`

func (r *journalRepository) Create(ctx context.Context, e *accounting.JournalEntry) error {
	const q = `INSERT INTO journal_entries (
		id, company_id, fiscal_period_id, number, entry_date, posting_date,
		description, source, source_id, status, reverses_entry_id,
		reversed_by_entry_id, posted_at, created_at, updated_at, created_by, posted_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		e.ID, e.CompanyID, e.FiscalPeriodID, e.Number, e.EntryDate,
		persistence.NullIfZeroTime(e.PostingDate), e.Description, e.Source.String(),
		persistence.NullIfEmptyUUID(e.SourceID), e.Status.String(),
		persistence.NullIfEmptyUUID(e.ReversesEntryID), persistence.NullIfEmptyUUID(e.ReversedByEntryID),
		persistence.NullIfZeroTime(e.PostedAt), e.CreatedAt, e.UpdatedAt,
		persistence.NullIfEmptyUUID(e.CreatedBy), persistence.NullIfEmptyUUID(e.PostedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	const lq = `INSERT INTO journal_entry_lines (
		id, journal_entry_id, line_number, account_id, description,
		debit, credit, currency_code, exchange_rate, amount_in_txn_currency, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	for _, line := range e.Lines {
		createdAt := line.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		_, err := persistence.Q(ctx, r.q).ExecContext(ctx, lq,
			line.ID, e.ID, line.LineNumber, line.AccountID, line.Description,
			line.Debit.String(), line.Credit.String(), line.CurrencyCode.String(),
			line.ExchangeRate.String(), line.AmountInTxnCurrency.String(), createdAt,
		)
		if err != nil {
			return persistence.Translate(err)
		}
	}
	return nil
}

func (r *journalRepository) Update(ctx context.Context, e *accounting.JournalEntry) error {
	const q = `UPDATE journal_entries SET
		number = $2, entry_date = $3, posting_date = $4, description = $5,
		source = $6, source_id = $7, status = $8, reverses_entry_id = $9,
		reversed_by_entry_id = $10, posted_at = $11, updated_at = $12,
		created_by = $13, posted_by = $14
	 WHERE id = $1`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		e.ID, e.Number, e.EntryDate, persistence.NullIfZeroTime(e.PostingDate), e.Description,
		e.Source.String(), persistence.NullIfEmptyUUID(e.SourceID), e.Status.String(),
		persistence.NullIfEmptyUUID(e.ReversesEntryID), persistence.NullIfEmptyUUID(e.ReversedByEntryID),
		persistence.NullIfZeroTime(e.PostedAt), time.Now().UTC(),
		persistence.NullIfEmptyUUID(e.CreatedBy), persistence.NullIfEmptyUUID(e.PostedBy),
	)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *journalRepository) GetByID(ctx context.Context, id uuid.UUID) (*accounting.JournalEntry, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+journalEntryColumns+` FROM journal_entries WHERE id = $1`, id)
	return r.scanEntryWithLines(ctx, row)
}

func (r *journalRepository) GetByNumber(ctx context.Context, companyID uuid.UUID, number string) (*accounting.JournalEntry, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+journalEntryColumns+` FROM journal_entries WHERE company_id = $1 AND number = $2`, companyID, number)
	return r.scanEntryWithLines(ctx, row)
}

func (r *journalRepository) List(ctx context.Context, filter accounting.JournalEntryFilter) (repositories.Page[*accounting.JournalEntry], error) {
	var (
		clauses []string
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.FiscalPeriodID != nil {
		clauses = append(clauses, fmt.Sprintf("fiscal_period_id = $%d", len(args)+1))
		args = append(args, *filter.FiscalPeriodID)
	}
	if filter.Source != "" {
		clauses = append(clauses, fmt.Sprintf("source = $%d", len(args)+1))
		args = append(args, filter.Source)
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	if !filter.EntryRange.From.IsZero() {
		clauses = append(clauses, fmt.Sprintf("entry_date >= $%d", len(args)+1))
		args = append(args, filter.EntryRange.From)
	}
	if !filter.EntryRange.To.IsZero() {
		clauses = append(clauses, fmt.Sprintf("entry_date <= $%d", len(args)+1))
		args = append(args, filter.EntryRange.To)
	}
	where := "1=1"
	if len(clauses) > 0 {
		where = persistence.JoinClauses(clauses)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM journal_entries WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*accounting.JournalEntry]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM journal_entries WHERE %s ORDER BY entry_date DESC, number DESC LIMIT $%d OFFSET $%d",
			journalEntryColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*accounting.JournalEntry]{}, persistence.Translate(err)
	}
	out := make([]*accounting.JournalEntry, 0, limit)
	if err := persistence.ScanRows(rows, func(rs *sql.Rows) error {
		e, err := scanJournalEntryFromRows(rs)
		if err != nil {
			return err
		}
		out = append(out, e)
		return nil
	}); err != nil {
		return repositories.Page[*accounting.JournalEntry]{}, err
	}
	for _, e := range out {
		lines, err := r.loadLines(ctx, e.ID)
		if err != nil {
			return repositories.Page[*accounting.JournalEntry]{}, err
		}
		e.Lines = lines
	}
	return repositories.Page[*accounting.JournalEntry]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *journalRepository) GetNextNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	year := time.Now().UTC().Year()
	prefix := fmt.Sprintf("JE-%d-", year)
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM journal_entries WHERE company_id = $1 AND number LIKE $2`, companyID, prefix+"%").Scan(&n); err != nil {
		return "", persistence.Translate(err)
	}
	return fmt.Sprintf("JE-%d-%05d", year, n+1), nil
}

func (r *journalRepository) scanEntryWithLines(ctx context.Context, row *sql.Row) (*accounting.JournalEntry, error) {
	e, err := scanJournalEntry(row)
	if err != nil {
		return nil, err
	}
	lines, err := r.loadLines(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.Lines = lines
	return e, nil
}

func (r *journalRepository) loadLines(ctx context.Context, entryID uuid.UUID) ([]*accounting.JournalEntryLine, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT `+journalEntryLineColumns+` FROM journal_entry_lines WHERE journal_entry_id = $1 ORDER BY line_number`, entryID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*accounting.JournalEntryLine, 0)
	if err := persistence.ScanRows(rows, func(rs *sql.Rows) error {
		l, err := scanJournalEntryLineFromRows(rs)
		if err != nil {
			return err
		}
		out = append(out, l)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func scanJournalEntry(row *sql.Row) (*accounting.JournalEntry, error) {
	e := &accounting.JournalEntry{}
	var (
		postingDate, postedAt                 sql.NullTime
		description                            sql.NullString
		sourceID, reversesID, reversedByID    sql.NullString
		createdBy, postedBy                    sql.NullString
		source, status                         string
	)
	err := persistence.ScanRow(row,
		&e.ID, &e.CompanyID, &e.FiscalPeriodID, &e.Number, &e.EntryDate,
		&postingDate, &description, &source, &sourceID, &status,
		&reversesID, &reversedByID, &postedAt,
		&e.CreatedAt, &e.UpdatedAt, &createdBy, &postedBy,
	)
	if err != nil {
		return nil, err
	}
	decodeJournalEntry(e, source, status, description, postingDate, postedAt, sourceID, reversesID, reversedByID, createdBy, postedBy)
	return e, nil
}

func scanJournalEntryFromRows(rows *sql.Rows) (*accounting.JournalEntry, error) {
	e := &accounting.JournalEntry{}
	var (
		postingDate, postedAt                 sql.NullTime
		description                            sql.NullString
		sourceID, reversesID, reversedByID    sql.NullString
		createdBy, postedBy                    sql.NullString
		source, status                         string
	)
	if err := rows.Scan(
		&e.ID, &e.CompanyID, &e.FiscalPeriodID, &e.Number, &e.EntryDate,
		&postingDate, &description, &source, &sourceID, &status,
		&reversesID, &reversedByID, &postedAt,
		&e.CreatedAt, &e.UpdatedAt, &createdBy, &postedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	decodeJournalEntry(e, source, status, description, postingDate, postedAt, sourceID, reversesID, reversedByID, createdBy, postedBy)
	return e, nil
}

func decodeJournalEntry(e *accounting.JournalEntry, source, status string, description sql.NullString, postingDate, postedAt sql.NullTime, sourceID, reversesID, reversedByID, createdBy, postedBy sql.NullString) {
	e.Source = persistence.ParseJournalType(source)
	e.Status = persistence.ParseJournalStatus(status)
	if description.Valid {
		e.Description = description.String
	}
	if postingDate.Valid {
		d := valueobjects.Date(postingDate.Time)
		e.PostingDate = &d
	}
	if postedAt.Valid {
		t := postedAt.Time
		e.PostedAt = &t
	}
	if sourceID.Valid {
		id := persistence.ParseUUID(sourceID.String)
		e.SourceID = &id
	}
	if reversesID.Valid {
		id := persistence.ParseUUID(reversesID.String)
		e.ReversesEntryID = &id
	}
	if reversedByID.Valid {
		id := persistence.ParseUUID(reversedByID.String)
		e.ReversedByEntryID = &id
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		e.CreatedBy = &id
	}
	if postedBy.Valid {
		id := persistence.ParseUUID(postedBy.String)
		e.PostedBy = &id
	}
}

func scanJournalEntryLineFromRows(rows *sql.Rows) (*accounting.JournalEntryLine, error) {
	l := &accounting.JournalEntryLine{}
	var (
		description                         sql.NullString
		debit, credit, amount               string
		currencyCode, exchangeRate          string
	)
	if err := rows.Scan(
		&l.ID, &l.JournalEntryID, &l.LineNumber, &l.AccountID, &description,
		&debit, &credit, &currencyCode, &exchangeRate, &amount, &l.CreatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if description.Valid {
		l.Description = description.String
	}
	cc, err := valueobjects.NewCurrencyCode(currencyCode)
	if err != nil {
		return nil, err
	}
	er, err := valueobjects.ExchangeRateFromString(exchangeRate)
	if err != nil {
		return nil, err
	}
	d, err := persistence.ParseMoney(debit)
	if err != nil {
		return nil, err
	}
	cr, err := persistence.ParseMoney(credit)
	if err != nil {
		return nil, err
	}
	am, err := persistence.ParseMoney(amount)
	if err != nil {
		return nil, err
	}
	l.CurrencyCode = cc
	l.ExchangeRate = er
	l.Debit = d
	l.Credit = cr
	l.AmountInTxnCurrency = am
	return l, nil
}

var _ accounting.JournalRepository = (*journalRepository)(nil)
