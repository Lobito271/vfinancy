// Package product implements the business logic for the product
// aggregate: lifecycle, price/cost changes, margin math.
package product

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
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
