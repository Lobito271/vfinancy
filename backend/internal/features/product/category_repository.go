package product

import (
	"context"

	"github.com/google/uuid"

)

// CategoryRepository persists product categories. Categories form a
// 1:N self-referencing tree; the path/depth fields are maintained by
// the application layer.
type CategoryRepository interface {
	Create(ctx context.Context, c *ProductCategory) error
	Update(ctx context.Context, c *ProductCategory) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*ProductCategory, error)
	List(ctx context.Context, companyID uuid.UUID) ([]*ProductCategory, error)
	// ListChildren returns the direct children of a parent (or root
	// categories if parentID is nil).
	ListChildren(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID) ([]*ProductCategory, error)
}
