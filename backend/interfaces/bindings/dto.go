package bindings

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
)

// PaginationRequest is the page/pageSize envelope sent by the frontend.
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// toPageRequest clamps the values into a repositories.PageRequest.
func (p PaginationRequest) toPageRequest() repositories.PageRequest {
	page := p.Page
	if page < 1 {
		page = 1
	}
	size := p.PageSize
	if size < 1 {
		size = 25
	}
	if size > 1000 {
		size = 1000
	}
	return repositories.PageRequest{Limit: size, Offset: (page - 1) * size}
}

// PageResult is the paginated response envelope sent to the frontend.
type PageResult struct {
	Items    interface{} `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func parseOptionalUUID(s string) (*uuid.UUID, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func uuidPtrString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// parseDate parses a "YYYY-MM-DD" string into a valueobjects.Date.
func parseDate(s string) (valueobjects.Date, error) {
	return valueobjects.DateFromString(strings.TrimSpace(s))
}

func parseDatePtr(s string) (*valueobjects.Date, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	d, err := parseDate(s)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// parseDay parses a "YYYY-MM-DD" string into a UTC timestamp.
func parseDay(s string) (time.Time, error) {
	return parseDate(s)
}

func parseDayPtr(s string) (*time.Time, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	t, err := parseDay(s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func dayStr(t time.Time) string {
	return t.Format("2006-01-02")
}

func dayStrPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return dayStr(*t)
}

func moneyFromFloat(f float64) (valueobjects.Money, error) {
	return valueobjects.MoneyFromFloat64(f)
}

func moneyPtrFromFloat(f float64) (*valueobjects.Money, error) {
	m, err := moneyFromFloat(f)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func moneyFloat(m valueobjects.Money) float64 {
	return m.Decimal().InexactFloat64()
}

func quantityFromFloat(f float64) (valueobjects.Quantity, error) {
	return valueobjects.QuantityFromDecimal(decimal.NewFromFloat(f))
}

func quantityFloat(q valueobjects.Quantity) float64 {
	return q.Decimal().InexactFloat64()
}

// batchDTO maps a batch with its product description and clearance
// deadline for the kardex table.
func batchDTO(a *App, b *inventory.InventoryBatch) (InventoryBatchDTO, error) {
	dto := InventoryBatchDTO{
		ID:                  b.ID.String(),
		ProductID:           b.ProductID.String(),
		PurchaseOrderItemID: uuidPtrString(b.PurchaseOrderItemID),
		ArrivalDate:         b.ArrivalDate.Format("2006-01-02"),
		Quantity:            quantityFloat(b.Quantity),
		OriginalQuantity:    quantityFloat(b.OriginalQuantity),
		UnitCost:            moneyFloat(b.UnitCost),
		Status:              string(b.Status),
		IsClearance:         b.IsClearance,
	}
	if p, err := a.productsSvc.GetByID(a.Context(), b.ProductID); err == nil {
		dto.ProductDescription = p.Description
	}
	days := a.inventorySvc.ClearanceDaysFor(a.Context())
	dto.MaxSaleDate = b.MaxSaleDate(days).Format("2006-01-02")
	return dto, nil
}
