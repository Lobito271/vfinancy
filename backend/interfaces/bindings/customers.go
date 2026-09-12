package bindings

import (
	"vfinancy/backend/internal/features/customer"
)

type CustomerDTO struct {
	ID             string  `json:"id"`
	DocumentType   string  `json:"documentType"`
	DocumentNumber string  `json:"documentNumber"`
	BusinessName   string  `json:"businessName"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	Address        string  `json:"address"`
	CurrentDebt    float64 `json:"currentDebt"`
	Status         string  `json:"status"`
}

func customerDTO(c *customer.Customer) CustomerDTO {
	return CustomerDTO{
		ID:             c.ID.String(),
		DocumentType:   string(c.DocumentType),
		DocumentNumber: c.DocumentNumber.String(),
		BusinessName:   c.BusinessName.String(),
		Email:          c.Email.String(),
		Phone:          c.Phone.String(),
		Address:        c.Address.String(),
		CurrentDebt:    moneyFloat(c.CurrentDebt),
		Status:         string(c.Status),
	}
}

// ListCustomers returns the customer portfolio.
func (a *App) ListCustomers(req PaginationRequest, search string) (PageResult, error) {
	page, err := a.customersSvc.List(a.Context(), customer.CustomerFilter{Search: search, PageRequest: req.toPageRequest()})
	if err != nil {
		return PageResult{}, err
	}
	items := make([]CustomerDTO, 0, len(page.Items))
	for _, c := range page.Items {
		items = append(items, customerDTO(c))
	}
	return PageResult{Items: items, Total: page.Total, Page: req.Page, PageSize: req.PageSize}, nil
}

// CustomerOptions returns active customers for the sale form select.
func (a *App) CustomerOptions() ([]CustomerDTO, error) {
	list, err := a.customersSvc.Options(a.Context())
	if err != nil {
		return nil, err
	}
	items := make([]CustomerDTO, 0, len(list))
	for _, c := range list {
		items = append(items, customerDTO(c))
	}
	return items, nil
}

type SaveCustomerRequest struct {
	ID             string `json:"id"`
	BusinessName   string `json:"businessName"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
}

// CreateCustomer registers a client with optional DNI/RUC validation
// (edge case 4.3: 8 digits / 11 digits with prefix 10 or 20).
func (a *App) CreateCustomer(req SaveCustomerRequest) (CustomerDTO, error) {
	c, err := a.customersSvc.Create(a.Context(), customer.CreateInput{
		BusinessName: req.BusinessName,
		DocType:      req.DocumentType,
		DocNumber:    req.DocumentNumber,
		Email:        req.Email,
		Phone:        req.Phone,
		Address:      req.Address,
	})
	if err != nil {
		return CustomerDTO{}, err
	}
	return customerDTO(c), nil
}

// UpdateCustomer edits the client data.
func (a *App) UpdateCustomer(req SaveCustomerRequest) (CustomerDTO, error) {
	id, err := parseUUID(req.ID)
	if err != nil {
		return CustomerDTO{}, err
	}
	c, err := a.customersSvc.Update(a.Context(), customer.UpdateInput{
		ID:          id,
		BusinessName: req.BusinessName,
		DocType:     req.DocumentType,
		DocNumber:   req.DocumentNumber,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
	})
	if err != nil {
		return CustomerDTO{}, err
	}
	return customerDTO(c), nil
}

// RemoveCustomer soft-deletes the client.
func (a *App) RemoveCustomer(id string) error {
	cid, err := parseUUID(id)
	if err != nil {
		return err
	}
	return a.customersSvc.Delete(a.Context(), cid)
}

// GetCustomerByDocument finds a client by its fiscal identity.
func (a *App) GetCustomerByDocument(docType, docNumber string) (CustomerDTO, error) {
	c, err := a.customersSvc.GetByDocument(a.Context(), docType, docNumber)
	if err != nil {
		return CustomerDTO{}, err
	}
	return customerDTO(c), nil
}
