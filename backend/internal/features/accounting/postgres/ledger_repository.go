package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/accounting"
)

type ledgerRepository struct {
	q persistence.Querier
}

func NewLedgerRepository(db *sql.DB) *ledgerRepository {
	return &ledgerRepository{q: persistence.FromDB(db)}
}

func (r *ledgerRepository) GetAccountBalance(ctx context.Context, accountID uuid.UUID, at time.Time) (string, error) {
	var accountType string
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT type FROM chart_of_accounts WHERE id = $1`, accountID).Scan(&accountType); err != nil {
		if persistence.IsPgNoRows(err) {
			return "", repositories.ErrNotFound
		}
		return "", persistence.Translate(err)
	}
	return signedBalance(ctx, persistence.Q(ctx, r.q), accountID, accountType, nil, &at)
}

func (r *ledgerRepository) GetTrialBalance(ctx context.Context, fiscalPeriodID uuid.UUID) ([]accounting.TrialBalanceRow, error) {
	var start, end time.Time
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT period_start, period_end FROM fiscal_periods WHERE id = $1`, fiscalPeriodID).Scan(&start, &end); err != nil {
		if persistence.IsPgNoRows(err) {
			return nil, repositories.ErrNotFound
		}
		return nil, persistence.Translate(err)
	}

	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT id, code, name, type FROM chart_of_accounts
		 WHERE company_id = (SELECT company_id FROM fiscal_periods WHERE id = $1)
		 ORDER BY code`, fiscalPeriodID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	type accountRow struct {
		id   uuid.UUID
		code string
		name string
		typ  string
	}
	var accounts []accountRow
	if err := persistence.ScanRows(rows, func(rs *sql.Rows) error {
		var a accountRow
		if err := rs.Scan(&a.id, &a.code, &a.name, &a.typ); err != nil {
			return persistence.Translate(err)
		}
		accounts = append(accounts, a)
		return nil
	}); err != nil {
		return nil, err
	}

	out := make([]accounting.TrialBalanceRow, 0, len(accounts))
	openingCutoff := start.AddDate(0, 0, -1)
	for _, a := range accounts {
		opening, err := signedBalance(ctx, persistence.Q(ctx, r.q), a.id, a.typ, nil, &openingCutoff)
		if err != nil {
			return nil, err
		}
		debitTotal, creditTotal, err := accountTotals(ctx, persistence.Q(ctx, r.q), a.id, &start, &end)
		if err != nil {
			return nil, err
		}
		op, err := decimal.NewFromString(opening)
		if err != nil {
			return nil, err
		}
		dr, err := decimal.NewFromString(debitTotal)
		if err != nil {
			return nil, err
		}
		cr, err := decimal.NewFromString(creditTotal)
		if err != nil {
			return nil, err
		}
		closing := op.Add(dr).Sub(cr)
		if a.typ != "asset" && a.typ != "expense" {
			closing = op.Add(cr).Sub(dr)
		}
		out = append(out, accounting.TrialBalanceRow{
			AccountID:       a.id,
			AccountCode:     a.code,
			AccountName:     a.name,
			AccountType:     a.typ,
			OpeningBalance:  opening,
			DebitTotal:      debitTotal,
			CreditTotal:     creditTotal,
			ClosingBalance:  closing.StringFixed(2),
		})
	}
	return out, nil
}

func signedBalance(ctx context.Context, q persistence.Querier, accountID uuid.UUID, accountType string, from, to *time.Time) (string, error) {
	clauses := []string{"l.account_id = $1", "e.status = 'posted'"}
	args := []any{accountID}
	if from != nil {
		clauses = append(clauses, fmt.Sprintf("e.entry_date >= $%d", len(args)+1))
		args = append(args, *from)
	}
	if to != nil {
		clauses = append(clauses, fmt.Sprintf("e.entry_date <= $%d", len(args)+1))
		args = append(args, *to)
	}
	expr := "COALESCE(SUM(CAST(l.debit AS REAL)) - SUM(CAST(l.credit AS REAL)), 0)"
	if accountType != "asset" && accountType != "expense" {
		expr = "COALESCE(SUM(CAST(l.credit AS REAL)) - SUM(CAST(l.debit AS REAL)), 0)"
	}
	var f float64
	err := q.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT %s FROM journal_entry_lines l
			JOIN journal_entries e ON e.id = l.journal_entry_id
			WHERE %s`, expr, persistence.JoinClauses(clauses)),
		args...).Scan(&f)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return "0.00", nil
		}
		return "", persistence.Translate(err)
	}
	return decimal.NewFromFloat(f).StringFixed(2), nil
}

func accountTotals(ctx context.Context, q persistence.Querier, accountID uuid.UUID, from, to *time.Time) (string, string, error) {
	clauses := []string{"l.account_id = $1", "e.status = 'posted'"}
	args := []any{accountID}
	if from != nil {
		clauses = append(clauses, fmt.Sprintf("e.entry_date >= $%d", len(args)+1))
		args = append(args, *from)
	}
	if to != nil {
		clauses = append(clauses, fmt.Sprintf("e.entry_date <= $%d", len(args)+1))
		args = append(args, *to)
	}
	var debit, credit float64
	err := q.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT COALESCE(SUM(CAST(l.debit AS REAL)), 0), COALESCE(SUM(CAST(l.credit AS REAL)), 0)
			FROM journal_entry_lines l
			JOIN journal_entries e ON e.id = l.journal_entry_id
			WHERE %s`, persistence.JoinClauses(clauses)),
		args...).Scan(&debit, &credit)
	if err != nil {
		if persistence.IsPgNoRows(err) {
			return "0.00", "0.00", nil
		}
		return "", "", persistence.Translate(err)
	}
	return decimal.NewFromFloat(debit).StringFixed(2), decimal.NewFromFloat(credit).StringFixed(2), nil
}

var _ accounting.LedgerRepository = (*ledgerRepository)(nil)
