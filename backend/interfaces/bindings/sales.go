package bindings

import (
	"time"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/features/sales"
)

type SaleItemDTO struct {
	ID             string  `json:"id"`
	ProductID      string  `json:"productId"`
	Description    string  `json:"description"`
	Quantity       float64 `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	LineTotal      float64 `json:"lineTotal"`
	CostSnapshot   float64 `json:"costSnapshot"`
}

type SaleDTO struct {
	ID              string        `json:"id"`
	CustomerID      string        `json:"customerId"`
	Number          string        `json:"number"`
	SaleDate        string        `json:"saleDate"`
	DueDate         string        `json:"dueDate"`
	Status          string        `json:"status"`
	SaleType        string        `json:"saleType"`
	Total           float64       `json:"total"`
	PaidAmount      float64       `json:"paidAmount"`
	CostTotal       float64       `json:"costTotal"`
	Profit          float64       `json:"profit"`
	Notes           string        `json:"notes"`
	CancelledAt     string        `json:"cancelledAt"`
	CancelledReason string        `json:"cancelledReason"`
	Items           []SaleItemDTO `json:"items"`
}

func saleDTO(s *sales.Sale) SaleDTO {
	items := make([]SaleItemDTO, 0, len(s.Items))
	for _, it := range s.Items {
		items = append(items, SaleItemDTO{
			ID:           it.ID.String(),
			ProductID:    it.ProductID.String(),
			Description:  it.Description,
			Quantity:     quantityFloat(it.Quantity),
			UnitPrice:    moneyFloat(it.UnitPrice),
			LineTotal:    moneyFloat(it.LineTotal),
			CostSnapshot: moneyFloat(it.CostSnapshot),
		})
	}
	return SaleDTO{
		ID:              s.ID.String(),
		CustomerID:      s.CustomerID.String(),
		Number:          s.Number,
		SaleDate:        dayStr(s.SaleDate),
		DueDate:         dayStrPtr(s.DueDate),
		Status:          string(s.Status),
		SaleType:        string(s.SaleType),
		Total:           moneyFloat(s.Total),
		PaidAmount:      moneyFloat(s.PaidAmount),
		CostTotal:       moneyFloat(s.CostTotal),
		Profit:          moneyFloat(s.Profit),
		Notes:           s.Notes,
		CancelledAt:     dayStrPtr(s.CancelledAt),
		CancelledReason: s.CancelledReason,
		Items:           items,
	}
}

type SaleFilterRequest struct {
	PaginationRequest
	Search     string `json:"search"`
	Status     string `json:"status"`
	SaleType   string `json:"saleType"`
	CustomerID string `json:"customerId"`
	From       string `json:"from"`
	To         string `json:"to"`
	OnlyUnpaid bool   `json:"onlyUnpaid"`
}

// ListSales returns local PEN sales.
func (a *App) ListSales(req SaleFilterRequest) (PageResult, error) {
	customerID, err := parseOptionalUUID(req.CustomerID)
	if err != nil {
		return PageResult{}, err
	}
	from, err := parseDayPtr(req.From)
	if err != nil {
		return PageResult{}, err
	}
	to, err := parseDayPtr(req.To)
	if err != nil {
		return PageResult{}, err
	}
	page, err := a.salesSvc.List(a.Context(), sales.SaleFilter{
		Search:      req.Search,
		Status:      req.Status,
		SaleType:    req.SaleType,
		CustomerID:  customerID,
		From:        from,
		To:          to,
		OnlyUnpaid:  req.OnlyUnpaid,
		PageRequest: req.toPageRequest(),
	})
	if err != nil {
		return PageResult{}, err
	}
	items := make([]SaleDTO, 0, len(page.Items))
	for _, s := range page.Items {
		items = append(items, saleDTO(s))
	}
	return PageResult{Items: items, Total: page.Total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetSale returns one sale with items.
func (a *App) GetSale(id string) (SaleDTO, error) {
	sid, err := parseUUID(id)
	if err != nil {
		return SaleDTO{}, err
	}
	s, err := a.salesSvc.GetByID(a.Context(), sid)
	if err != nil {
		return SaleDTO{}, err
	}
	return saleDTO(s), nil
}

type SaleLineRequest struct {
	ProductID string  `json:"productId"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type CreateSaleRequest struct {
	CustomerID    string            `json:"customerId"`
	SaleType      string            `json:"saleType"`
	Date          string            `json:"date"`
	DueDate       string            `json:"dueDate"`
	PaymentMethod string            `json:"paymentMethod"`
	Notes         string            `json:"notes"`
	InitialPayment float64          `json:"initialPayment"`
	Items         []SaleLineRequest `json:"items"`
}

// CreateSale records a PEN sale. A "stock" sale without stock is
// blocked (edge 4.3); a "client_order" sale registers the import
// requirement without touching stock.
func (a *App) CreateSale(req CreateSaleRequest) (SaleDTO, error) {
	cid, err := parseUUID(req.CustomerID)
	if err != nil {
		return SaleDTO{}, err
	}
	date := time.Now().UTC()
	if req.Date != "" {
		t, err := parseDay(req.Date)
		if err != nil {
			return SaleDTO{}, err
		}
		date = t
	}
	due, err := parseDayPtr(req.DueDate)
	if err != nil {
		return SaleDTO{}, err
	}
	initial, err := moneyFromFloat(req.InitialPayment)
	if err != nil {
		return SaleDTO{}, err
	}
	items := make([]sales.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		pid, err := parseUUID(it.ProductID)
		if err != nil {
			return SaleDTO{}, err
		}
		qty, err := quantityFromFloat(it.Quantity)
		if err != nil {
			return SaleDTO{}, err
		}
		price, err := moneyFromFloat(it.UnitPrice)
		if err != nil {
			return SaleDTO{}, err
		}
		items = append(items, sales.ItemInput{ProductID: pid, Quantity: qty, UnitPrice: price})
	}
	result, err := a.salesSvc.Create(a.Context(), sales.CreateInput{
		CustomerID:    cid,
		SaleType:      enums.SaleType(req.SaleType),
		Date:          date,
		DueDate:       due,
		PaymentMethod: enums.PaymentMethod(req.PaymentMethod),
		Notes:         req.Notes,
		InitialPayment: initial,
		Items:         items,
	})
	if err != nil {
		return SaleDTO{}, err
	}
	return saleDTO(result.Sale), nil
}

type CancelSaleRequest struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// CancelSale voids the sale and returns the stock.
func (a *App) CancelSale(req CancelSaleRequest) (SaleDTO, error) {
	sid, err := parseUUID(req.ID)
	if err != nil {
		return SaleDTO{}, err
	}
	s, err := a.salesSvc.Cancel(a.Context(), sales.CancelInput{ID: sid, Reason: req.Reason})
	if err != nil {
		return SaleDTO{}, err
	}
	return saleDTO(s), nil
}

type SalePaymentRequest struct {
	SaleID        string  `json:"saleId"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"paymentMethod"`
	Reference     string  `json:"reference"`
	Date          string  `json:"date"`
}

// RegisterSalePayment records a collection ("cobro") against the sale
// and reduces the customer debt.
func (a *App) RegisterSalePayment(req SalePaymentRequest) (SaleDTO, error) {
	sid, err := parseUUID(req.SaleID)
	if err != nil {
		return SaleDTO{}, err
	}
	amount, err := moneyFromFloat(req.Amount)
	if err != nil {
		return SaleDTO{}, err
	}
	date := time.Now().UTC()
	if req.Date != "" {
		t, err := parseDay(req.Date)
		if err != nil {
			return SaleDTO{}, err
		}
		date = t
	}
	if _, err := a.salesSvc.ApplyPayment(a.Context(), sid, sales.PaymentInput{
		Amount:        amount,
		PaymentMethod: enums.PaymentMethod(req.PaymentMethod),
		Reference:     req.Reference,
		Date:          date,
	}); err != nil {
		return SaleDTO{}, err
	}
	s, err := a.salesSvc.GetByID(a.Context(), sid)
	if err != nil {
		return SaleDTO{}, err
	}
	return saleDTO(s), nil
}

type CustomerPaymentDTO struct {
	ID            string  `json:"id"`
	CustomerID    string  `json:"customerId"`
	Number        string  `json:"number"`
	PaymentDate   string  `json:"paymentDate"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"paymentMethod"`
	Reference     string  `json:"reference"`
	Status        string  `json:"status"`
}

// ListSalePayments returns the collections ledger (optionally scoped to
// one sale or customer).
func (a *App) ListSalePayments(req PaginationRequest, customerID, saleID string) (PageResult, error) {
	cid, err := parseOptionalUUID(customerID)
	if err != nil {
		return PageResult{}, err
	}
	sid, err := parseOptionalUUID(saleID)
	if err != nil {
		return PageResult{}, err
	}
	page, err := a.salesSvc.ListPayments(a.Context(), sales.CustomerPaymentFilter{
		CustomerID:  cid,
		SaleID:      sid,
		PageRequest: req.toPageRequest(),
	})
	if err != nil {
		return PageResult{}, err
	}
	items := make([]CustomerPaymentDTO, 0, len(page.Items))
	for _, p := range page.Items {
		items = append(items, CustomerPaymentDTO{
			ID:            p.ID.String(),
			CustomerID:    p.CustomerID.String(),
			Number:        p.Number,
			PaymentDate:   dayStr(p.PaymentDate),
			Amount:        moneyFloat(p.Amount),
			PaymentMethod: string(p.PaymentMethod),
			Reference:     p.Reference,
			Status:        p.Status,
		})
	}
	return PageResult{Items: items, Total: page.Total, Page: req.Page, PageSize: req.PageSize}, nil
}
