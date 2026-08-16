package bindings

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/product"
	"vfinancy/backend/internal/shared/apperrors"
)

// ProductDTO is the serializable view of a product.
type ProductDTO struct {
	ID           string `json:"id"`
	SKU          string `json:"sku"`
	Barcode      string `json:"barcode"`
	Description  string `json:"description"`
	CategoryID   string `json:"categoryId"`
	BrandID      string `json:"brandId"`
	UnitID       string `json:"unitId"`
	TaxID        string `json:"taxId"`
	Category     string `json:"category"`
	Brand        string `json:"brand"`
	Unit         string `json:"unit"`
	Cost         string `json:"cost"`
	SalePrice    string `json:"salePrice"`
	SaleCurrency string `json:"saleCurrency"`
	MinStock     string `json:"minStock"`
	MaxStock     string `json:"maxStock"`
	Weight       string `json:"weight"`
	IsActive     bool   `json:"isActive"`
	IsService    bool   `json:"isService"`
	CreatedAt    string `json:"createdAt"`
}

func toProductDTO(p *product.Product) *ProductDTO {
	return &ProductDTO{
		ID:           p.ID.String(),
		SKU:          p.SKU.String(),
		Barcode:      p.Barcode.String(),
		Description:  p.Description,
		CategoryID:   uuidPtrString(p.CategoryID),
		BrandID:      uuidPtrString(p.BrandID),
		UnitID:       p.UnitID.String(),
		TaxID:        p.TaxID.String(),
		Category:     p.CategoryName,
		Brand:        p.BrandName,
		Unit:         p.UnitName,
		Cost:         p.Cost.String(),
		SalePrice:    p.SalePrice.String(),
		SaleCurrency: p.SaleCurrency.String(),
		MinStock:     p.MinStock.String(),
		MaxStock:     p.MaxStock.String(),
		Weight:       p.Weight.String(),
		IsActive:     p.IsActive,
		IsService:    p.IsService,
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
	}
}

// ListProductsRequest filters the product catalog.
type ListProductsRequest struct {
	Search string `json:"search"`
	Status string `json:"status"`
	PaginationRequest
}

// ListProducts returns paged products.
func (a *App) ListProducts(req ListProductsRequest) (PageResult, error) {
	var isActive *bool
	switch req.Status {
	case "active":
		v := true
		isActive = &v
	case "inactive":
		v := false
		isActive = &v
	}
	filter := product.ProductFilter{
		CompanyID:   &demoCompanyID,
		Search:      req.Search,
		IsActive:    isActive,
		PageRequest: req.toPageRequest(),
	}
	page, err := a.productsSvc.List(a.Context(), filter)
	if err != nil {
		return PageResult{}, err
	}
	items := make([]*ProductDTO, 0, len(page.Items))
	for _, p := range page.Items {
		items = append(items, toProductDTO(p))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// GetProduct returns a single product by ID.
func (a *App) GetProduct(id string) (*ProductDTO, error) {
	pid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	p, err := a.productsSvc.GetByID(a.Context(), pid)
	if err != nil {
		return nil, err
	}
	return toProductDTO(p), nil
}

// CreateProductRequest creates a product.
type CreateProductRequest struct {
	SKU          string `json:"sku"`
	Barcode      string `json:"barcode"`
	Description  string `json:"description"`
	CategoryID   string `json:"categoryId"`
	BrandID      string `json:"brandId"`
	UnitID       string `json:"unitId"`
	TaxID        string `json:"taxId"`
	Cost         string `json:"cost"`
	SalePrice    string `json:"salePrice"`
	SaleCurrency string `json:"saleCurrency"`
	MinStock     string `json:"minStock"`
	MaxStock     string `json:"maxStock"`
	Weight       string `json:"weight"`
	IsService    bool   `json:"isService"`
}

// CreateProduct persists a new product.
func (a *App) CreateProduct(req CreateProductRequest) (*ProductDTO, error) {
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		return nil, err
	}
	brandID, err := parseOptionalUUID(req.BrandID)
	if err != nil {
		return nil, err
	}
	unitID, err := parseOptionalUUID(req.UnitID)
	if err != nil {
		return nil, err
	}
	if unitID == nil {
		return nil, apperrors.Errorf(apperrors.ErrValidation, "unit id is required")
	}
	taxID, err := parseOptionalUUID(req.TaxID)
	if err != nil {
		return nil, err
	}
	if taxID == nil {
		return nil, apperrors.Errorf(apperrors.ErrValidation, "tax id is required")
	}
	currency, err := currencyOrDefault(req.SaleCurrency)
	if err != nil {
		return nil, err
	}
	barcode, err := valueobjects.OptionalBarcode(req.Barcode)
	if err != nil {
		return nil, err
	}
	cost, err := moneyOrZero(req.Cost)
	if err != nil {
		return nil, err
	}
	salePrice, err := moneyOrZero(req.SalePrice)
	if err != nil {
		return nil, err
	}
	minStock, err := quantityOrZero(req.MinStock)
	if err != nil {
		return nil, err
	}
	maxStock, err := quantityOrZero(req.MaxStock)
	if err != nil {
		return nil, err
	}
	weight, err := quantityOrZero(req.Weight)
	if err != nil {
		return nil, err
	}
	sku, err := valueobjects.NewSKU(req.SKU)
	if err != nil {
		return nil, err
	}
	in := product.CreateInput{
		CompanyID:    demoCompanyID,
		SKU:          sku,
		Barcode:      barcode,
		Description:  req.Description,
		CategoryID:   categoryID,
		BrandID:      brandID,
		UnitID:       *unitID,
		TaxID:        *taxID,
		Cost:         cost,
		SalePrice:    salePrice,
		SaleCurrency: currency,
		MinStock:     minStock,
		MaxStock:     maxStock,
		Weight:       weight,
		IsService:    req.IsService,
	}
	p, err := a.productsSvc.Create(a.Context(), in)
	if err != nil {
		return nil, err
	}
	return toProductDTO(p), nil
}

// UpdateProductRequest updates a product.
type UpdateProductRequest struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	CategoryID  string `json:"categoryId"`
	BrandID     string `json:"brandId"`
	UnitID      string `json:"unitId"`
	Cost        string `json:"cost"`
	SalePrice   string `json:"salePrice"`
	MinStock    string `json:"minStock"`
	MaxStock    string `json:"maxStock"`
	IsActive    *bool  `json:"isActive"`
}

// UpdateProduct updates the mutable product fields.
func (a *App) UpdateProduct(req UpdateProductRequest) (*ProductDTO, error) {
	pid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	in := product.UpdateInput{ID: pid, Description: req.Description, IsActive: req.IsActive}
	if req.CategoryID != "" {
		id, err := parseOptionalUUID(req.CategoryID)
		if err != nil {
			return nil, err
		}
		in.CategoryID = id
	}
	if req.BrandID != "" {
		id, err := parseOptionalUUID(req.BrandID)
		if err != nil {
			return nil, err
		}
		in.BrandID = id
	}
	if req.UnitID != "" {
		id, err := parseOptionalUUID(req.UnitID)
		if err != nil {
			return nil, err
		}
		in.UnitID = id
	}
	if req.Cost != "" {
		cost, err := valueobjects.MoneyFromString(req.Cost)
		if err != nil {
			return nil, err
		}
		in.Cost = &cost
	}
	if req.SalePrice != "" {
		price, err := valueobjects.MoneyFromString(req.SalePrice)
		if err != nil {
			return nil, err
		}
		in.SalePrice = &price
	}
	if req.MinStock != "" && req.MaxStock != "" {
		min, err := valueobjects.QuantityFromString(req.MinStock)
		if err != nil {
			return nil, err
		}
		max, err := valueobjects.QuantityFromString(req.MaxStock)
		if err != nil {
			return nil, err
		}
		in.MinStock = &min
		in.MaxStock = &max
	}
	p, err := a.productsSvc.Update(a.Context(), in)
	if err != nil {
		return nil, err
	}
	return toProductDTO(p), nil
}

// RemoveProduct hard-deletes a product. Referenced products (e.g. ones
// already in a sale) are rejected with a friendly error.
func (a *App) RemoveProduct(id string) error {
	pid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return a.productsSvc.Delete(a.Context(), pid)
}

// --- helpers ---

func moneyOrZero(s string) (valueobjects.Money, error) {
	if strings.TrimSpace(s) == "" {
		return valueobjects.Zero(), nil
	}
	return valueobjects.MoneyFromString(s)
}

func quantityOrZero(s string) (valueobjects.Quantity, error) {
	if strings.TrimSpace(s) == "" {
		return valueobjects.ZeroQuantity(), nil
	}
	return valueobjects.QuantityFromString(s)
}

func currencyOrDefault(s string) (valueobjects.CurrencyCode, error) {
	if strings.TrimSpace(s) == "" {
		return valueobjects.MustCurrencyCode("PEN"), nil
	}
	return valueobjects.NewCurrencyCode(s)
}
