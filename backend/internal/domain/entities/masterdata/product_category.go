package masterdata

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// ProductCategory is a hierarchical category (1:N self-referencing).
// The path field materializes the dotted path for fast lookups
// (e.g. "1.2.5") — populated by the application layer at insert time.
type ProductCategory struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Code      valueobjects.ShortCode
	Name      valueobjects.FullName
	ParentID  *uuid.UUID
	Path      string
	Depth     int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
}

// NewProductCategoryOptions is the input to NewProductCategory.
type NewProductCategoryOptions struct {
	CompanyID uuid.UUID
	Code      valueobjects.ShortCode
	Name      valueobjects.FullName
	ParentID  *uuid.UUID
	Depth     int  // set by the application; 0 for root, >0 otherwise
}

// NewProductCategory validates and constructs a category.
func NewProductCategory(now time.Time, opts NewProductCategoryOptions) (*ProductCategory, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if opts.Depth < 0 {
		return nil, errors.Wrap(errors.ErrOutOfRange, errField("depth cannot be negative"))
	}
	return &ProductCategory{
		ID:        uuid.New(),
		CompanyID: opts.CompanyID,
		Code:      opts.Code,
		Name:      opts.Name,
		ParentID:  opts.ParentID,
		Depth:     opts.Depth,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Rename updates the display name.
func (c *ProductCategory) Rename(name valueobjects.FullName) { c.Name = name }
