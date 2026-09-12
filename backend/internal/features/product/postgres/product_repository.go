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
	"vfinancy/backend/internal/features/product"
)

// productRepository is the SQL implementation of
// product.ProductRepository. The same struct can run against either
// *sql.DB (auto-commit) or *sql.Tx (transaction-bound); the Querier
// field is set by the constructor.
type productRepository struct {
	q persistence.Querier
}

// NewProductRepository returns an auto-commit implementation.
func NewProductRepository(db *sql.DB) *productRepository {
	return &productRepository{q: persistence.FromDB(db)}
}

const productColumns = `
	id, sku, description, unit_code,
	cost_usd, sale_price, is_active,
	created_at, updated_at, deleted_at, created_by, updated_by
`

func (r *productRepository) Create(ctx context.Context, p *product.Product) error {
	const q = `INSERT INTO products (
		id, sku, description, unit_code,
		cost_usd, sale_price, is_active,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.ID,
		p.SKU.String(),
		p.Description,
		p.UnitCode,
		p.CostUSD.String(),
		p.SalePrice.String(),
		p.IsActive,
		p.CreatedAt, p.UpdatedAt,
		persistence.NullIfZeroTime(p.DeletedAt),
		persistence.NullIfEmpty(p.CreatedBy),
		persistence.NullIfEmpty(p.UpdatedBy),
	)
	return persistence.Translate(err)
}

func (r *productRepository) Update(ctx context.Context, p *product.Product) error {
	const q = `UPDATE products SET
		sku = $1, description = $2, unit_code = $3,
		cost_usd = $4, sale_price = $5, is_active = $6,
		updated_at = $7, updated_by = $8
	 WHERE id = $9 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.SKU.String(),
		p.Description,
		p.UnitCode,
		p.CostUSD.String(),
		p.SalePrice.String(),
		p.IsActive,
		time.Now().UTC(),
		persistence.NullIfEmpty(p.UpdatedBy),
		p.ID,
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

func (r *productRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE products SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, time.Now().UTC(), id)
	if err != nil {
		return persistence.Translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	q := `SELECT ` + productColumns + ` FROM products WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanProduct(row)
}

func (r *productRepository) GetByDescription(ctx context.Context, description string) (*product.Product, error) {
	q := `SELECT ` + productColumns + ` FROM products WHERE description = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, description)
	return scanProduct(row)
}

func (r *productRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, `SELECT 1 FROM products WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&n); err != nil {
		if persistence.IsPgNoRows(err) {
			return false, nil
		}
		return false, persistence.Translate(err)
	}
	return n == 1, nil
}

func (r *productRepository) List(ctx context.Context, filter product.ProductFilter) (repositories.Page[*product.Product], error) {
	var clauses []string
	if !filter.IncludeDeleted {
		clauses = append(clauses, "deleted_at IS NULL")
	}
	var args []any
	if filter.IsActive != nil {
		clauses = append(clauses, fmt.Sprintf("is_active = $%d", len(args)+1))
		args = append(args, *filter.IsActive)
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf(
			"(LOWER(description) LIKE LOWER($%d) OR LOWER(sku) LIKE LOWER($%d))",
			len(args)+1, len(args)+2,
		))
		like := "%" + filter.Search + "%"
		args = append(args, like, like)
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + persistence.JoinClauses(clauses)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 1000)

	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM products"+where, args...).Scan(&total); err != nil {
		return repositories.Page[*product.Product]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM products%s ORDER BY description LIMIT $%d OFFSET $%d",
			productColumns, where, limitPos, offsetPos),
		args...)
	if err != nil {
		return repositories.Page[*product.Product]{}, persistence.Translate(err)
	}
	out := make([]*product.Product, 0, limit)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		p, err := scanProductFromRows(r)
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return repositories.Page[*product.Product]{}, err
	}
	return repositories.Page[*product.Product]{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

// --- scanning helpers ---

// productRow is the raw scan target shared by the *sql.Row and
// *sql.Rows variants.
type productRow struct {
	sku                  string
	description          string
	unitCode             string
	costUSD, salePrice   string
	isActive             bool
	deletedAt            sql.NullTime
	createdBy, updatedBy sql.NullString
}

// scanProduct is the *sql.Row variant. NotFound rows are translated
// to repositories.ErrNotFound by persistence.ScanRow.
func scanProduct(row *sql.Row) (*product.Product, error) {
	p := &product.Product{}
	pr := &productRow{}
	if err := persistence.ScanRow(row,
		&p.ID, &pr.sku, &pr.description, &pr.unitCode,
		&pr.costUSD, &pr.salePrice, &pr.isActive,
		&p.CreatedAt, &p.UpdatedAt, &pr.deletedAt, &pr.createdBy, &pr.updatedBy,
	); err != nil {
		return nil, err
	}
	return pr.fill(p)
}

// scanProductFromRows is the *sql.Rows variant.
func scanProductFromRows(rows *sql.Rows) (*product.Product, error) {
	p := &product.Product{}
	pr := &productRow{}
	if err := rows.Scan(
		&p.ID, &pr.sku, &pr.description, &pr.unitCode,
		&pr.costUSD, &pr.salePrice, &pr.isActive,
		&p.CreatedAt, &p.UpdatedAt, &pr.deletedAt, &pr.createdBy, &pr.updatedBy,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	return pr.fill(p)
}

// fill converts the raw scanned columns into the entity's value
// objects.
func (pr *productRow) fill(p *product.Product) (*product.Product, error) {
	sku, err := valueobjects.NewSKU(pr.sku)
	if err != nil {
		return nil, err
	}
	p.SKU = sku
	p.Description = pr.description
	p.UnitCode = pr.unitCode
	cost, err := persistence.ParseMoney(pr.costUSD)
	if err != nil {
		return nil, err
	}
	p.CostUSD = cost
	price, err := persistence.ParseMoney(pr.salePrice)
	if err != nil {
		return nil, err
	}
	p.SalePrice = price
	p.IsActive = pr.isActive
	if pr.deletedAt.Valid {
		t := pr.deletedAt.Time
		p.DeletedAt = &t
	}
	if pr.createdBy.Valid {
		p.CreatedBy = pr.createdBy.String
	}
	if pr.updatedBy.Valid {
		p.UpdatedBy = pr.updatedBy.String
	}
	return p, nil
}
