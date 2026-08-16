package bindings

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/purchasing"
)

// PurchaseItemDTO is a serializable purchase order line.
type PurchaseItemDTO struct {
	ID              string `json:"id"`
	ProductID       string `json:"productId"`
	LineNumber      int    `json:"lineNumber"`
	Quantity        string `json:"quantity"`
	UnitPrice       string `json:"unitPrice"`
	DiscountPercent string `json:"discountPercent"`
	DiscountAmount  string `json:"discountAmount"`
	TaxRate         string `json:"taxRate"`
	TaxAmount       string `json:"taxAmount"`
	Description     string `json:"description"`
}

// PurchaseOrderDTO is the serializable view of a purchase order.
type PurchaseOrderDTO struct {
	ID           string             `json:"id"`
	Number       string             `json:"number"`
	SupplierID   string             `json:"supplierId"`
	OrderDate    string             `json:"orderDate"`
	Status       string             `json:"status"`
	Subtotal     string             `json:"subtotal"`
	Discount     string             `json:"discount"`
	Tax          string             `json:"tax"`
	Total        string             `json:"total"`
	Paid         string             `json:"paid"`
	Notes        string             `json:"notes"`
	Items        []*PurchaseItemDTO `json:"items"`
}

func toPurchaseOrderDTO(po *purchasing.PurchaseOrder) *PurchaseOrderDTO {
	items := make([]*PurchaseItemDTO, 0, len(po.Items))
	for _, it := range po.Items {
		items = append(items, &PurchaseItemDTO{
			ID:              it.ID.String(),
			ProductID:       it.ProductID.String(),
			LineNumber:      it.LineNumber,
			Quantity:        it.Quantity.String(),
			UnitPrice:       it.UnitPrice.String(),
			DiscountPercent: it.DiscountPercent.String(),
			DiscountAmount:  it.DiscountAmount.String(),
			TaxRate:         it.TaxRate.String(),
			TaxAmount:       it.TaxAmount.String(),
			Description:     it.Description,
		})
	}
	return &PurchaseOrderDTO{
		ID:         po.ID.String(),
		Number:     po.Number,
		SupplierID: po.SupplierID.String(),
		OrderDate:  po.OrderDate.Format("2006-01-02"),
		Status:     po.Status.String(),
		Subtotal:   po.Subtotal.String(),
		Discount:   po.DiscountAmount.String(),
		Tax:        po.TaxAmount.String(),
		Total:      po.Total.String(),
		Paid:       po.Paid.String(),
		Notes:      po.Notes,
		Items:      items,
	}
}

// ListPurchaseOrdersRequest filters the purchase order listing.
type ListPurchaseOrdersRequest struct {
	Status string `json:"status"`
	PaginationRequest
}

// ListPurchaseOrders returns paged purchase orders.
func (a *App) ListPurchaseOrders(req ListPurchaseOrdersRequest) (PageResult, error) {
	ctx := a.Context()
	filter := purchasing.PurchaseFilter{
		CompanyID:   &demoCompanyID,
		Status:      req.Status,
		PageRequest: req.toPageRequest(),
	}
	page, err := a.purchasingSvc.List(ctx, filter)
	if err != nil {
		return PageResult{}, err
	}
	items := make([]*PurchaseOrderDTO, 0, len(page.Items))
	for _, po := range page.Items {
		items = append(items, toPurchaseOrderDTO(po))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// GetPurchaseOrder returns a single purchase order with its lines.
func (a *App) GetPurchaseOrder(id string) (*PurchaseOrderDTO, error) {
	pid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	po, err := a.purchasingSvc.GetByID(a.Context(), pid)
	if err != nil {
		return nil, err
	}
	return toPurchaseOrderDTO(po), nil
}

// CreatePurchaseOrderItemRequest is one line of a new purchase order.
type CreatePurchaseOrderItemRequest struct {
	ProductID       string `json:"productId"`
	Quantity        string `json:"quantity"`
	UnitPrice       string `json:"unitPrice"`
	DiscountPercent string `json:"discountPercent"`
	DiscountAmount  string `json:"discountAmount"`
	TaxRate         string `json:"taxRate"`
	TaxAmount       string `json:"taxAmount"`
	Description     string `json:"description"`
}

// CreatePurchaseOrderRequest creates a purchase order.
type CreatePurchaseOrderRequest struct {
	SupplierID   string                           `json:"supplierId"`
	CurrencyCode string                           `json:"currencyCode"`
	ExchangeRate string                           `json:"exchangeRate"`
	OrderDate    string                           `json:"orderDate"`
	Notes        string                           `json:"notes"`
	Items        []CreatePurchaseOrderItemRequest `json:"items"`
}

// CreatePurchaseOrder persists a purchase order.
func (a *App) CreatePurchaseOrder(req CreatePurchaseOrderRequest) (*PurchaseOrderDTO, error) {
	ctx := a.Context()
	supplierID, err := uuid.Parse(req.SupplierID)
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
	orderDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		return nil, err
	}
	items := make([]purchasing.CreateItemInput, 0, len(req.Items))
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
		items = append(items, purchasing.CreateItemInput{
			ProductID:       productID,
			Quantity:        qty,
			UnitPrice:       unitPrice,
			DiscountPercent: discountPct,
			DiscountAmount:  discountAmount,
			TaxRate:         taxRate,
			TaxAmount:       taxAmount,
			Description:     it.Description,
		})
	}
	in := purchasing.CreateInput{
		CompanyID:    demoCompanyID,
		Number:       "",
		SupplierID:   supplierID,
		CurrencyCode: cc,
		ExchangeRate: rate,
		OrderDate:    valueobjects.Date(orderDate),
		Notes:        req.Notes,
		Items:        items,
	}
	po, err := a.purchasingSvc.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return toPurchaseOrderDTO(po), nil
}

// RegisterPurchasePaymentRequest registers a full payment for a purchase.
type RegisterPurchasePaymentRequest struct {
	ID          string `json:"id"`
	PaymentDate string `json:"paymentDate"`
	Method      string `json:"method"`
	Reference   string `json:"reference"`
	Notes       string `json:"notes"`
}

// RegisterPurchasePayment records a supplier payment covering the
// purchase's full balance and marks the purchase as paid.
func (a *App) RegisterPurchasePayment(req RegisterPurchasePaymentRequest) (*PurchaseOrderDTO, error) {
	ctx := a.Context()
	pid, err := uuid.Parse(req.ID)
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
	po, err := a.purchasingSvc.MarkPaid(ctx, pid, purchasing.MarkPaidInput{
		CompanyID:   demoCompanyID,
		PaymentDate: pd,
		Method:      method,
		Reference:   req.Reference,
		Notes:       req.Notes,
	})
	if err != nil {
		return nil, err
	}
	return toPurchaseOrderDTO(po), nil
}

// CancelPurchaseOrderRequest cancels an existing purchase order.
type CancelPurchaseOrderRequest struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// CancelPurchaseOrder cancels an existing purchase order.
func (a *App) CancelPurchaseOrder(req CancelPurchaseOrderRequest) error {
	pid, err := uuid.Parse(req.ID)
	if err != nil {
		return err
	}
	return a.purchasingSvc.Cancel(a.Context(), pid, req.Reason)
}
