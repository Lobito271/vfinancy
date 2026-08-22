package bindings

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/customer"
	"vfinancy/backend/internal/features/customerpayments"
	"vfinancy/backend/internal/features/sales"
)

// SaleItemDTO is a serializable sale line.
type SaleItemDTO struct {
	ID              string `json:"id"`
	ProductID       string `json:"productId"`
	LineNumber      int    `json:"lineNumber"`
	Quantity        string `json:"quantity"`
	UnitPrice       string `json:"unitPrice"`
	DiscountPercent string `json:"discountPercent"`
	DiscountAmount  string `json:"discountAmount"`
	TaxRate         string `json:"taxRate"`
	TaxAmount       string `json:"taxAmount"`
	CostSnapshot    string `json:"costSnapshot"`
	Description     string `json:"description"`
}

// SaleDTO is the serializable view of a sale.
type SaleDTO struct {
	ID           string         `json:"id"`
	Number       string         `json:"number"`
	CustomerID   string         `json:"customerId"`
	CustomerName string         `json:"customerName"`
	Date         string         `json:"date"`
	Status       string         `json:"status"`
	Subtotal     string         `json:"subtotal"`
	Tax          string         `json:"tax"`
	Discount     string         `json:"discount"`
	Total        string         `json:"total"`
	Cost         string         `json:"cost"`
	Profit       string         `json:"profit"`
	Paid         string         `json:"paid"`
	Balance      string         `json:"balance"`
	Items        []*SaleItemDTO `json:"items"`
}

// CustomerPaymentDTO is the serializable view of a customer payment.
type CustomerPaymentDTO struct {
	ID           string `json:"id"`
	Number       string `json:"number"`
	CustomerID   string `json:"customerId"`
	PaymentDate  string `json:"paymentDate"`
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
	Method       string `json:"method"`
	Status       string `json:"status"`
	Reference    string `json:"reference"`
	Notes        string `json:"notes"`
}

// CustomerAdvanceDTO is the serializable view of a customer advance.
type CustomerAdvanceDTO struct {
	ID           string `json:"id"`
	Number       string `json:"number"`
	CustomerID   string `json:"customerId"`
	AdvanceDate  string `json:"advanceDate"`
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
	Method       string `json:"method"`
	Remaining    string `json:"remaining"`
}

func toSaleDTO(ctx context.Context, customers *customer.CustomerService, s *sales.Sale) *SaleDTO {
	customerName := ""
	if cust, err := customers.GetByID(ctx, s.CustomerID); err == nil {
		customerName = cust.BusinessName.String()
	}
	items := make([]*SaleItemDTO, 0, len(s.Items))
	for _, it := range s.Items {
		items = append(items, &SaleItemDTO{
			ID:              it.ID.String(),
			ProductID:       it.ProductID.String(),
			LineNumber:      it.LineNumber,
			Quantity:        it.Quantity.String(),
			UnitPrice:       it.UnitPrice.String(),
			DiscountPercent: it.DiscountPercent.String(),
			DiscountAmount:  it.DiscountAmount.String(),
			TaxRate:         it.TaxRate.String(),
			TaxAmount:       it.TaxAmount.String(),
			CostSnapshot:    it.CostSnapshot.String(),
			Description:     it.Description,
		})
	}
	return &SaleDTO{
		ID:           s.ID.String(),
		Number:       s.Number,
		CustomerID:   s.CustomerID.String(),
		CustomerName: customerName,
		Date:         s.CreatedAt.Format(time.RFC3339),
		Status:       s.Status.String(),
		Subtotal:     s.Subtotal.String(),
		Tax:          s.TaxAmount.String(),
		Discount:     s.DiscountAmount.String(),
		Total:        s.Total.String(),
		Cost:         s.CostTotal.String(),
		Profit:       s.Profit.String(),
		Paid:         s.Paid.String(),
		Balance:      s.Balance().String(),
		Items:        items,
	}
}

// ListSalesRequest filters the sale listing.
type ListSalesRequest struct {
	CustomerID string `json:"customerId"`
	Status     string `json:"status"`
	PaginationRequest
}

// ListSales returns paged sales.
func (a *App) ListSales(req ListSalesRequest) (PageResult, error) {
	ctx := a.Context()
	filter := sales.SaleFilter{
		CompanyID:   a.companyIDPtr(),
		Status:      req.Status,
		PageRequest: req.toPageRequest(),
	}
	customerID, err := parseOptionalUUID(req.CustomerID)
	if err != nil {
		return PageResult{}, err
	}
	filter.CustomerID = customerID
	page, err := a.salesSvc.List(ctx, filter)
	if err != nil {
		return PageResult{}, err
	}
	items := make([]*SaleDTO, 0, len(page.Items))
	for _, s := range page.Items {
		items = append(items, toSaleDTO(ctx, a.customersSvc, s))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// GetSale returns a single sale with its lines.
func (a *App) GetSale(id string) (*SaleDTO, error) {
	ctx := a.Context()
	sid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	s, err := a.salesSvc.GetByID(ctx, sid)
	if err != nil {
		return nil, err
	}
	return toSaleDTO(ctx, a.customersSvc, s), nil
}

// CreateSaleItemRequest is one line of a new sale.
type CreateSaleItemRequest struct {
	ProductID       string `json:"productId"`
	Quantity        string `json:"quantity"`
	UnitPrice       string `json:"unitPrice"`
	DiscountPercent string `json:"discountPercent"`
	DiscountAmount  string `json:"discountAmount"`
	TaxRate         string `json:"taxRate"`
	TaxAmount       string `json:"taxAmount"`
	CostSnapshot    string `json:"costSnapshot"`
	Description     string `json:"description"`
}

// CreateSaleRequest creates a sale.
type CreateSaleRequest struct {
	CustomerID   string                  `json:"customerId"`
	CurrencyCode string                  `json:"currencyCode"`
	ExchangeRate string                  `json:"exchangeRate"`
	DueDate      string                  `json:"dueDate"`
	Notes        string                  `json:"notes"`
	Items        []CreateSaleItemRequest `json:"items"`
}

// CreateSale persists a sale and records the customer's debt.
func (a *App) CreateSale(req CreateSaleRequest) (*SaleDTO, error) {
	ctx := a.Context()
	cid, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return nil, err
	}
	cc, err := valueobjects.NewCurrencyCode(req.CurrencyCode)
	if err != nil {
		return nil, err
	}
	rate, err := valueobjects.ExchangeRateFromString(req.ExchangeRate)
	if err != nil {
		return nil, err
	}
	var dueDate *valueobjects.Date
	if req.DueDate != "" {
		d, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, err
		}
		dd := valueobjects.Date(d)
		dueDate = &dd
	}
	items := make([]sales.CreateItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		productID, err := uuid.Parse(it.ProductID)
		if err != nil {
			return nil, err
		}
		qty, err := valueobjects.QuantityFromString(it.Quantity)
		if err != nil {
			return nil, err
		}
		unitPrice, err := valueobjects.MoneyFromString(it.UnitPrice)
		if err != nil {
			return nil, err
		}
		discountPct, err := valueobjects.PercentageFromString(it.DiscountPercent)
		if err != nil {
			return nil, err
		}
		discountAmount, err := valueobjects.MoneyFromString(it.DiscountAmount)
		if err != nil {
			return nil, err
		}
		taxRate, err := valueobjects.PercentageFromString(it.TaxRate)
		if err != nil {
			return nil, err
		}
		taxAmount, err := valueobjects.MoneyFromString(it.TaxAmount)
		if err != nil {
			return nil, err
		}
		cost, err := valueobjects.MoneyFromString(it.CostSnapshot)
		if err != nil {
			return nil, err
		}
		items = append(items, sales.CreateItemInput{
			ProductID:       productID,
			Quantity:        qty,
			UnitPrice:       unitPrice,
			DiscountPercent: discountPct,
			DiscountAmount:  discountAmount,
			TaxRate:         taxRate,
			TaxAmount:       taxAmount,
			CostSnapshot:    cost,
			Description:     it.Description,
		})
	}
	in := sales.CreateInput{
		CompanyID:    a.companyID(),
		Number:       "",
		CustomerID:   cid,
		CurrencyCode: cc,
		ExchangeRate: rate,
		DueDate:      dueDate,
		Notes:        req.Notes,
		Items:        items,
	}
	res, err := a.salesSvc.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return toSaleDTO(ctx, a.customersSvc, res.Sale), nil
}

// CancelSaleRequest cancels an existing sale.
type CancelSaleRequest struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// CancelSale cancels an existing sale.
func (a *App) CancelSale(req CancelSaleRequest) (*SaleDTO, error) {
	ctx := a.Context()
	sid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	s, err := a.salesSvc.Cancel(ctx, sales.CancelInput{ID: sid, Reason: req.Reason})
	if err != nil {
		return nil, err
	}
	return toSaleDTO(ctx, a.customersSvc, s), nil
}

// RegisterSalePaymentRequest registers a full payment for a sale.
type RegisterSalePaymentRequest struct {
	ID          string `json:"id"`
	PaymentDate string `json:"paymentDate"`
	Method      string `json:"method"`
	Reference   string `json:"reference"`
	Notes       string `json:"notes"`
}

// RegisterSalePayment records a customer payment covering the sale's
// full balance, applies it to the sale and reduces the customer's
// debt. The sale status becomes "paid".
func (a *App) RegisterSalePayment(req RegisterSalePaymentRequest) (*SaleDTO, error) {
	ctx := a.Context()
	sid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	var pd valueobjects.Date
	if req.PaymentDate != "" {
		t, err := time.Parse("2006-01-02", req.PaymentDate)
		if err != nil {
			return nil, err
		}
		pd = valueobjects.Date(t)
	} else {
		pd = valueobjects.Date(time.Now().UTC())
	}
	method := enums.PaymentMethod(req.Method)
	if !method.Valid() {
		method = enums.PaymentMethodCash
	}
	s, err := a.paymentSvc.MarkPaid(ctx, sid, customerpayments.MarkPaidInput{
		CompanyID:   a.companyID(),
		PaymentDate: pd,
		Method:      method,
		Reference:   req.Reference,
		Notes:       req.Notes,
	})
	if err != nil {
		return nil, err
	}
	return toSaleDTO(ctx, a.customersSvc, s), nil
}

// ListCustomerPaymentsRequest lists payments for a customer.
type ListCustomerPaymentsRequest struct {
	CustomerID string `json:"customerId"`
	PaginationRequest
}

// ListCustomerPayments returns paged customer payments.
func (a *App) ListCustomerPayments(req ListCustomerPaymentsRequest) (PageResult, error) {
	ctx := a.Context()
	cid, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return PageResult{}, err
	}
	filter := sales.CustomerPaymentFilter{
		CompanyID:   a.companyIDPtr(),
		CustomerID:  &cid,
		PageRequest: req.toPageRequest(),
	}
	page, err := a.paymentSvc.ListPayments(ctx, filter)
	if err != nil {
		return PageResult{}, err
	}
	items := make([]*CustomerPaymentDTO, 0, len(page.Items))
	for _, p := range page.Items {
		items = append(items, &CustomerPaymentDTO{
			ID:           p.ID.String(),
			Number:       p.Number,
			CustomerID:   p.CustomerID.String(),
			PaymentDate:  p.PaymentDate.Format("2006-01-02"),
			Amount:       p.Amount.String(),
			CurrencyCode: p.CurrencyCode.String(),
			Method:       p.Method.String(),
			Status:       p.Status,
			Reference:    p.Reference,
			Notes:        p.Notes,
		})
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// ListCustomerAdvances returns the advances of a customer.
func (a *App) ListCustomerAdvances(customerID string) ([]*CustomerAdvanceDTO, error) {
	ctx := a.Context()
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return nil, err
	}
	advances, err := a.paymentSvc.ListAdvances(ctx, cid)
	if err != nil {
		return nil, err
	}
	items := make([]*CustomerAdvanceDTO, 0, len(advances))
	for _, ad := range advances {
		items = append(items, &CustomerAdvanceDTO{
			ID:           ad.ID.String(),
			Number:       ad.Number,
			CustomerID:   ad.CustomerID.String(),
			AdvanceDate:  ad.AdvanceDate.Format("2006-01-02"),
			Amount:       ad.Amount.String(),
			CurrencyCode: ad.CurrencyCode.String(),
			Method:       ad.Method.String(),
			Remaining:    ad.Remaining().String(),
		})
	}
	return items, nil
}

type RegisterCustomerAdvanceRequest struct {
	CustomerID   string `json:"customerId"`
	AdvanceDate  string `json:"advanceDate"`
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
	ExchangeRate string `json:"exchangeRate"`
	Method       string `json:"method"`
	Notes        string `json:"notes"`
}

func (a *App) RegisterCustomerAdvance(req RegisterCustomerAdvanceRequest) (*CustomerAdvanceDTO, error) {
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return nil, err
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return nil, err
	}
	currency, err := valueobjects.NewCurrencyCode(req.CurrencyCode)
	if err != nil {
		return nil, err
	}
	rate, err := valueobjects.ExchangeRateFromString(req.ExchangeRate)
	if err != nil {
		return nil, err
	}
	date := time.Now().UTC()
	if req.AdvanceDate != "" {
		date, err = time.Parse("2006-01-02", req.AdvanceDate)
		if err != nil {
			return nil, err
		}
	}
	method := enums.PaymentMethod(req.Method)
	if !method.Valid() {
		method = enums.PaymentMethodCash
	}
	advance, err := a.paymentSvc.RegisterAdvance(a.Context(), customerpayments.AdvanceInput{
		CompanyID: a.companyID(), CustomerID: customerID, AdvanceDate: valueobjects.Date(date), Amount: amount,
		CurrencyCode: currency, ExchangeRate: rate, Method: method, Notes: req.Notes,
	})
	if err != nil {
		return nil, err
	}
	return &CustomerAdvanceDTO{ID: advance.ID.String(), Number: advance.Number, CustomerID: advance.CustomerID.String(), AdvanceDate: advance.AdvanceDate.Format("2006-01-02"), Amount: advance.Amount.String(), CurrencyCode: advance.CurrencyCode.String(), Method: advance.Method.String(), Remaining: advance.Remaining().String()}, nil
}

type ApplyCustomerAdvanceRequest struct {
	AdvanceID string `json:"advanceId"`
	SaleID    string `json:"saleId"`
	Amount    string `json:"amount"`
}

func (a *App) ApplyCustomerAdvance(req ApplyCustomerAdvanceRequest) (string, error) {
	advanceID, err := uuid.Parse(req.AdvanceID)
	if err != nil {
		return "", err
	}
	saleID, err := uuid.Parse(req.SaleID)
	if err != nil {
		return "", err
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return "", err
	}
	remaining, err := a.paymentSvc.ApplyAdvanceToSale(a.Context(), advanceID, saleID, amount)
	if err != nil {
		return "", err
	}
	return remaining.String(), nil
}

type RegisterPartialSalePaymentRequest struct {
	ID          string `json:"id"`
	Amount      string `json:"amount"`
	PaymentDate string `json:"paymentDate"`
	Method      string `json:"method"`
	Reference   string `json:"reference"`
	Notes       string `json:"notes"`
}

func (a *App) RegisterPartialSalePayment(req RegisterPartialSalePaymentRequest) (*SaleDTO, error) {
	saleID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
	if err != nil {
		return nil, err
	}
	date := time.Now().UTC()
	if req.PaymentDate != "" {
		date, err = time.Parse("2006-01-02", req.PaymentDate)
		if err != nil {
			return nil, err
		}
	}
	method := enums.PaymentMethod(req.Method)
	if !method.Valid() {
		method = enums.PaymentMethodCash
	}
	sale, err := a.paymentSvc.ApplyPayment(a.Context(), saleID, customerpayments.ApplyPaymentInput{
		CompanyID: a.companyID(), PaymentDate: valueobjects.Date(date), Amount: amount, Method: method,
		Reference: req.Reference, Notes: req.Notes,
	})
	if err != nil {
		return nil, err
	}
	return toSaleDTO(a.Context(), a.customersSvc, sale), nil
}
