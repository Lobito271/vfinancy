package bindings

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/customer"
	"vfinancy/backend/internal/utils"
)

// CustomerDTO is the serializable view of a customer.
type CustomerDTO struct {
	ID              string `json:"id"`
	DocumentType    string `json:"documentType"`
	DocumentNumber  string `json:"documentNumber"`
	BusinessName    string `json:"businessName"`
	TradeName       string `json:"tradeName"`
	CreditLimit     string `json:"creditLimit"`
	CurrentDebt     string `json:"currentDebt"`
	PaymentTermDays int    `json:"paymentTermDays"`
	Status          string `json:"status"`
	BlockedReason   string `json:"blockedReason"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	CreatedAt       string `json:"createdAt"`
}

func toCustomerDTO(c *customer.Customer) *CustomerDTO {
	return &CustomerDTO{
		ID:              c.ID.String(),
		DocumentType:    c.Document.Type().String(),
		DocumentNumber:  c.Document.Number(),
		BusinessName:    c.BusinessName.String(),
		TradeName:       c.TradeName.String(),
		CreditLimit:     c.CreditLimit.String(),
		CurrentDebt:     c.CurrentDebt.String(),
		PaymentTermDays: c.PaymentTermDays,
		Status:          c.Status.String(),
		BlockedReason:   c.BlockedReason,
		Email:           c.Email.String(),
		Phone:           c.Phone.String(),
		Address:         c.Address.String(),
		CreatedAt:       c.CreatedAt.Format(time.RFC3339),
	}
}

// ListCustomersRequest filters the customer listing.
type ListCustomersRequest struct {
	Search string `json:"search"`
	Status string `json:"status"`
	PaginationRequest
}

// ListCustomers returns paged customers.
func (a *App) ListCustomers(req ListCustomersRequest) (PageResult, error) {
	ctx := a.Context()
	filter := customer.CustomerFilter{
		CompanyID:   a.companyIDPtr(),
		Search:      req.Search,
		Status:      req.Status,
		PageRequest: req.toPageRequest(),
	}
	page, err := a.customersSvc.List(ctx, filter)
	if err != nil {
		return PageResult{}, utils.ProcessError(err)
	}
	items := make([]*CustomerDTO, 0, len(page.Items))
	for _, c := range page.Items {
		items = append(items, toCustomerDTO(c))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// GetCustomer returns a single customer by ID.
func (a *App) GetCustomer(id string) (*CustomerDTO, error) {
	cid, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	c, err := a.customersSvc.GetByID(a.Context(), cid)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toCustomerDTO(c), nil
}

// CreateCustomerRequest creates a customer.
type CreateCustomerRequest struct {
	DocumentType    string `json:"documentType"`
	DocumentNumber  string `json:"documentNumber"`
	BusinessName    string `json:"businessName"`
	TradeName       string `json:"tradeName"`
	CreditLimit     string `json:"creditLimit"`
	PaymentTermDays int    `json:"paymentTermDays"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
}

// CreateCustomer persists a new customer.
func (a *App) CreateCustomer(req CreateCustomerRequest) (*CustomerDTO, error) {
	limit, err := valueobjects.MoneyFromString(req.CreditLimit)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	in := customer.CreateInput{
		CompanyID:       a.companyID(),
		DocumentType:    enums.DocumentType(req.DocumentType),
		DocumentNumber:  req.DocumentNumber,
		BusinessName:    req.BusinessName,
		TradeName:       req.TradeName,
		TaxCategory:     enums.TaxCategoryTaxed,
		CreditLimit:     limit,
		PaymentTermDays: req.PaymentTermDays,
		Email:           req.Email,
		Phone:           req.Phone,
		Address:         req.Address,
	}
	c, err := a.customersSvc.Create(a.Context(), in)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toCustomerDTO(c), nil
}

// UpdateCustomerRequest updates a customer.
type UpdateCustomerRequest struct {
	ID              string `json:"id"`
	BusinessName    string `json:"businessName"`
	TradeName       string `json:"tradeName"`
	CreditLimit     string `json:"creditLimit"`
	PaymentTermDays int    `json:"paymentTermDays"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	IsActive        *bool  `json:"isActive"`
}

// UpdateCustomer updates the mutable customer fields.
func (a *App) UpdateCustomer(req UpdateCustomerRequest) (*CustomerDTO, error) {
	cid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	limit, err := valueobjects.MoneyFromString(req.CreditLimit)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	days := req.PaymentTermDays
	in := customer.UpdateInput{
		ID:              cid,
		BusinessName:    req.BusinessName,
		TradeName:       req.TradeName,
		CreditLimit:     &limit,
		PaymentTermDays: &days,
		Email:           req.Email,
		Phone:           req.Phone,
		Address:         req.Address,
		IsActive:        req.IsActive,
	}
	c, err := a.customersSvc.Update(a.Context(), in)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toCustomerDTO(c), nil
}

// RemoveCustomer hard-deletes a customer. Customers referenced by
// sales, payments or advances are rejected with a friendly error.
func (a *App) RemoveCustomer(id string) error {
	cid, err := uuid.Parse(id)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.customersSvc.Delete(a.Context(), cid))
}

type BlockCustomerRequest struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

func (a *App) BlockCustomer(req BlockCustomerRequest) (*CustomerDTO, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	if err := a.customersSvc.Block(a.Context(), id, req.Reason); err != nil {
		return nil, utils.ProcessError(err)
	}
	c, err := a.customersSvc.GetByID(a.Context(), id)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toCustomerDTO(c), nil
}

func (a *App) UnblockCustomer(id string) (*CustomerDTO, error) {
	cid, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	if err := a.customersSvc.Unblock(a.Context(), cid); err != nil {
		return nil, utils.ProcessError(err)
	}
	c, err := a.customersSvc.GetByID(a.Context(), cid)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	return toCustomerDTO(c), nil
}
