// Package product implements the business logic for the product
// aggregate: creation, price/cost changes, lifecycle.
package product

import (
	"strings"
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// DefaultUnitCode is the unit of measure assigned to new products.
const DefaultUnitCode = "Unidad"

// Product is the catalog entity: the static description of a sellable
// item. Per-warehouse stock lives in the inventory feature.
type Product struct {
	ID          uuid.UUID
	SKU         valueobjects.SKU
	Description string
	UnitCode    string
	CostUSD     valueobjects.Money
	SalePrice   valueobjects.Money
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedBy   string
	UpdatedBy   string
}

// NewProduct validates inputs and constructs a Product with an
// auto-generated SKU, the default unit and active status.
func NewProduct(description string, costUSD, salePrice valueobjects.Money) (*Product, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("description is required"))
	}
	if costUSD.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("cost usd cannot be negative"))
	}
	if salePrice.IsNegative() {
		return nil, derrors.Wrap(derrors.ErrNegativeMoney, errField("sale price cannot be negative"))
	}
	now := time.Now().UTC()
	id := uuid.New()
	sku, err := valueobjects.NewSKU("P-" + strings.ReplaceAll(id.String(), "-", "")[:8])
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          id,
		SKU:         sku,
		Description: description,
		UnitCode:    DefaultUnitCode,
		CostUSD:     costUSD,
		SalePrice:   salePrice,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Touch stamps UpdatedAt with the current UTC time.
func (p *Product) Touch() { p.UpdatedAt = time.Now().UTC() }

// Validate checks the product invariants: non-blank description and
// unit code, non-negative money fields.
func (p *Product) Validate() error {
	if strings.TrimSpace(p.Description) == "" {
		return derrors.Wrap(derrors.ErrRequired, errField("description is required"))
	}
	if strings.TrimSpace(p.UnitCode) == "" {
		return derrors.Wrap(derrors.ErrRequired, errField("unit code is required"))
	}
	if len(p.UnitCode) > 20 {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("unit code is too long"))
	}
	if p.CostUSD.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("cost usd cannot be negative"))
	}
	if p.SalePrice.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("sale price cannot be negative"))
	}
	return nil
}
