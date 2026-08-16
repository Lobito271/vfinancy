package product

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Product is the catalog entity. It is the *static* description of a
// sellable item; per-warehouse stock lives in InventoryBatch (and is
// computed from movements, never stored on the product).
type Product struct {
	ID           uuid.UUID
	CompanyID    uuid.UUID
	SKU          valueobjects.SKU
	Barcode      valueobjects.Barcode
	Description  string
	CategoryID   *uuid.UUID
	BrandID      *uuid.UUID
	UnitID       uuid.UUID
	TaxID        uuid.UUID
	Cost         valueobjects.Money
	SalePrice    valueobjects.Money
	SaleCurrency valueobjects.CurrencyCode
	MinStock     valueobjects.Quantity
	MaxStock     valueobjects.Quantity
	Weight       valueobjects.Quantity
	IsActive     bool
	IsService    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	CreatedBy    *uuid.UUID
	UpdatedBy    *uuid.UUID

	// Read-model enrichment populated by the repository via joins; never
	// persisted. Used by the UI to render category / brand / unit names.
	CategoryName string
	BrandName    string
	UnitName     string
}

// NewProductOptions is the input to NewProduct.
type NewProductOptions struct {
	CompanyID    uuid.UUID
	SKU          valueobjects.SKU
	Barcode      valueobjects.Barcode
	Description  string
	CategoryID   *uuid.UUID
	BrandID      *uuid.UUID
	UnitID       uuid.UUID
	TaxID        uuid.UUID
	Cost         valueobjects.Money
	SalePrice    valueobjects.Money
	SaleCurrency valueobjects.CurrencyCode
	MinStock     valueobjects.Quantity
	MaxStock     valueobjects.Quantity
	Weight       valueobjects.Quantity
	IsService    bool
}

// NewProduct validates and constructs a Product.
func NewProduct(now time.Time, opts NewProductOptions) (*Product, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if opts.UnitID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("unit id is required"))
	}
	if opts.TaxID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("tax id is required"))
	}
	if opts.Cost.IsNegative() {
		return nil, errors.Wrap(errors.ErrNegativeMoney, errField("cost cannot be negative"))
	}
	if opts.SalePrice.IsNegative() {
		return nil, errors.Wrap(errors.ErrNegativeMoney, errField("sale price cannot be negative"))
	}
	if opts.MaxStock.LessThan(opts.MinStock) {
		return nil, errors.Wrap(errors.ErrOutOfRange, errField("max stock must be >= min stock"))
	}
	if opts.Description == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("description is required"))
	}
	p := &Product{
		ID:           uuid.New(),
		CompanyID:    opts.CompanyID,
		SKU:          opts.SKU,
		Barcode:      opts.Barcode,
		Description:  opts.Description,
		CategoryID:   opts.CategoryID,
		BrandID:      opts.BrandID,
		UnitID:       opts.UnitID,
		TaxID:        opts.TaxID,
		Cost:         opts.Cost,
		SalePrice:    opts.SalePrice,
		SaleCurrency: opts.SaleCurrency,
		MinStock:     opts.MinStock,
		MaxStock:     opts.MaxStock,
		Weight:       opts.Weight,
		IsActive:     true,
		IsService:    opts.IsService,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if p.IsService {
		p.MinStock = valueobjects.ZeroQuantity()
		p.MaxStock = valueobjects.ZeroQuantity()
	}
	return p, nil
}

// ChangeSalePrice updates the sale price. Negative is rejected.
func (p *Product) ChangeSalePrice(price valueobjects.Money) error {
	if price.IsNegative() {
		return errors.Wrap(errors.ErrNegativeMoney, errField("sale price cannot be negative"))
	}
	p.SalePrice = price
	return nil
}

// ChangeCost updates the standard cost. Negative is rejected.
func (p *Product) ChangeCost(cost valueobjects.Money) error {
	if cost.IsNegative() {
		return errors.Wrap(errors.ErrNegativeMoney, errField("cost cannot be negative"))
	}
	p.Cost = cost
	return nil
}

// CalculateMargin returns the unit margin (sale price - cost) as Money.
func (p *Product) CalculateMargin() valueobjects.Money {
	return p.SalePrice.Sub(p.Cost)
}

// MarginPercent returns the margin as a percentage of the sale price.
// Zero sale price is reported as 0% to avoid division by zero.
func (p *Product) MarginPercent() float64 {
	if p.SalePrice.IsZero() {
		return 0
	}
	margin := p.SalePrice.Decimal().Sub(p.Cost.Decimal()).
		Div(p.SalePrice.Decimal()).
		Mul(decimal.NewFromInt(100))
	f, _ := margin.Float64()
	return f
}

// ChangeStockLimits updates the min and max stock thresholds.
func (p *Product) ChangeStockLimits(min, max valueobjects.Quantity) error {
	if max.LessThan(min) {
		return errors.Wrap(errors.ErrOutOfRange, errField("max stock must be >= min stock"))
	}
	p.MinStock = min
	p.MaxStock = max
	return nil
}

// ChangeDescription updates the description.
func (p *Product) ChangeDescription(s string) error {
	if s == "" {
		return errors.Wrap(errors.ErrRequired, errField("description is required"))
	}
	p.Description = s
	return nil
}

// Activate / Deactivate.
func (p *Product) Activate()   { p.IsActive = true }
func (p *Product) Deactivate() { p.IsActive = false }
