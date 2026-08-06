package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/masterdata"
)

// CategoryRepository persists product categories. Categories form a
// 1:N self-referencing tree; the path/depth fields are maintained by
// the application layer.
type CategoryRepository interface {
	Create(ctx context.Context, c *masterdata.ProductCategory) error
	Update(ctx context.Context, c *masterdata.ProductCategory) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*masterdata.ProductCategory, error)
	List(ctx context.Context, companyID uuid.UUID) ([]*masterdata.ProductCategory, error)
	// ListChildren returns the direct children of a parent (or root
	// categories if parentID is nil).
	ListChildren(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID) ([]*masterdata.ProductCategory, error)
}
