package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/treasury"
)

type bankTransactionRepository struct {
	q persistence.Querier
}

func NewBankTransactionRepository(db *sql.DB) *bankTransactionRepository {
	return &bankTransactionRepository{q: persistence.FromDB(db)}
}

const bankTransactionColumns = `
	id, bank_account_id, transaction_date, value_date, description, amount,
	type, reference, balance_after, is_reconciled, reconciled_at,
	reconciled_by, journal_entry_id, created_at, updated_at
`

func (r *bankTransactionRepository) Create(ctx context.Context, t *treasury.BankTransaction) error {
	const q = `INSERT INTO bank_transactions (
		id, bank_account_id, transaction_date, value_date, description, amount,
		type, reference, balance_after, is_reconciled, reconciled_at,
		reconciled_by, journal_entry_id, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		t.ID, t.BankAccountID, t.TransactionDate, persistence.NullIfZeroTime(&t.ValueDate),
		persistence.NullIfEmpty(t.Description), t.Amount.String(), t.Type.String(),
		persistence.NullIfEmpty(t.Reference), t.BalanceAfter.String(), t.IsReconciled,
		persistence.NullIfZeroTime(t.ReconciledAt),
		persistence.NullIfEmptyUUID(t.ReconciledBy), persistence.NullIfEmptyUUID(t.JournalEntryID),
		t.CreatedAt, t.UpdatedAt,
	)
	return persistence.Translate(err)
}

func (r *bankTransactionRepository) Update(ctx context.Context, t *treasury.BankTransaction) error {
	const q = `UPDATE bank_transactions SET
		transaction_date = $1, value_date = $2, description = $3, amount = $4,
		type = $5, reference = $6, balance_after = $7, is_reconciled = $8,
		reconciled_at = $9, reconciled_by = $10, journal_entry_id = $11, updated_at = $12
	 WHERE id = $13`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		t.TransactionDate, persistence.NullIfZeroTime(&t.ValueDate),
		persistence.NullIfEmpty(t.Description), t.Amount.String(), t.Type.String(),
		persistence.NullIfEmpty(t.Reference), t.BalanceAfter.String(), t.IsReconciled,
		persistence.NullIfZeroTime(t.ReconciledAt),
		persistence.NullIfEmptyUUID(t.ReconciledBy), persistence.NullIfEmptyUUID(t.JournalEntryID),
		time.Now().UTC(), t.ID,
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

func (r *bankTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*treasury.BankTransaction, error) {
	q := `SELECT ` + bankTransactionColumns + ` FROM bank_transactions WHERE id = $1`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanBankTransaction(row)
}

func (r *bankTransactionRepository) List(ctx context.Context, filter treasury.BankTransactionFilter) (repositories.Page[*treasury.BankTransaction], error) {
	var (
		clauses []string
		args    []any
	)
	if filter.BankAccountID != nil {
		clauses = append(clauses, fmt.Sprintf("bank_account_id = $%d", len(args)+1))
		args = append(args, *filter.BankAccountID)
	}
	if filter.Reconciled != nil {
		clauses = append(clauses, fmt.Sprintf("is_reconciled = $%d", len(args)+1))
		args = append(args, *filter.Reconciled)
	}
	if !filter.OccurredRange.From.IsZero() {
		clauses = append(clauses, fmt.Sprintf("transaction_date >= $%d", len(args)+1))
		args = append(args, filter.OccurredRange.From)
	}
	if !filter.OccurredRange.To.IsZero() {
		clauses = append(clauses, fmt.Sprintf("transaction_date <= $%d", len(args)+1))
		args = append(args, filter.OccurredRange.To)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := "1=1"
	if len(clauses) > 0 {
		where = persistence.JoinClauses(clauses)
	}
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM bank_transactions WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*treasury.BankTransaction]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM bank_transactions WHERE %s ORDER BY transaction_date DESC LIMIT $%d OFFSET $%d",
			bankTransactionColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*treasury.BankTransaction]{}, persistence.Translate(err)
	}
	out := make([]*treasury.BankTransaction, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		t, err := scanBankTransactionFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, t)
		return nil
	}); err != nil {
		return repositories.Page[*treasury.BankTransaction]{}, err
	}
	return repositories.Page[*treasury.BankTransaction]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func scanBankTransaction(row *sql.Row) (*treasury.BankTransaction, error) {
	t := &treasury.BankTransaction{}
	var (
		valueDate, reconciledAt                 sql.NullTime
		reference, description, reconciledBy    sql.NullString
		journalEntryID                          sql.NullString
		balanceAfter                            sql.NullString
		amount, txType                          string
	)
	err := persistence.ScanRow(row,
		&t.ID, &t.BankAccountID, &t.TransactionDate, &valueDate, &description,
		&amount, &txType, &reference, &balanceAfter, &t.IsReconciled,
		&reconciledAt, &reconciledBy, &journalEntryID, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := decodeBankTransaction(t, valueDate, reconciledAt, reference, description, reconciledBy, journalEntryID, amount, balanceAfter, txType); err != nil {
		return nil, err
	}
	return t, nil
}

func scanBankTransactionFromRows(rows *sql.Rows) (*treasury.BankTransaction, error) {
	t := &treasury.BankTransaction{}
	var (
		valueDate, reconciledAt                 sql.NullTime
		reference, description, reconciledBy    sql.NullString
		journalEntryID                          sql.NullString
		balanceAfter                            sql.NullString
		amount, txType                          string
	)
	if err := rows.Scan(
		&t.ID, &t.BankAccountID, &t.TransactionDate, &valueDate, &description,
		&amount, &txType, &reference, &balanceAfter, &t.IsReconciled,
		&reconciledAt, &reconciledBy, &journalEntryID, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	if err := decodeBankTransaction(t, valueDate, reconciledAt, reference, description, reconciledBy, journalEntryID, amount, balanceAfter, txType); err != nil {
		return nil, err
	}
	return t, nil
}

func decodeBankTransaction(t *treasury.BankTransaction, valueDate, reconciledAt sql.NullTime, reference, description, reconciledBy, journalEntryID sql.NullString, amount string, balanceAfter sql.NullString, txType string) error {
	if valueDate.Valid {
		t.ValueDate = valueDate.Time
	}
	if description.Valid {
		t.Description = description.String
	}
	if reference.Valid {
		t.Reference = reference.String
	}
	am, err := persistence.ParseMoney(amount)
	if err != nil {
		return err
	}
	t.Amount = am
	t.Type = enums.BankTransactionType(txType)
	if balanceAfter.Valid {
		bal, err := persistence.ParseMoney(balanceAfter.String)
		if err != nil {
			return err
		}
		t.BalanceAfter = bal
	}
	if reconciledAt.Valid {
		rt := reconciledAt.Time
		t.ReconciledAt = &rt
	}
	if reconciledBy.Valid {
		id := persistence.ParseUUID(reconciledBy.String)
		t.ReconciledBy = &id
	}
	if journalEntryID.Valid {
		id := persistence.ParseUUID(journalEntryID.String)
		t.JournalEntryID = &id
	}
	return nil
}

var _ treasury.BankTransactionRepository = (*bankTransactionRepository)(nil)
