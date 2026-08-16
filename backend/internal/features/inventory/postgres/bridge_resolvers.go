package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/persistence"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/inventory"
)

var _ inventory.WarehouseResolver = (*warehouseResolver)(nil)

type warehouseResolver struct {
	q persistence.Querier
}

// NewWarehouseResolver returns a WarehouseResolver backed by the
// warehouses table.
func NewWarehouseResolver(db *sql.DB) *warehouseResolver {
	return &warehouseResolver{q: persistence.FromDB(db)}
}

func (r *warehouseResolver) DefaultWarehouseID(ctx context.Context, companyID uuid.UUID) (uuid.UUID, error) {
	var id string
	q := `SELECT id FROM warehouses
		WHERE company_id = $1 AND is_default = TRUE AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY created_at LIMIT 1`
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, companyID).Scan(&id); err != nil {
		if persistence.IsPgNoRows(err) {
			return uuid.Nil, errors.Wrap(errors.ErrRequired, repositories.ErrNotFound)
		}
		return uuid.Nil, persistence.Translate(err)
	}
	return persistence.ParseUUID(id), nil
}

var _ inventory.ProductClassifier = (*productClassifier)(nil)

type productClassifier struct {
	q persistence.Querier
}

// NewProductClassifier returns a ProductClassifier backed by the
// products table.
func NewProductClassifier(db *sql.DB) *productClassifier {
	return &productClassifier{q: persistence.FromDB(db)}
}

func (r *productClassifier) IsService(ctx context.Context, productID uuid.UUID) (bool, error) {
	var isService bool
	q := `SELECT is_service FROM products WHERE id = $1 AND deleted_at IS NULL`
	if err := persistence.Q(ctx, r.q).QueryRowContext(ctx, q, productID).Scan(&isService); err != nil {
		if persistence.IsPgNoRows(err) {
			return false, persistence.Translate(repositories.ErrNotFound)
		}
		return false, persistence.Translate(err)
	}
	return isService, nil
}
