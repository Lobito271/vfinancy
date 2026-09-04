package bindings

import (
	"errors"
	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/product"
	"vfinancy/backend/internal/shared/apperrors"
	"vfinancy/backend/internal/utils"
)

// UnitDTO is the serializable view of a unit of measure.
type UnitDTO struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Symbol         string `json:"symbol"`
	AllowsDecimals bool   `json:"allowsDecimals"`
}

func toUnitDTO(u *product.UnitOfMeasure) *UnitDTO {
	return &UnitDTO{
		ID:             u.ID.String(),
		Code:           u.Code,
		Name:           u.Name,
		Symbol:         u.Symbol,
		AllowsDecimals: u.AllowsDecimals,
	}
}

// ListUnits returns the units of measure available for the active company.
func (a *App) ListUnits() ([]UnitDTO, error) {
	units, err := a.productsSvc.ListUnits(a.Context(), a.companyID())
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	out := make([]UnitDTO, 0, len(units))
	for _, u := range units {
		out = append(out, *toUnitDTO(u))
	}
	return out, nil
}

// CategoryDTO is the serializable view of a product category.
type CategoryDTO struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// BrandDTO is the serializable view of a product brand.
type BrandDTO struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// ListCategories returns the product categories available for the active company.
func (a *App) ListCategories() ([]CategoryDTO, error) {
	categories, err := a.productsSvc.ListCategories(a.Context(), a.companyID())
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	out := make([]CategoryDTO, 0, len(categories))
	for _, c := range categories {
		out = append(out, CategoryDTO{ID: c.ID.String(), Code: c.Code.String(), Name: c.Name.String()})
	}
	return out, nil
}

// ListBrands returns the product brands available for the active company.
func (a *App) ListBrands() ([]BrandDTO, error) {
	brands, err := a.productsSvc.ListBrands(a.Context(), a.companyID())
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	out := make([]BrandDTO, 0, len(brands))
	for _, b := range brands {
		out = append(out, BrandDTO{ID: b.ID.String(), Code: b.Code.String(), Name: b.Name.String()})
	}
	return out, nil
}

// --- Category CRUD ---

// CategoryCodeNameRequest carries the mutable category fields.
type CategoryCodeNameRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CreateCategoryRequest creates a category.
type CreateCategoryRequest = CategoryCodeNameRequest

// UpdateCategoryRequest updates a category.
type UpdateCategoryRequest struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func parseCategoryFields(code, name string) (valueobjects.ShortCode, valueobjects.FullName, error) {
	c, err := valueobjects.NewShortCode(code)
	if err != nil {
		return valueobjects.ShortCode(""), valueobjects.FullName{}, apperrors.Errorf(apperrors.ErrValidation, utils.ProcessError(err).Error())
	}
	n, err := valueobjects.NewFullName(name)
	if err != nil {
		return valueobjects.ShortCode(""), valueobjects.FullName{}, apperrors.Errorf(apperrors.ErrValidation, utils.ProcessError(err).Error())
	}
	return c, n, nil
}

func mapCatalogError(err error) error {
	if errors.Is(err, repositories.ErrDuplicate) {
		return utils.ProcessError(apperrors.Errorf(apperrors.ErrConflict, "ya existe un registro con ese código"))
	}
	return utils.ProcessError(err)
}

// CreateCategory persists a new product category.
func (a *App) CreateCategory(req CreateCategoryRequest) (*CategoryDTO, error) {
	code, name, err := parseCategoryFields(req.Code, req.Name)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	c, err := a.productsSvc.CreateCategory(a.Context(), product.CategoryInput{
		CompanyID: a.companyID(),
		Code:      code,
		Name:      name,
	})
	if err != nil {
		return nil, utils.ProcessError(mapCatalogError(err))
	}
	return &CategoryDTO{ID: c.ID.String(), Code: c.Code.String(), Name: c.Name.String()}, nil
}

// UpdateCategory updates a product category.
func (a *App) UpdateCategory(req UpdateCategoryRequest) (*CategoryDTO, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, utils.ProcessError(apperrors.Errorf(apperrors.ErrValidation, "id inválido"))
	}
	code, name, err := parseCategoryFields(req.Code, req.Name)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	c, err := a.productsSvc.UpdateCategory(a.Context(), product.CategoryUpdateInput{
		ID:   id,
		Code: code,
		Name: name,
	})
	if err != nil {
		return nil, utils.ProcessError(mapCatalogError(err))
	}
	return &CategoryDTO{ID: c.ID.String(), Code: c.Code.String(), Name: c.Name.String()}, nil
}

// DeleteCategory removes a product category (soft delete).
func (a *App) DeleteCategory(id string) error {
	pid, err := uuid.Parse(id)
	if err != nil {
		return utils.ProcessError(apperrors.Errorf(apperrors.ErrValidation, "id inválido"))
	}
	return utils.ProcessError(mapCatalogError(a.productsSvc.DeleteCategory(a.Context(), pid)))
}

// --- Brand CRUD ---

// CreateBrandRequest creates a brand.
type CreateBrandRequest = CategoryCodeNameRequest

// UpdateBrandRequest updates a brand.
type UpdateBrandRequest = UpdateCategoryRequest

// CreateBrand persists a new product brand.
func (a *App) CreateBrand(req CreateBrandRequest) (*BrandDTO, error) {
	code, name, err := parseCategoryFields(req.Code, req.Name)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	b, err := a.productsSvc.CreateBrand(a.Context(), product.BrandInput{
		CompanyID: a.companyID(),
		Code:      code,
		Name:      name,
	})
	if err != nil {
		return nil, utils.ProcessError(mapCatalogError(err))
	}
	return &BrandDTO{ID: b.ID.String(), Code: b.Code.String(), Name: b.Name.String()}, nil
}

// UpdateBrand updates a product brand.
func (a *App) UpdateBrand(req UpdateBrandRequest) (*BrandDTO, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, utils.ProcessError(apperrors.Errorf(apperrors.ErrValidation, "id inválido"))
	}
	code, name, err := parseCategoryFields(req.Code, req.Name)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	b, err := a.productsSvc.UpdateBrand(a.Context(), product.BrandInput{
		ID:   id,
		Code: code,
		Name: name,
	})
	if err != nil {
		return nil, utils.ProcessError(mapCatalogError(err))
	}
	return &BrandDTO{ID: b.ID.String(), Code: b.Code.String(), Name: b.Name.String()}, nil
}

// DeleteBrand removes a product brand (soft delete).
func (a *App) DeleteBrand(id string) error {
	pid, err := uuid.Parse(id)
	if err != nil {
		return utils.ProcessError(apperrors.Errorf(apperrors.ErrValidation, "id inválido"))
	}
	return utils.ProcessError(mapCatalogError(a.productsSvc.DeleteBrand(a.Context(), pid)))
}
