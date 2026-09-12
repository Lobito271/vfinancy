package bindings

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/purchasing"
)

type PurchaseItemDTO struct {
	ID               string  `json:"id"`
	ProductID        string  `json:"productId"`
	LineNumber       int     `json:"lineNumber"`
	Description      string  `json:"description"`
	UnitCode         string  `json:"unitCode"`
	Quantity         float64 `json:"quantity"`
	QuantityReceived float64 `json:"quantityReceived"`
	UnitCostUSD      float64 `json:"unitCostUsd"`
	LineTotalUSD     float64 `json:"lineTotalUsd"`
	SalePricePen     float64 `json:"salePricePen"`
}

type PurchaseOrderDTO struct {
	ID                 string             `json:"id"`
	Number             string             `json:"number"`
	OrderDate          string             `json:"orderDate"`
	ExpectedDate       string             `json:"expectedDate"`
	ReceivedDate       string             `json:"receivedDate"`
	ArrivalDate        string             `json:"arrivalDate"`
	Status             string             `json:"status"`
	CurrencyCode       string             `json:"currencyCode"`
	ExchangeRate       float64            `json:"exchangeRate"`
	Notes              string             `json:"notes"`
	OrderType          string             `json:"orderType"`
	CustomerID         string             `json:"customerId"`
	CreditCardID       string             `json:"creditCardId"`
	CostUSD            float64            `json:"costUsd"`
	SalePricePen       float64            `json:"salePricePen"`
	RealCostPen        float64            `json:"realCostPen"`
	RefundAmount       float64            `json:"refundAmount"`
	Faulty             bool               `json:"faulty"`
	FaultyReason       string             `json:"faultyReason"`
	CancelledAt        string             `json:"cancelledAt"`
	CancelledReason    string             `json:"cancelledReason"`
	Items              []PurchaseItemDTO  `json:"items"`
}

func purchaseDTO(po *purchasing.PurchaseOrder) PurchaseOrderDTO {
	items := make([]PurchaseItemDTO, 0, len(po.Items))
	for _, it := range po.Items {
		items = append(items, PurchaseItemDTO{
			ID:               it.ID.String(),
			ProductID:        uuidPtrString(it.ProductID),
			LineNumber:       it.LineNumber,
			Description:      it.Description,
			UnitCode:         it.UnitCode,
			Quantity:         quantityFloat(it.QuantityOrdered),
			QuantityReceived: quantityFloat(it.QuantityReceived),
			UnitCostUSD:      moneyFloat(it.UnitCostUSD),
			LineTotalUSD:     moneyFloat(it.LineTotalUSD),
			SalePricePen:     moneyFloat(it.SalePricePen),
		})
	}
	return PurchaseOrderDTO{
		ID:                 po.ID.String(),
		Number:             po.Number,
		OrderDate:          dayStr(po.OrderDate),
		ExpectedDate:       dayStrPtr(po.ExpectedDate),
		ReceivedDate:       dayStrPtr(po.ReceivedDate),
		ArrivalDate:        dayStrPtr(po.ArrivalDate),
		Status:             string(po.Status),
		CurrencyCode:       po.CurrencyCode,
		ExchangeRate:       po.ExchangeRate.Decimal().InexactFloat64(),
		Notes:              po.Notes,
		OrderType:          string(po.OrderType),
		CustomerID:         uuidPtrString(po.CustomerID),
		CreditCardID:       uuidPtrString(po.CreditCardID),
		CostUSD:            moneyFloat(po.CostUSD),
		SalePricePen:       moneyFloat(po.SalePricePen),
		RealCostPen:        moneyFloat(po.RealCostPen),
		RefundAmount:       moneyFloat(po.RefundAmount),
		Faulty:             po.Faulty,
		FaultyReason:       po.FaultyReason,
		CancelledAt:        dayStrPtr(po.CancelledAt),
		CancelledReason:    po.CancelledReason,
		Items:              items,
	}
}

type PurchaseFilterRequest struct {
	PaginationRequest
	Search       string `json:"search"`
	Status       string `json:"status"`
	OrderType    string `json:"orderType"`
	CreditCardID string `json:"creditCardId"`
	ImportLotID  string `json:"importLotId"`
	From         string `json:"from"`
	To           string `json:"to"`
}

// ListPurchaseOrders supports the advanced filters drawer (lot, date
// range, credit card).
func (a *App) ListPurchaseOrders(req PurchaseFilterRequest) (PageResult, error) {
	cardID, err := parseOptionalUUID(req.CreditCardID)
	if err != nil {
		return PageResult{}, err
	}
	lotID, err := parseOptionalUUID(req.ImportLotID)
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
	page, err := a.purchasingSvc.List(a.Context(), purchasing.PurchaseFilter{
		Search:       req.Search,
		Status:       req.Status,
		OrderType:    req.OrderType,
		CreditCardID: cardID,
		ImportLotID:  lotID,
		From:         from,
		To:           to,
		PageRequest:  req.toPageRequest(),
	})
	if err != nil {
		return PageResult{}, err
	}
	items := make([]PurchaseOrderDTO, 0, len(page.Items))
	for _, po := range page.Items {
		items = append(items, purchaseDTO(po))
	}
	return PageResult{Items: items, Total: page.Total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetPurchaseOrder returns one order with its items.
func (a *App) GetPurchaseOrder(id string) (PurchaseOrderDTO, error) {
	oid, err := parseUUID(id)
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	po, err := a.purchasingSvc.GetByID(a.Context(), oid)
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	return purchaseDTO(po), nil
}

type PurchaseItemRequest struct {
	ProductID    string  `json:"productId"`
	Description  string  `json:"description"`
	Quantity     float64 `json:"quantity"`
	UnitCostUSD  float64 `json:"unitCostUsd"`
	SalePricePen float64 `json:"salePricePen"`
}

type CreatePurchaseRequest struct {
	OrderType    string                `json:"orderType"`
	CustomerID   string                `json:"customerId"`
	CreditCardID string                `json:"creditCardId"`
	ExchangeRate float64               `json:"exchangeRate"`
	OrderDate    string                `json:"orderDate"`
	ExpectedDate string                `json:"expectedDate"`
	Notes        string                `json:"notes"`
	Items        []PurchaseItemRequest `json:"items"`
}

// CreatePurchase registers a USD order (general or client) tied to a
// credit card, applying the TC and the import cost factor.
func (a *App) CreatePurchase(req CreatePurchaseRequest) (PurchaseOrderDTO, error) {
	in, err := a.purchaseInput(req)
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	po, err := a.purchasingSvc.Create(a.Context(), in)
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	return purchaseDTO(po), nil
}

func (a *App) purchaseInput(req CreatePurchaseRequest) (purchasing.CreateInput, error) {
	cardID, err := parseOptionalUUID(req.CreditCardID)
	if err != nil {
		return purchasing.CreateInput{}, err
	}
	customerID, err := parseOptionalUUID(req.CustomerID)
	if err != nil {
		return purchasing.CreateInput{}, err
	}
	orderDate := time.Now().UTC()
	if req.OrderDate != "" {
		t, err := parseDay(req.OrderDate)
		if err != nil {
			return purchasing.CreateInput{}, err
		}
		orderDate = t
	}
	expected, err := parseDayPtr(req.ExpectedDate)
	if err != nil {
		return purchasing.CreateInput{}, err
	}
	rate, err := valueobjects.ExchangeRateFromDecimal(rated(req.ExchangeRate))
	if err != nil {
		return purchasing.CreateInput{}, err
	}
	items := make([]purchasing.CreateItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		pid, err := parseOptionalUUID(it.ProductID)
		if err != nil {
			return purchasing.CreateInput{}, err
		}
		qty, err := quantityFromFloat(it.Quantity)
		if err != nil {
			return purchasing.CreateInput{}, err
		}
		cost, err := moneyFromFloat(it.UnitCostUSD)
		if err != nil {
			return purchasing.CreateInput{}, err
		}
		price, err := moneyFromFloat(it.SalePricePen)
		if err != nil {
			return purchasing.CreateInput{}, err
		}
		items = append(items, purchasing.CreateItemInput{
			ProductID:    pid,
			Description:  it.Description,
			Quantity:     qty,
			UnitCostUSD:  cost,
			SalePricePen: price,
		})
	}
	return purchasing.CreateInput{
		OrderType:    enums.OrderType(req.OrderType),
		CustomerID:   customerID,
		CreditCardID: cardID,
		ExchangeRate: rate,
		OrderDate:    orderDate,
		ExpectedDate: expected,
		Notes:        req.Notes,
		Items:        items,
	}, nil
}

// MarkPurchaseReceived routes the goods into the main warehouse and
// starts the permanence clock for the clearance rule.
func (a *App) MarkPurchaseReceived(id string, receivedDate string) error {
	oid, err := parseUUID(id)
	if err != nil {
		return err
	}
	at := valueobjects.NewDateFromTime(time.Now().UTC())
	if receivedDate != "" {
		d, err := parseDate(receivedDate)
		if err != nil {
			return err
		}
		at = d
	}
	return a.purchasingSvc.MarkAsReceived(a.Context(), oid, at)
}

type CancelPurchaseRequest struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// CancelPurchase annuls the order: reverts stock and releases the card
// charge; when the billing cycle was already settled the amount is
// recorded as a refund credit (edge case 4.1).
func (a *App) CancelPurchase(req CancelPurchaseRequest) (PurchaseOrderDTO, error) {
	oid, err := parseUUID(req.ID)
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	po, err := a.purchasingSvc.Cancel(a.Context(), oid, req.Reason)
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	return purchaseDTO(po), nil
}

// MarkPurchaseFaulty records an arrival in bad condition and annuls the
// order with a full refund.
func (a *App) MarkPurchaseFaulty(req CancelPurchaseRequest) (PurchaseOrderDTO, error) {
	oid, err := parseUUID(req.ID)
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	po, err := a.purchasingSvc.MarkFaulty(a.Context(), purchasing.FaultyInput{ID: oid, Reason: req.Reason})
	if err != nil {
		return PurchaseOrderDTO{}, err
	}
	return purchaseDTO(po), nil
}

type ImportLotDTO struct {
	ID             string             `json:"id"`
	Code           string             `json:"code"`
	Description    string             `json:"description"`
	Status         string             `json:"status"`
	TotalUSD       float64            `json:"totalUsd"`
	OverLimit      bool               `json:"overLimit"`
	CustomsLimitUSD float64           `json:"customsLimitUsd"`
	Members        []PurchaseOrderDTO `json:"members"`
}

// CreateImportLot groups orders for customs control. A total above the
// customs cap returns the lot with OverLimit=true (non-blocking
// warning; the UI asks for explicit confirmation).
func (a *App) CreateImportLot(description string, purchaseIDs []string) (ImportLotDTO, error) {
	ids, err := parseUUIDs(purchaseIDs)
	if err != nil {
		return ImportLotDTO{}, err
	}
	lot, total, over, err := a.purchasingSvc.CreateImportLot(a.Context(), purchasing.ImportLotInput{Description: description, PurchaseIDs: ids})
	if err != nil {
		return ImportLotDTO{}, err
	}
	return lotDTO(a, lot, total, over), nil
}

// AddToImportLot adds orders to an existing lot and re-evaluates the
// customs cap.
func (a *App) AddToImportLot(lotID string, purchaseIDs []string) (ImportLotDTO, error) {
	lid, err := parseUUID(lotID)
	if err != nil {
		return ImportLotDTO{}, err
	}
	ids, err := parseUUIDs(purchaseIDs)
	if err != nil {
		return ImportLotDTO{}, err
	}
	total, over, err := a.purchasingSvc.AddToImportLot(a.Context(), lid, ids)
	if err != nil {
		return ImportLotDTO{}, err
	}
	lot, _, _, err := a.purchasingSvc.GetImportLot(a.Context(), lid)
	if err != nil {
		return ImportLotDTO{}, err
	}
	return lotDTO(a, lot, total, over), nil
}

// RemoveFromImportLot detaches an order from its lot.
func (a *App) RemoveFromImportLot(lotID, purchaseID string) (ImportLotDTO, error) {
	lid, err := parseUUID(lotID)
	if err != nil {
		return ImportLotDTO{}, err
	}
	pid, err := parseUUID(purchaseID)
	if err != nil {
		return ImportLotDTO{}, err
	}
	total, err := a.purchasingSvc.RemoveFromImportLot(a.Context(), lid, pid)
	if err != nil {
		return ImportLotDTO{}, err
	}
	lot, _, over, err := a.purchasingSvc.GetImportLot(a.Context(), lid)
	if err != nil {
		return ImportLotDTO{}, err
	}
	return lotDTO(a, lot, total, over), nil
}

// CloseImportLot marks a customs lot as closed.
func (a *App) CloseImportLot(id string) (ImportLotDTO, error) {
	lid, err := parseUUID(id)
	if err != nil {
		return ImportLotDTO{}, err
	}
	if _, err := a.purchasingSvc.CloseImportLot(a.Context(), lid); err != nil {
		return ImportLotDTO{}, err
	}
	lot, total, over, err := a.purchasingSvc.GetImportLot(a.Context(), lid)
	if err != nil {
		return ImportLotDTO{}, err
	}
	return lotDTO(a, lot, total, over), nil
}

// ListImportLots returns the import groups.
func (a *App) ListImportLots(req PaginationRequest, search string) (PageResult, error) {
	page, err := a.purchasingSvc.ListImportLots(a.Context(), purchasing.ImportLotFilter{Search: search, PageRequest: req.toPageRequest()})
	if err != nil {
		return PageResult{}, err
	}
	items := make([]ImportLotDTO, 0, len(page.Items))
	for _, lot := range page.Items {
		_, total, over, err := a.purchasingSvc.GetImportLot(a.Context(), lot.ID)
		if err != nil {
			return PageResult{}, err
		}
		items = append(items, lotDTO(a, lot, total, over))
	}
	return PageResult{Items: items, Total: page.Total, Page: req.Page, PageSize: req.PageSize}, nil
}

// ListLotMembers returns the orders grouped in a lot.
func (a *App) ListLotMembers(lotID string) ([]PurchaseOrderDTO, error) {
	lid, err := parseUUID(lotID)
	if err != nil {
		return nil, err
	}
	members, err := a.purchasingSvc.ListLotMembers(a.Context(), lid)
	if err != nil {
		return nil, err
	}
	items := make([]PurchaseOrderDTO, 0, len(members))
	for _, po := range members {
		items = append(items, purchaseDTO(po))
	}
	return items, nil
}

func lotDTO(a *App, lot *purchasing.ImportLot, total valueobjects.Money, over bool) ImportLotDTO {
	return ImportLotDTO{
		ID:              lot.ID.String(),
		Code:            lot.Code,
		Description:     lot.Description,
		Status:          string(lot.Status),
		TotalUSD:        moneyFloat(total),
		OverLimit:       over,
		CustomsLimitUSD: a.customsLimit(a.Context()),
	}
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(ids))
	for _, s := range ids {
		id, err := parseUUID(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func rated(f float64) decimal.Decimal {
	if f <= 0 {
		return decimal.NewFromFloat(1)
	}
	return decimal.NewFromFloat(f)
}
