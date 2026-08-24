package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/product"
	"vfinancy/backend/infrastructure/persistence"
)

// productRepository is the SQL implementation of product.ProductRepository.
// The same struct can run against either *sql.DB (auto-commit) or *sql.Tx
// (transaction-bound); the Querier field is set by the constructor.
type productRepository struct {
	q persistence.Querier
}

// NewProductRepository returns an auto-commit implementation.
func NewProductRepository(db *sql.DB) *productRepository {
	return &productRepository{q: persistence.FromDB(db)}
}

// productColumns selects the products row plus the category / brand / unit
// display names via LEFT JOINs (read-model enrichment).
const productColumns = `
	p.id, p.company_id, p.sku, p.barcode, p.description,
	p.category_id, p.brand_id, p.unit_id, p.tax_id,
	p.cost_usd, p.sale_price, p.sale_currency,
	p.min_stock, p.max_stock, p.weight,
	p.is_active, p.is_service,
	p.created_at, p.updated_at, p.deleted_at, p.created_by, p.updated_by,
	pc.name AS category_name, pb.name AS brand_name, uom.name AS unit_name
`

const productFrom = `
	FROM products p
	LEFT JOIN product_categories pc ON pc.id = p.category_id AND pc.deleted_at IS NULL
	LEFT JOIN product_brands pb ON pb.id = p.brand_id AND pb.deleted_at IS NULL
	LEFT JOIN units_of_measure uom ON uom.id = p.unit_id
`

func (r *productRepository) Create(ctx context.Context, p *product.Product) error {
	const q = `INSERT INTO products (
		id, company_id, sku, barcode, description,
		category_id, brand_id, unit_id, tax_id,
		cost_usd, sale_price, sale_currency,
		min_stock, max_stock, weight,
		is_active, is_service,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.ID, p.CompanyID, p.SKU.String(), persistence.NullIfEmpty(p.Barcode.String()),
		p.Description,
		persistence.NullIfEmptyUUID(p.CategoryID), persistence.NullIfEmptyUUID(p.BrandID),
		p.UnitID, p.TaxID,
		p.CostUSD.String(), p.SalePrice.String(), p.SaleCurrency.String(),
		p.MinStock.String(), p.MaxStock.String(), p.Weight.String(),
		p.IsActive, p.IsService,
		p.CreatedAt, p.UpdatedAt, persistence.NullIfZeroTime(p.DeletedAt), p.CreatedBy, p.UpdatedBy,
	)
	return persistence.Translate(err)
}

func (r *productRepository) Update(ctx context.Context, p *product.Product) error {
	const q = `UPDATE products SET
		description = $1, category_id = $2, brand_id = $3, unit_id = $4, tax_id = $5,
		cost_usd = $6, sale_price = $7, sale_currency = $8,
		min_stock = $9, max_stock = $10, weight = $11,
		is_active = $12, is_service = $13, updated_at = $14, updated_by = $15
	 WHERE id = $16 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		p.Description,
		persistence.NullIfEmptyUUID(p.CategoryID), persistence.NullIfEmptyUUID(p.BrandID),
		p.UnitID, p.TaxID,
		p.CostUSD.String(), p.SalePrice.String(), p.SaleCurrency.String(),
		p.MinStock.String(), p.MaxStock.String(), p.Weight.String(),
		p.IsActive, p.IsService,
		time.Now().UTC(), p.UpdatedBy, p.ID,
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

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM products WHERE id = $1`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q, id)
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
	q := `SELECT ` + productColumns + productFrom + ` WHERE p.id = $1 AND p.deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	p, err := scanProduct(row)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *productRepository) GetBySKU(ctx context.Context, companyID uuid.UUID, sku string) (*product.Product, error) {
	q := `SELECT ` + productColumns + productFrom + ` WHERE p.company_id = $1 AND p.sku = $2 AND p.deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, sku)
	p, err := scanProduct(row)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *productRepository) GetByBarcode(ctx context.Context, companyID uuid.UUID, barcode string) (*product.Product, error) {
	q := `SELECT ` + productColumns + productFrom + ` WHERE p.company_id = $1 AND p.barcode = $2 AND p.deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID, barcode)
	p, err := scanProduct(row)
	if err != nil {
		return nil, err
	}
	return p, nil
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
	var (
		clauses = []string{"p.deleted_at IS NULL"}
		args    []any
	)
	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("p.company_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}
	if filter.CategoryID != nil {
		clauses = append(clauses, fmt.Sprintf("p.category_id = $%d", len(args)+1))
		args = append(args, *filter.CategoryID)
	}
	if filter.BrandID != nil {
		clauses = append(clauses, fmt.Sprintf("p.brand_id = $%d", len(args)+1))
		args = append(args, *filter.BrandID)
	}
	if filter.IsActive != nil {
		clauses = append(clauses, fmt.Sprintf("p.is_active = $%d", len(args)+1))
		args = append(args, *filter.IsActive)
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf(
			"(LOWER(p.description) LIKE LOWER($%d) OR LOWER(p.sku) LIKE LOWER($%d) OR LOWER(COALESCE(p.barcode, '')) LIKE LOWER($%d))",
			len(args)+1, len(args)+2, len(args)+3,
		))
		like := "%" + filter.Search + "%"
		args = append(args, like, like, like)
	}
	limit, offset := persistence.LimitOffset(filter.PageRequest, 25, 200)

	where := persistence.JoinClauses(clauses)
	var total int
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, "SELECT count(*) FROM products p WHERE "+where, args...).Scan(&total); err != nil {
		return repositories.Page[*product.Product]{}, persistence.Translate(err)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2
	args = append(args, limit, offset)
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		fmt.Sprintf("SELECT %s%s WHERE %s ORDER BY p.description LIMIT $%d OFFSET $%d",
			productColumns, productFrom, where, limitPos, offsetPos),
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

// ListUnits returns the units of measure available for a company,
// ordered by display name.
func (r *productRepository) ListUnits(ctx context.Context, companyID uuid.UUID) ([]*product.UnitOfMeasure, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT id, company_id, code, name, symbol, allows_decimals
		   FROM units_of_measure
		  WHERE company_id = $1
		  ORDER BY name`,
		companyID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*product.UnitOfMeasure, 0, 8)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		u := &product.UnitOfMeasure{}
		var symbol sql.NullString
		if err := r.Scan(&u.ID, &u.CompanyID, &u.Code, &u.Name, &symbol, &u.AllowsDecimals); err != nil {
			return persistence.Translate(err)
		}
		if symbol.Valid {
			u.Symbol = symbol.String
		}
		out = append(out, u)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// ListCategories returns the product categories available for a company,
// ordered by display name. Categories are hierarchical; the full tree is
// returned so the UI can present it however it likes.
func (r *productRepository) ListCategories(ctx context.Context, companyID uuid.UUID) ([]*product.ProductCategory, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT id, company_id, code, name, parent_id, path, depth,
		        created_at, updated_at, deleted_at, created_by, updated_by
		   FROM product_categories
		  WHERE company_id = $1 AND deleted_at IS NULL
		  ORDER BY name`,
		companyID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*product.ProductCategory, 0, 16)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		c := &product.ProductCategory{}
		var (
			parentID, path, createdBy, updatedBy sql.NullString
			deletedAt                            sql.NullTime
			code, name                           string
		)
		if err := r.Scan(&c.ID, &c.CompanyID, &code, &name, &parentID, &path, &c.Depth,
			&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy); err != nil {
			return persistence.Translate(err)
		}
		fillCategory(c, code, name, parentID, path, deletedAt, createdBy, updatedBy)
		out = append(out, c)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// ListBrands returns the product brands available for a company,
// ordered by display name.
func (r *productRepository) ListBrands(ctx context.Context, companyID uuid.UUID) ([]*product.ProductBrand, error) {
	rows, err := persistence.Q(ctx, r.q).QueryContext(ctx,
		`SELECT id, company_id, code, name,
		        created_at, updated_at, deleted_at, created_by, updated_by
		   FROM product_brands
		  WHERE company_id = $1 AND deleted_at IS NULL
		  ORDER BY name`,
		companyID)
	if err != nil {
		return nil, persistence.Translate(err)
	}
	out := make([]*product.ProductBrand, 0, 16)
	if err := persistence.ScanRows(rows, func(r *sql.Rows) error {
		b := &product.ProductBrand{}
		var (
			createdBy, updatedBy sql.NullString
			deletedAt            sql.NullTime
			code, name           string
		)
		if err := r.Scan(&b.ID, &b.CompanyID, &code, &name,
			&b.CreatedAt, &b.UpdatedAt, &deletedAt, &createdBy, &updatedBy); err != nil {
			return persistence.Translate(err)
		}
		fillBrand(b, code, name, deletedAt, createdBy, updatedBy)
		out = append(out, b)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// --- category CRUD ---

func (r *productRepository) CreateCategory(ctx context.Context, c *product.ProductCategory) error {
	const q = `INSERT INTO product_categories (
		id, company_id, code, name, parent_id, path, depth,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		c.ID, c.CompanyID, c.Code.String(), c.Name.String(),
		persistence.NullIfEmptyUUID(c.ParentID), c.Path, c.Depth,
		c.CreatedAt, c.UpdatedAt, persistence.NullIfZeroTime(c.DeletedAt), c.CreatedBy, c.UpdatedBy,
	)
	return persistence.Translate(err)
}

func (r *productRepository) UpdateCategory(ctx context.Context, c *product.ProductCategory) error {
	const q = `UPDATE product_categories SET
		code = $1, name = $2, parent_id = $3, path = $4, depth = $5,
		updated_at = $6, updated_by = $7
	 WHERE id = $8 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		c.Code.String(), c.Name.String(),
		persistence.NullIfEmptyUUID(c.ParentID), c.Path, c.Depth,
		time.Now().UTC(), c.UpdatedBy, c.ID,
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

func (r *productRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE product_categories SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
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

func (r *productRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*product.ProductCategory, error) {
	const q = `SELECT id, company_id, code, name, parent_id, path, depth,
		        created_at, updated_at, deleted_at, created_by, updated_by
		   FROM product_categories
		  WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanCategory(row)
}

// scanCategory is the *sql.Row variant for a product category row.
func scanCategory(row *sql.Row) (*product.ProductCategory, error) {
	c := &product.ProductCategory{}
	var (
		parentID, path, createdBy, updatedBy sql.NullString
		deletedAt                            sql.NullTime
		code, name                           string
	)
	if err := persistence.ScanRow(row, &c.ID, &c.CompanyID, &code, &name, &parentID, &path, &c.Depth,
		&c.CreatedAt, &c.UpdatedAt, &deletedAt, &createdBy, &updatedBy); err != nil {
		return nil, err
	}
	fillCategory(c, code, name, parentID, path, deletedAt, createdBy, updatedBy)
	return c, nil
}

// fillCategory converts the raw scanned columns into value objects.
func fillCategory(c *product.ProductCategory, code, name string,
	parentID, path sql.NullString, deletedAt sql.NullTime, createdBy, updatedBy sql.NullString,
) {
	if v, err := valueobjects.NewShortCode(code); err == nil {
		c.Code = v
	}
	if v, err := valueobjects.NewFullName(name); err == nil {
		c.Name = v
	}
	c.Path = path.String
	if parentID.Valid {
		id := persistence.ParseUUID(parentID.String)
		c.ParentID = &id
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		c.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		c.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		c.DeletedAt = &t
	}
}

// --- brand CRUD ---

func (r *productRepository) CreateBrand(ctx context.Context, b *product.ProductBrand) error {
	const q = `INSERT INTO product_brands (
		id, company_id, code, name,
		created_at, updated_at, deleted_at, created_by, updated_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		b.ID, b.CompanyID, b.Code.String(), b.Name.String(),
		b.CreatedAt, b.UpdatedAt, persistence.NullIfZeroTime(b.DeletedAt), b.CreatedBy, b.UpdatedBy,
	)
	return persistence.Translate(err)
}

func (r *productRepository) UpdateBrand(ctx context.Context, b *product.ProductBrand) error {
	const q = `UPDATE product_brands SET
		code = $1, name = $2, updated_at = $3, updated_by = $4
	 WHERE id = $5 AND deleted_at IS NULL`
	res, err := persistence.Q(ctx, r.q).ExecContext(ctx, q,
		b.Code.String(), b.Name.String(), time.Now().UTC(), b.UpdatedBy, b.ID,
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

func (r *productRepository) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE product_brands SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
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

func (r *productRepository) GetBrandByID(ctx context.Context, id uuid.UUID) (*product.ProductBrand, error) {
	const q = `SELECT id, company_id, code, name,
		        created_at, updated_at, deleted_at, created_by, updated_by
		   FROM product_brands
		  WHERE id = $1 AND deleted_at IS NULL`
	row := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, id)
	return scanBrand(row)
}

// scanBrand is the *sql.Row variant for a product brand row.
func scanBrand(row *sql.Row) (*product.ProductBrand, error) {
	b := &product.ProductBrand{}
	var (
		createdBy, updatedBy sql.NullString
		deletedAt            sql.NullTime
		code, name           string
	)
	if err := persistence.ScanRow(row, &b.ID, &b.CompanyID, &code, &name,
		&b.CreatedAt, &b.UpdatedAt, &deletedAt, &createdBy, &updatedBy); err != nil {
		return nil, err
	}
	fillBrand(b, code, name, deletedAt, createdBy, updatedBy)
	return b, nil
}

// fillBrand converts the raw scanned columns into value objects.
func fillBrand(b *product.ProductBrand, code, name string,
	deletedAt sql.NullTime, createdBy, updatedBy sql.NullString,
) {
	if v, err := valueobjects.NewShortCode(code); err == nil {
		b.Code = v
	}
	if v, err := valueobjects.NewFullName(name); err == nil {
		b.Name = v
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		b.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		b.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		b.DeletedAt = &t
	}
}

// --- scanning helpers ---

// scanProduct is the *sql.Row variant.
func scanProduct(row *sql.Row) (*product.Product, error) {
	p := &product.Product{}
	var (
		categoryID, brandID, createdBy, updatedBy sql.NullString
		barcode                                   sql.NullString
		deletedAt                                 sql.NullTime
		categoryName, brandName, unitName         sql.NullString
		sku, description, currency                string
		costUSD, salePrice, minStock, maxStock, weight string
		isActive, isService                       bool
	)
	err := persistence.ScanRow(row,
		&p.ID, &p.CompanyID, &sku, &barcode, &description,
		&categoryID, &brandID, &p.UnitID, &p.TaxID,
		&costUSD, &salePrice, &currency,
		&minStock, &maxStock, &weight,
		&isActive, &isService,
		&p.CreatedAt, &p.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
		&categoryName, &brandName, &unitName,
	)
	if err != nil {
		return nil, err
	}
	fillProduct(p, sku, barcode, description, categoryID, brandID, categoryName, brandName, unitName, currency, costUSD, salePrice, minStock, maxStock, weight, isActive, isService, deletedAt, createdBy, updatedBy)
	return p, nil
}

// scanProductFromRows is the *sql.Rows variant.
func scanProductFromRows(rows *sql.Rows) (*product.Product, error) {
	p := &product.Product{}
	var (
		categoryID, brandID, createdBy, updatedBy sql.NullString
		barcode                                   sql.NullString
		deletedAt                                 sql.NullTime
		categoryName, brandName, unitName         sql.NullString
		sku, description, currency                string
		costUSD, salePrice, minStock, maxStock, weight string
		isActive, isService                       bool
	)
	if err := rows.Scan(
		&p.ID, &p.CompanyID, &sku, &barcode, &description,
		&categoryID, &brandID, &p.UnitID, &p.TaxID,
		&costUSD, &salePrice, &currency,
		&minStock, &maxStock, &weight,
		&isActive, &isService,
		&p.CreatedAt, &p.UpdatedAt, &deletedAt, &createdBy, &updatedBy,
		&categoryName, &brandName, &unitName,
	); err != nil {
		return nil, persistence.Translate(err)
	}
	fillProduct(p, sku, barcode, description, categoryID, brandID, categoryName, brandName, unitName, currency, costUSD, salePrice, minStock, maxStock, weight, isActive, isService, deletedAt, createdBy, updatedBy)
	return p, nil
}

// fillProduct converts the raw scanned columns into value objects. It is
// shared by both scan variants.
func fillProduct(p *product.Product,
	sku string, barcode sql.NullString, description string,
	categoryID, brandID, categoryName, brandName, unitName sql.NullString,
	currency, costUSD, salePrice, minStock, maxStock, weight string,
	isActive, isService bool,
	deletedAt sql.NullTime, createdBy, updatedBy sql.NullString,
) {
	p.CategoryName = categoryName.String
	p.BrandName = brandName.String
	p.UnitName = unitName.String
	v, _ := valueobjects.NewSKU(sku)
	p.SKU = v
	if barcode.Valid {
		bc, _ := valueobjects.OptionalBarcode(barcode.String)
		p.Barcode = bc
	}
	if categoryID.Valid {
		id := persistence.ParseUUID(categoryID.String)
		p.CategoryID = &id
	}
	if brandID.Valid {
		id := persistence.ParseUUID(brandID.String)
		p.BrandID = &id
	}
	if createdBy.Valid {
		id := persistence.ParseUUID(createdBy.String)
		p.CreatedBy = &id
	}
	if updatedBy.Valid {
		id := persistence.ParseUUID(updatedBy.String)
		p.UpdatedBy = &id
	}
	if !deletedAt.Time.IsZero() {
		t := deletedAt.Time
		p.DeletedAt = &t
	}
	p.IsActive = isActive
	p.IsService = isService
	p.Description = description
	if v, err := persistence.ParseMoney(costUSD); err == nil {
		p.CostUSD = v
	}
	if v, err := persistence.ParseMoney(salePrice); err == nil {
		p.SalePrice = v
	}
	if q, err := valueobjects.QuantityFromString(minStock); err == nil {
		p.MinStock = q
	}
	if q, err := valueobjects.QuantityFromString(maxStock); err == nil {
		p.MaxStock = q
	}
	if q, err := valueobjects.QuantityFromString(weight); err == nil {
		p.Weight = q
	}
	if cc, err := valueobjects.NewCurrencyCode(currency); err == nil {
		p.SaleCurrency = cc
	}
}
