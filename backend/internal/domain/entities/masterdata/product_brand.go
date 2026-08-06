package masterdata

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/valueobjects"
)

// ProductBrand is a flat (non-hierarchical) brand reference.
type ProductBrand struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Code      valueobjects.ShortCode
	Name      valueobjects.FullName
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
}

// NewProductBrand validates and constructs a brand.
func NewProductBrand(now time.Time, companyID uuid.UUID, code valueobjects.ShortCode, name valueobjects.FullName) *ProductBrand {
	return &ProductBrand{
		ID:        uuid.New(),
		CompanyID: companyID,
		Code:      code,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
