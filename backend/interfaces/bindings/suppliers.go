package bindings

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/features/supplier"
	"vfinancy/backend/internal/utils"
)

// SupplierDTO is the serializable view of a supplier.
type SupplierDTO struct {
	ID              string `json:"id"`
	DocumentType    string `json:"documentType"`
	DocumentNumber  string `json:"documentNumber"`
	BusinessName    string `json:"businessName"`
	TradeName       string `json:"tradeName"`
	TaxID           string `json:"taxId"`
	IsInternational bool   `json:"isInternational"`
	DefaultCurrency string `json:"defaultCurrency"`
	CurrentDebt     string `json:"currentDebt"`
	PaymentTermDays int    `json:"paymentTermDays"`
	Status          string `json:"status"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	CreatedAt       string `json:"createdAt"`
}

func toSupplierDTO(s *supplier.Supplier) *SupplierDTO {
	return &SupplierDTO{
		ID:              s.ID.String(),
		DocumentType:    s.Document.Type().String(),
		DocumentNumber:  s.Document.Number(),
		BusinessName:    s.BusinessName.String(),
		TradeName:       s.TradeName.String(),
		TaxID:           s.TaxID,
		IsInternational: s.IsInternational,
		DefaultCurrency: s.DefaultCurrency.String(),
		CurrentDebt:     s.CurrentDebt.String(),
		PaymentTermDays: s.PaymentTermDays,
		Status:          s.Status.String(),
		Email:           s.Email.String(),
		Phone:           s.Phone.String(),
		Address:         s.Address.String(),
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
	}
}

// ListSuppliersRequest filters the supplier listing.
type ListSuppliersRequest struct {
	Search string `json:"search"`
	Status string `json:"status"`
	PaginationRequest
}

// ListSuppliers returns paged suppliers.
func (a *App) ListSuppliers(req ListSuppliersRequest) (PageResult, error) {
	filter := supplier.SupplierFilter{
		CompanyID:   a.companyIDPtr(),
		Search:      req.Search,
		Status:      req.Status,
		PageRequest: req.toPageRequest(),
	}
	page, err := a.suppliersSvc.List(a.Context(), filter)
	if err != nil {
		return PageResult{}, utils.ProcessError(err)
	}
	items := make([]*SupplierDTO, 0, len(page.Items))
	for _, s := range page.Items {
		items = append(items, toSupplierDTO(s))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// GetSupplier returns a single supplier by ID.
func (a *App) GetSupplier(id string) (*SupplierDTO, error) {
	sid, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	s, err := a.suppliersSvc.GetByID(a.Context(), sid)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toSupplierDTO(s), nil
}

// CreateSupplierRequest creates a supplier.
type CreateSupplierRequest struct {
	DocumentType    string `json:"documentType"`
	DocumentNumber  string `json:"documentNumber"`
	BusinessName    string `json:"businessName"`
	TradeName       string `json:"tradeName"`
	TaxID           string `json:"taxId"`
	IsInternational bool   `json:"isInternational"`
	DefaultCurrency string `json:"defaultCurrency"`
	PaymentTermDays int    `json:"paymentTermDays"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
}

// CreateSupplier persists a new supplier.
func (a *App) CreateSupplier(req CreateSupplierRequest) (*SupplierDTO, error) {
	currency, err := currencyOrDefault(req.DefaultCurrency)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	in := supplier.CreateInput{
		CompanyID:       a.companyID(),
		DocumentType:    enums.DocumentType(req.DocumentType),
		DocumentNumber:  req.DocumentNumber,
		BusinessName:    req.BusinessName,
		TradeName:       req.TradeName,
		TaxID:           req.TaxID,
		IsInternational: req.IsInternational,
		DefaultCurrency: currency,
		PaymentTermDays: req.PaymentTermDays,
		Email:           req.Email,
		Phone:           req.Phone,
		Address:         req.Address,
	}
	s, err := a.suppliersSvc.Create(a.Context(), in)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toSupplierDTO(s), nil
}

// UpdateSupplierRequest updates a supplier.
type UpdateSupplierRequest struct {
	ID              string `json:"id"`
	BusinessName    string `json:"businessName"`
	TradeName       string `json:"tradeName"`
	TaxID           string `json:"taxId"`
	PaymentTermDays int    `json:"paymentTermDays"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	IsActive        *bool  `json:"isActive"`
}

// UpdateSupplier updates the mutable supplier fields.
func (a *App) UpdateSupplier(req UpdateSupplierRequest) (*SupplierDTO, error) {
	sid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	in := supplier.UpdateInput{
		ID:           sid,
		BusinessName: req.BusinessName,
		TradeName:    req.TradeName,
		TaxID:        req.TaxID,
		Email:        req.Email,
		Phone:        req.Phone,
		Address:      req.Address,
		IsActive:     req.IsActive,
	}
	days := req.PaymentTermDays
	in.PaymentTermDays = &days
	s, err := a.suppliersSvc.Update(a.Context(), in)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toSupplierDTO(s), nil
}

// RemoveSupplier hard-deletes a supplier. Suppliers referenced by
// purchases, payments or returns are rejected with a friendly error.
func (a *App) RemoveSupplier(id string) error {
	sid, err := uuid.Parse(id)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.suppliersSvc.Delete(a.Context(), sid))
}
