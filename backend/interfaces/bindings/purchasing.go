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
	ID                  string                    `json:"id"`
	Number              string                    `json:"number"`
	SupplierID          string                    `json:"supplierId"`
	CustomerID          string                    `json:"customerId"`
	CreditCardID        string                    `json:"creditCardId"`
	OrderType           string                    `json:"orderType"`
	OrderDate           string                    `json:"orderDate"`
	Status              string                    `json:"status"`
	Subtotal            string                    `json:"subtotal"`
	Discount            string                    `json:"discount"`
	Tax                 string                    `json:"tax"`
	Total               string                    `json:"total"`
	Paid                string                    `json:"paid"`
	CostUSD             string                    `json:"costUSD"`
	SalePricePEN        string                    `json:"salePricePEN"`
	RealCostPEN         string                    `json:"realCostPEN"`
	ProjectedProfitPEN  string                    `json:"projectedProfitPEN"`
	Anticipo            string                    `json:"anticipo"`
	AnticipoDate        string                    `json:"anticipoDate"`
	PorCobrar           string                    `json:"porCobrar"`
	Faulty              bool                      `json:"faulty"`
	FaultyReason        string                    `json:"faultyReason"`
	RefundedAmount      string                    `json:"refundedAmount"`
	ArrivalDate         string                    `json:"arrivalDate"`
	SupplierOrderNumber string                    `json:"supplierOrderNumber"`
	Notes               string                    `json:"notes"`
	Items               []*PurchaseItemDTO        `json:"items"`
	Payments            []*CustomerOrderPaymentDTO `json:"payments"`
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
	payments := make([]*CustomerOrderPaymentDTO, 0, len(po.CustomerPayments))
	for _, pm := range po.CustomerPayments {
		payments = append(payments, toCustomerOrderPaymentDTO(pm))
	}
	return &PurchaseOrderDTO{
		ID:                  po.ID.String(),
		Number:              po.Number,
		SupplierID:          po.SupplierID.String(),
		CustomerID:          persistenceStringFromUUID(po.CustomerID),
		CreditCardID:        persistenceStringFromUUID(po.CreditCardID),
		OrderType:           po.OrderType.String(),
		OrderDate:           po.OrderDate.Format("2006-01-02"),
		Status:              po.Status.String(),
		Subtotal:            po.Subtotal.String(),
		Discount:            po.DiscountAmount.String(),
		Tax:                 po.TaxAmount.String(),
		Total:               po.Total.String(),
		Paid:                po.Paid.String(),
		CostUSD:             po.CostUSD.String(),
		SalePricePEN:        po.SalePricePEN.String(),
		RealCostPEN:         po.RealCostPEN.String(),
		ProjectedProfitPEN:  po.ProjectedProfitPEN.String(),
		Anticipo:            po.Anticipo.String(),
		AnticipoDate:        dateString(po.AnticipoDate),
		PorCobrar:           po.PorCobrar().String(),
		Faulty:              po.Faulty,
		FaultyReason:        po.FaultyReason,
		RefundedAmount:      po.RefundedAmount.String(),
		ArrivalDate:         dateString(po.ArrivalDate),
		SupplierOrderNumber: po.SupplierOrderNumber,
		Notes:               po.Notes,
		Items:               items,
		Payments:            payments,
	}
}

func persistenceStringFromUUID(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func dateString(d *valueobjects.Date) string {
	if d == nil {
		return ""
	}
	return d.Format("2006-01-02")
}

// ListPurchaseOrdersRequest filters the purchase order listing.
type ListPurchaseOrdersRequest struct {
	Status    string `json:"status"`
	OrderType string `json:"orderType"`
	Search    string `json:"search"`
	PaginationRequest
}

// ListPurchaseOrders returns paged purchase orders.
func (a *App) ListPurchaseOrders(req ListPurchaseOrdersRequest) (PageResult, error) {
	ctx := a.Context()
	filter := purchasing.PurchaseFilter{
		CompanyID:   a.companyIDPtr(),
		Status:      req.Status,
		OrderType:   req.OrderType,
		Search:      req.Search,
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
	SupplierID          string                           `json:"supplierId"`
	CustomerID          string                           `json:"customerId"`
	CreditCardID        string                           `json:"creditCardId"`
	OrderType           string                           `json:"orderType"`
	CurrencyCode        string                           `json:"currencyCode"`
	ExchangeRate        string                           `json:"exchangeRate"`
	OrderDate           string                           `json:"orderDate"`
	SupplierOrderNumber string                           `json:"supplierOrderNumber"`
	CostUSD             string                           `json:"costUSD"`
	SalePricePEN        string                           `json:"salePricePEN"`
	Anticipo            string                           `json:"anticipo"`
	AnticipoDate        string                           `json:"anticipoDate"`
	Notes               string                           `json:"notes"`
	Items               []CreatePurchaseOrderItemRequest `json:"items"`
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
	orderType := enums.OrderType(req.OrderType)
	if orderType == "" {
		orderType = enums.OrderTypeGeneral
	}
	var customerID *uuid.UUID
	if req.CustomerID != "" {
		cid, err := uuid.Parse(req.CustomerID)
		if err != nil {
			return nil, err
		}
		customerID = &cid
	}
	var creditCardID *uuid.UUID
	if req.CreditCardID != "" {
		cid, err := uuid.Parse(req.CreditCardID)
		if err != nil {
			return nil, err
		}
		creditCardID = &cid
	}
	costUSD, err := valueobjects.MoneyFromString(req.CostUSD)
	if err != nil {
		return nil, err
	}
	salePricePEN, err := valueobjects.MoneyFromString(req.SalePricePEN)
	if err != nil {
		return nil, err
	}
	anticipo, err := valueobjects.MoneyFromString(req.Anticipo)
	if err != nil {
		return nil, err
	}
	var anticipoDate *valueobjects.Date
	if req.AnticipoDate != "" {
		t, err := time.Parse("2006-01-02", req.AnticipoDate)
		if err != nil {
			return nil, err
		}
		ad := valueobjects.Date(t)
		anticipoDate = &ad
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
		CompanyID:           a.companyID(),
		Number:              "",
		SupplierID:          supplierID,
		CustomerID:          customerID,
		CreditCardID:        creditCardID,
		OrderType:           orderType,
		CurrencyCode:        cc,
		ExchangeRate:        rate,
		OrderDate:           valueobjects.Date(orderDate),
		SupplierOrderNumber: req.SupplierOrderNumber,
		CostUSD:             costUSD,
		SalePricePEN:        salePricePEN,
		Anticipo:            anticipo,
		AnticipoDate:        anticipoDate,
		Notes:               req.Notes,
		Items:               items,
	}
	po, err := a.purchasingSvc.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return toPurchaseOrderDTO(po), nil
}

// MarkPurchaseReceivedRequest marks a purchase order as received.
type MarkPurchaseReceivedRequest struct {
	ID          string `json:"id"`
	ArrivalDate string `json:"arrivalDate"`
}

// MarkPurchaseReceived records that the goods physically arrived (step 2
// of the order flow) and injects them into inventory as batches.
func (a *App) MarkPurchaseReceived(req MarkPurchaseReceivedRequest) (*PurchaseOrderDTO, error) {
	ctx := a.Context()
	pid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	var ad valueobjects.Date
	if req.ArrivalDate != "" {
		t, err := time.Parse("2006-01-02", req.ArrivalDate)
		if err != nil {
			return nil, err
		}
		ad = valueobjects.Date(t)
	} else {
		ad = valueobjects.Date(time.Now().UTC())
	}
	if err := a.purchasingSvc.MarkAsReceived(ctx, pid, ad); err != nil {
		return nil, err
	}
	po, err := a.purchasingSvc.GetByID(ctx, pid)
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
		CompanyID:   a.companyID(),
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

// MarkPurchaseFaultyRequest marks a customer order as faulty ("Llegó en
// mal estado"), voiding it and refunding its down payments.
type MarkPurchaseFaultyRequest struct {
	ID          string `json:"id"`
	ArrivalDate string `json:"arrivalDate"`
	Reason      string `json:"reason"`
}

// MarkPurchaseFaulty voids a customer order that arrived in bad state
// and refunds every down payment recorded against it.
func (a *App) MarkPurchaseFaulty(req MarkPurchaseFaultyRequest) (*PurchaseOrderDTO, error) {
	pid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	var ad valueobjects.Date
	if req.ArrivalDate != "" {
		t, err := time.Parse("2006-01-02", req.ArrivalDate)
		if err != nil {
			return nil, err
		}
		ad = valueobjects.Date(t)
	} else {
		ad = valueobjects.Date(time.Now().UTC())
	}
	po, err := a.purchasingSvc.MarkFaulty(a.Context(), purchasing.FaultyInput{
		ID:          pid,
		ArrivalDate: ad,
		Reason:      req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return toPurchaseOrderDTO(po), nil
}

// RegisterCustomerOrderPaymentRequest records a down payment (anticipo)
// against a customer order.
type RegisterCustomerOrderPaymentRequest struct {
	PurchaseID   string `json:"purchaseId"`
	PaymentDate  string `json:"paymentDate"`
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
	ExchangeRate string `json:"exchangeRate"`
	Method       string `json:"method"`
	Reference    string `json:"reference"`
	Notes        string `json:"notes"`
}

// CustomerOrderPaymentDTO is the serializable view of a down payment.
type CustomerOrderPaymentDTO struct {
	ID              string `json:"id"`
	PurchaseOrderID string `json:"purchaseOrderId"`
	Number          string `json:"number"`
	PaymentDate     string `json:"paymentDate"`
	Amount          string `json:"amount"`
	Method          string `json:"method"`
	CurrencyCode    string `json:"currencyCode"`
	ExchangeRate    string `json:"exchangeRate"`
	Reference       string `json:"reference"`
	Notes           string `json:"notes"`
	Status          string `json:"status"`
	RefundedAmount  string `json:"refundedAmount"`
	RefundedAt      string `json:"refundedAt"`
	RefundReason    string `json:"refundReason"`
}

func toCustomerOrderPaymentDTO(pm *purchasing.CustomerOrderPayment) *CustomerOrderPaymentDTO {
	dto := &CustomerOrderPaymentDTO{
		ID:              pm.ID.String(),
		PurchaseOrderID: pm.PurchaseOrderID.String(),
		Number:          pm.Number,
		PaymentDate:     pm.PaymentDate.Format("2006-01-02"),
		Amount:          pm.Amount.String(),
		Method:          pm.Method.String(),
		CurrencyCode:    pm.CurrencyCode.String(),
		ExchangeRate:    pm.ExchangeRate.String(),
		Reference:       pm.Reference,
		Notes:           pm.Notes,
		Status:          pm.Status,
		RefundedAmount:  pm.RefundedAmount.String(),
		RefundReason:    pm.RefundReason,
	}
	if pm.RefundedAt != nil {
		dto.RefundedAt = pm.RefundedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return dto
}

// RegisterCustomerOrderPayment persists a customer down payment and
// advances the order's anticipo.
func (a *App) RegisterCustomerOrderPayment(req RegisterCustomerOrderPaymentRequest) (*CustomerOrderPaymentDTO, error) {
	pid, err := uuid.Parse(req.PurchaseID)
	if err != nil {
		return nil, err
	}
	amount, err := valueobjects.MoneyFromString(req.Amount)
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
	pm, err := a.purchasingSvc.RegisterCustomerOrderPayment(a.Context(), pid, purchasing.CustomerPaymentInput{
		CompanyID:    a.companyID(),
		PaymentDate:  pd,
		Amount:       amount,
		CurrencyCode: cc,
		ExchangeRate: rate,
		Method:       method,
		Reference:    req.Reference,
		Notes:        req.Notes,
	})
	if err != nil {
		return nil, err
	}
	return toCustomerOrderPaymentDTO(pm), nil
}
