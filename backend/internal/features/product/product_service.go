// Package product implements the business logic for the product
// aggregate: lifecycle, price/cost changes, margin math.
package product

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/shared/apperrors"
	"vfinancy/backend/internal/shared/logger"
)

// ProductService owns the product slice.
type ProductService struct {
	repo ProductRepository
	txm  repositories.TransactionManager
	log  *logger.Logger
}

// New returns a ProductService ready for use.
func New(repo ProductRepository, txm repositories.TransactionManager, log *logger.Logger) *ProductService {
	return &ProductService{repo: repo, txm: txm, log: log}
}

// CreateInput is the payload for CreateProduct. SKU is required; cost
// and price default to zero if unset.
type CreateInput struct {
	CompanyID    uuid.UUID
	SKU          valueobjects.SKU
	Barcode      valueobjects.Barcode // optional
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
	Weight       valueobjects.Quantity // optional
	IsService    bool
}

// CreateProduct persists a new product. All validation happens in the
// domain constructor.
func (s *ProductService) Create(ctx context.Context, in CreateInput) (*Product, error) {
	var out *Product
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := NewProduct(time.Now().UTC(), NewProductOptions{
			CompanyID:    in.CompanyID,
			SKU:          in.SKU,
			Barcode:      in.Barcode,
			Description:  in.Description,
			CategoryID:   in.CategoryID,
			BrandID:      in.BrandID,
			UnitID:       in.UnitID,
			TaxID:        in.TaxID,
			Cost:         in.Cost,
			SalePrice:    in.SalePrice,
			SaleCurrency: in.SaleCurrency,
			MinStock:     in.MinStock,
			MaxStock:     in.MaxStock,
			Weight:       in.Weight,
			IsService:    in.IsService,
		})
		if err != nil {
			return err
		}
		if err := s.repo.Create(ctx, p); err != nil {
			return err
		}
		out = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("product created", "product_id", out.ID, "sku", out.SKU.String())
	return out, nil
}

// UpdateCost changes the product's standard cost.
func (s *ProductService) UpdateCost(ctx context.Context, id uuid.UUID, cost valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := p.ChangeCost(cost); err != nil {
			return err
		}
		return s.repo.Update(ctx, p)
	})
	if err != nil {
		return err
	}
	s.log.Info("product cost updated", "product_id", id, "new_cost", cost)
	return nil
}

// UpdateSalePrice changes the product's sale price. The new margin is
// logged for visibility.
func (s *ProductService) UpdateSalePrice(ctx context.Context, id uuid.UUID, price valueobjects.Money) error {
	var marginPct float64
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := p.ChangeSalePrice(price); err != nil {
			return err
		}
		marginPct = p.MarginPercent()
		return s.repo.Update(ctx, p)
	})
	if err != nil {
		return err
	}
	s.log.Info("product sale price updated", "product_id", id, "new_price", price, "margin_pct", marginPct)
	return nil
}

// UpdateStockLimits changes the min and max stock thresholds.
func (s *ProductService) UpdateStockLimits(ctx context.Context, id uuid.UUID, min, max valueobjects.Quantity) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := p.ChangeStockLimits(min, max); err != nil {
			return err
		}
		return s.repo.Update(ctx, p)
	})
	if err != nil {
		return err
	}
	s.log.Info("product stock limits updated", "product_id", id, "min", min, "max", max)
	return nil
}

// Activate / Deactivate.
func (s *ProductService) Activate(ctx context.Context, id uuid.UUID) error {
	return s.mutateStatus(ctx, id, true)
}

func (s *ProductService) Deactivate(ctx context.Context, id uuid.UUID) error {
	return s.mutateStatus(ctx, id, false)
}

func (s *ProductService) mutateStatus(ctx context.Context, id uuid.UUID, active bool) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if active {
			p.Activate()
		} else {
			p.Deactivate()
		}
		return s.repo.Update(ctx, p)
	})
	if err != nil {
		return err
	}
	if active {
		s.log.Info("product activated", "product_id", id)
	} else {
		s.log.Info("product deactivated", "product_id", id)
	}
	return nil
}

// Delete physically removes the product. It fails with a friendly
// error if the product is referenced by other records (sales,
// inventory, purchases, ...), in which case the caller should edit the
// record and set it to Inactive instead.
func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, repositories.ErrForeignKey) {
		return apperrors.Errorf(apperrors.ErrConflict,
			"no se puede eliminar porque tiene transacciones asociadas. Por favor, edite el registro y cambie su estado a Inactivo")
	}
	if err != nil {
		return err
	}
	s.log.Info("product deleted", "product_id", id)
	return nil
}

// UpdateInput is the payload for UpdateProduct. Only the non-nil /
// non-zero fields are applied.
type UpdateInput struct {
	ID          uuid.UUID
	Description string
	CategoryID  *uuid.UUID
	BrandID     *uuid.UUID
	UnitID      *uuid.UUID
	Cost        *valueobjects.Money
	SalePrice   *valueobjects.Money
	MinStock    *valueobjects.Quantity
	MaxStock    *valueobjects.Quantity
	IsActive    *bool
}

// UpdateProduct applies the requested changes in a single transaction.
func (s *ProductService) Update(ctx context.Context, in UpdateInput) (*Product, error) {
	var out *Product
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.repo.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if in.Description != "" {
			if err := p.ChangeDescription(in.Description); err != nil {
				return err
			}
		}
		if in.Cost != nil {
			if err := p.ChangeCost(*in.Cost); err != nil {
				return err
			}
		}
		if in.SalePrice != nil {
			if err := p.ChangeSalePrice(*in.SalePrice); err != nil {
				return err
			}
		}
		if in.MinStock != nil && in.MaxStock != nil {
			if err := p.ChangeStockLimits(*in.MinStock, *in.MaxStock); err != nil {
				return err
			}
		}
		if in.CategoryID != nil {
			p.CategoryID = in.CategoryID
		}
		if in.BrandID != nil {
			p.BrandID = in.BrandID
		}
		if in.UnitID != nil {
			p.UnitID = *in.UnitID
		}
		if in.IsActive != nil {
			if *in.IsActive {
				p.Activate()
			} else {
				p.Deactivate()
			}
		}
		if err := s.repo.Update(ctx, p); err != nil {
			return err
		}
		out = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("product updated", "product_id", out.ID)
	return out, nil
}

// List returns a page of products matching the filter.
func (s *ProductService) List(ctx context.Context, filter ProductFilter) (repositories.Page[*Product], error) {
	return s.repo.List(ctx, filter)
}

// GetByID returns a single product.
func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	return s.repo.GetByID(ctx, id)
}

// ListUnits returns the units of measure available for a company.
func (s *ProductService) ListUnits(ctx context.Context, companyID uuid.UUID) ([]*UnitOfMeasure, error) {
	return s.repo.ListUnits(ctx, companyID)
}

// ListCategories returns the product categories available for a company.
func (s *ProductService) ListCategories(ctx context.Context, companyID uuid.UUID) ([]*ProductCategory, error) {
	return s.repo.ListCategories(ctx, companyID)
}

// ListBrands returns the product brands available for a company.
func (s *ProductService) ListBrands(ctx context.Context, companyID uuid.UUID) ([]*ProductBrand, error) {
	return s.repo.ListBrands(ctx, companyID)
}

// --- Category CRUD ---

// CategoryInput is the payload for CreateCategory.
type CategoryInput struct {
	CompanyID uuid.UUID
	Code      valueobjects.ShortCode
	Name      valueobjects.FullName
}

// CreateCategory persists a new root category. The dotted path is
// materialized as the category code (the same convention the seed data
// uses for top-level categories).
func (s *ProductService) CreateCategory(ctx context.Context, in CategoryInput) (*ProductCategory, error) {
	var out *ProductCategory
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := NewProductCategory(time.Now().UTC(), NewProductCategoryOptions{
			CompanyID: in.CompanyID,
			Code:      in.Code,
			Name:      in.Name,
			Depth:     1,
		})
		if err != nil {
			return err
		}
		c.Path = c.Code.String()
		if err := s.repo.CreateCategory(ctx, c); err != nil {
			return err
		}
		out = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("category created", "category_id", out.ID, "code", out.Code.String())
	return out, nil
}

// CategoryUpdateInput is the payload for UpdateCategory.
type CategoryUpdateInput struct {
	ID   uuid.UUID
	Code valueobjects.ShortCode
	Name valueobjects.FullName
}

// UpdateCategory renames a category and updates its code.
func (s *ProductService) UpdateCategory(ctx context.Context, in CategoryUpdateInput) (*ProductCategory, error) {
	var out *ProductCategory
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.repo.GetCategoryByID(ctx, in.ID)
		if err != nil {
			return err
		}
		c.Code = in.Code
		c.Rename(in.Name)
		c.Path = in.Code.String()
		if err := s.repo.UpdateCategory(ctx, c); err != nil {
			return err
		}
		out = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("category updated", "category_id", out.ID)
	return out, nil
}

// DeleteCategory soft-deletes a category (sets deleted_at).
func (s *ProductService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return err
	}
	s.log.Info("category deleted", "category_id", id)
	return nil
}

// --- Brand CRUD ---

// BrandInput is the payload for CreateBrand / UpdateBrand.
type BrandInput struct {
	ID        uuid.UUID // empty for CreateBrand
	CompanyID uuid.UUID // ignored for UpdateBrand
	Code      valueobjects.ShortCode
	Name      valueobjects.FullName
}

// CreateBrand persists a new brand.
func (s *ProductService) CreateBrand(ctx context.Context, in BrandInput) (*ProductBrand, error) {
	var out *ProductBrand
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		b := NewProductBrand(time.Now().UTC(), in.CompanyID, in.Code, in.Name)
		if err := s.repo.CreateBrand(ctx, b); err != nil {
			return err
		}
		out = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("brand created", "brand_id", out.ID, "code", out.Code.String())
	return out, nil
}

// UpdateBrand renames a brand and updates its code.
func (s *ProductService) UpdateBrand(ctx context.Context, in BrandInput) (*ProductBrand, error) {
	var out *ProductBrand
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		b, err := s.repo.GetBrandByID(ctx, in.ID)
		if err != nil {
			return err
		}
		b.Code = in.Code
		b.Name = in.Name
		if err := s.repo.UpdateBrand(ctx, b); err != nil {
			return err
		}
		out = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("brand updated", "brand_id", out.ID)
	return out, nil
}

// DeleteBrand soft-deletes a brand (sets deleted_at).
func (s *ProductService) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteBrand(ctx, id); err != nil {
		return err
	}
	s.log.Info("brand deleted", "brand_id", id)
	return nil
}

// CalculateMargin returns the unit margin and percentage for a given
// price and cost. Used by the UI to show margin at quote time.
func CalculateMargin(price, cost valueobjects.Money) (valueobjects.Money, float64) {
	margin := price.Sub(cost)
	pct := 0.0
	if !price.IsZero() {
		costDec := cost.Decimal()
		priceDec := price.Decimal()
		hundred := decimal.NewFromInt(100)
		marginDec := priceDec.Sub(costDec).Div(priceDec).Mul(hundred)
		pctF, _ := marginDec.Float64()
		pct = pctF
	}
	return margin, pct
}
