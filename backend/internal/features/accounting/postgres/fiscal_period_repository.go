package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/accounting"
)

type fiscalPeriodRepository struct {
	q persistence.Querier
}

func NewFiscalPeriodRepository(db *sql.DB) *fiscalPeriodRepository {
	return &fiscalPeriodRepository{q: persistence.FromDB(db)}
}

const fiscalPeriodColumns = `
	id, company_id, name, period_start, period_end, status, closed_at,
	created_at, updated_at, created_by, updated_by
`

func (r *fiscalPeriodRepository) Create(ctx context.Context, p *accounting.FiscalPeriod) error {
	const q = `INSERT INTO fiscal_periods (
		id, company_id, name, period_start, period_end, status, closed_at,
		created_at, updated_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.ID, p.CompanyID, p.Name, p.PeriodStart, p.PeriodEnd, p.Status,
		persistence.NullIfZeroTime(p.ClosedAt),
		p.CreatedAt, p.UpdatedAt,
		persistence.NullIfEmptyUUID(p.CreatedBy), persistence.NullIfEmptyUUID(p.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *fiscalPeriodRepository) GetByID(ctx context.Context, id uuid.UUID) (*accounting.FiscalPeriod, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+fiscalPeriodColumns+` FROM fiscal_periods WHERE id = $1`, id)
	return scanFiscalPeriod(row)
}

func (r *fiscalPeriodRepository) GetOpenForDate(ctx context.Context, companyID uuid.UUID, date time.Time) (*accounting.FiscalPeriod, error) {
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx,
		`SELECT `+fiscalPeriodColumns+` FROM fiscal_periods
		 WHERE company_id = $1 AND status = 'open' AND period_start <= $2 AND period_end >= $2
		 ORDER BY period_start DESC LIMIT 1`, companyID, date)
	return scanFiscalPeriod(row)
}

func (r *fiscalPeriodRepository) List(ctx context.Context, filter accounting.FiscalPeriodFilter) (repositories.Page[*accounting.FiscalPeriod], error) {
	var (
		clauses []string
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}
	where := "1=1"
	if len(clauses) > 0 {
		where = persistence.JoinClauses(clauses)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM fiscal_periods WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*accounting.FiscalPeriod]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM fiscal_periods WHERE %s ORDER BY period_start DESC LIMIT $%d OFFSET $%d",
			fiscalPeriodColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*accounting.FiscalPeriod]{}, persistence.Translate(err)
	}
	out := make([]*accounting.FiscalPeriod, 0, limit)
	if err := persistence.ScanRows(rows, func(rs *sql.Rows) error {
		p, err := scanFiscalPeriodFromRows(rs)
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return repositories.Page[*accounting.FiscalPeriod]{}, err
	}
	return repositories.Page[*accounting.FiscalPeriod]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func scanFiscalPeriod(row *sql.Row) (*accounting.FiscalPeriod, error) {
	p := &accounting.FiscalPeriod{}
	var (
		closedAt                    sql.NullTime
		createdAt, updatedAt        sql.NullTime
		createdBy, updatedBy        sql.NullString
	)
	err := persistence.ScanRow(row,
		&p.ID, &p.CompanyID, &p.Name, &p.PeriodStart, &p.PeriodEnd, &p.Status,
		&closedAt, &createdAt, &updatedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}
	decodeFiscalPeriod(p, closedAt, createdAt, updatedAt, createdBy, updatedBy)
	return p, nil
}

func scanFiscalPeriodFromRows(rows *sql.Rows) (*accounting.FiscalPeriod, error) {
	p := &accounting.FiscalPeriod{}
	var (
		closedAt                    sql.NullTime
		createdAt, updatedAt        sql.NullTime
		createdBy, updatedBy        sql.NullString
	)
	if err := rows.Scan(
		&p.ID, &p.CompanyID, &p.Name, &p.PeriodStart, &p.PeriodEnd, &p.Status,
		&closedAt, &createdAt, &updatedAt, &createdBy, &updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	decodeFiscalPeriod(p, closedAt, createdAt, updatedAt, createdBy, updatedBy)
	return p, nil
}

func decodeFiscalPeriod(p *accounting.FiscalPeriod, closedAt, createdAt, updatedAt sql.NullTime, createdBy, updatedBy sql.NullString) {
	if closedAt.Valid {
		t := closedAt.Time
		p.ClosedAt = &t
	}
	p.CreatedAt = createdAt.Time
	p.UpdatedAt = updatedAt.Time
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		p.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		p.UpdatedBy = &id
	}
}

var _ accounting.FiscalPeriodRepository = (*fiscalPeriodRepository)(nil)
