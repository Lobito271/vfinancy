package bindings

import (
	"time"
	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/internal/utils"
)

// InventoryBatchDTO is the serializable view of a stock batch.
type InventoryBatchDTO struct {
	ID              string `json:"id"`
	ProductID       string `json:"productId"`
	WarehouseID     string `json:"warehouseId"`
	LotNumber       string `json:"lotNumber"`
	ArrivalDate     string `json:"arrivalDate"`
	ExpiryDate      string `json:"expiryDate"`
	InitialQuantity string `json:"initialQuantity"`
	CurrentQuantity string `json:"currentQuantity"`
	UnitCost        string `json:"unitCost"`
	CurrencyCode    string `json:"currencyCode"`
	Status          string `json:"status"`
	MaxSaleDate     string `json:"maxSaleDate"`
	IsClearance     bool   `json:"isClearance"`
}

// InventoryMovementDTO is the serializable view of a stock movement.
type InventoryMovementDTO struct {
	ID          string `json:"id"`
	ProductID   string `json:"productId"`
	WarehouseID string `json:"warehouseId"`
	BatchID     string `json:"batchId"`
	Type        string `json:"type"`
	Quantity    string `json:"quantity"`
	UnitCost    string `json:"unitCost"`
	OccurredAt  string `json:"occurredAt"`
	Notes       string `json:"notes"`
}

func toInventoryBatchDTO(today valueobjects.Date, b *inventory.InventoryBatch) *InventoryBatchDTO {
	expiry := ""
	if b.ExpiryDate != nil {
		expiry = b.ExpiryDate.Format("2006-01-02")
	}
	return &InventoryBatchDTO{
		ID:              b.ID.String(),
		ProductID:       b.ProductID.String(),
		WarehouseID:     b.WarehouseID.String(),
		LotNumber:       b.LotNumber.String(),
		ArrivalDate:     b.ArrivalDate.Format("2006-01-02"),
		ExpiryDate:      expiry,
		InitialQuantity: b.InitialQuantity.String(),
		CurrentQuantity: b.CurrentQuantity.String(),
		UnitCost:        b.UnitCost.String(),
		CurrencyCode:    b.CurrencyCode.String(),
		Status:          b.Status,
		MaxSaleDate:     b.MaximumSaleDate().Format("2006-01-02"),
		IsClearance:     b.IsClearance(today),
	}
}

func toInventoryMovementDTO(m *inventory.InventoryMovement) *InventoryMovementDTO {
	batchID := ""
	if m.BatchID != nil {
		batchID = m.BatchID.String()
	}
	return &InventoryMovementDTO{
		ID:          m.ID.String(),
		ProductID:   m.ProductID.String(),
		WarehouseID: m.WarehouseID.String(),
		BatchID:     batchID,
		Type:        m.Type.String(),
		Quantity:    m.Quantity.String(),
		UnitCost:    m.UnitCost.String(),
		OccurredAt:  m.OccurredAt.Format(time.RFC3339),
		Notes:       m.Notes,
	}
}

// ListInventoryBatchesRequest filters the batch listing.
type ListInventoryBatchesRequest struct {
	OnlyClearance bool `json:"onlyClearance"`
	PaginationRequest
}

// WarehouseDTO is the serializable view of a warehouse.
type WarehouseDTO struct {
	ID              string `json:"id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	IsDefault       bool   `json:"isDefault"`
	AllowsClearance bool   `json:"allowsClearance"`
	IsActive        bool   `json:"isActive"`
}

// ListWarehouses returns the warehouses of the active company.
func (a *App) ListWarehouses() ([]WarehouseDTO, error) {
	rows, err := a.db.QueryContext(a.Context(),
		`SELECT id, code, name, is_default, allows_clearance, is_active
		   FROM warehouses
		  WHERE company_id = $1 AND deleted_at IS NULL
		  ORDER BY is_default DESC, name`, a.companyID())
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	defer rows.Close()
	out := make([]WarehouseDTO, 0, 8)
	for rows.Next() {
		var w WarehouseDTO
		if err := rows.Scan(&w.ID, &w.Code, &w.Name, &w.IsDefault, &w.AllowsClearance, &w.IsActive); err != nil {
			return nil, utils.ProcessError(err)
		}
		out = append(out, w)
	}
	if err := rows.Err(); err != nil {
		return nil, utils.ProcessError(err)
	}
	return out, nil
}

// ListInventoryBatches returns paged stock batches.
func (a *App) ListInventoryBatches(req ListInventoryBatchesRequest) (PageResult, error) {
	ctx := a.Context()
	filter := inventory.InventoryBatchFilter{
		CompanyID:     a.companyIDPtr(),
		OnlyClearance: req.OnlyClearance,
		PageRequest:   req.toPageRequest(),
	}
	page, err := a.inventorySvc.ListBatches(ctx, filter)
	if err != nil {
		return PageResult{}, utils.ProcessError(err)
	}
	today := valueobjects.Date(time.Now().UTC())
	items := make([]*InventoryBatchDTO, 0, len(page.Items))
	for _, b := range page.Items {
		items = append(items, toInventoryBatchDTO(today, b))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// ListInventoryMovementsRequest filters the movement listing.
type ListInventoryMovementsRequest struct {
	ProductID string `json:"productId"`
	PaginationRequest
}

// ListInventoryMovements returns paged stock movements.
func (a *App) ListInventoryMovements(req ListInventoryMovementsRequest) (PageResult, error) {
	ctx := a.Context()
	filter := inventory.InventoryMovementFilter{
		CompanyID:   a.companyIDPtr(),
		PageRequest: req.toPageRequest(),
	}
	productID, err := parseOptionalUUID(req.ProductID)
	if err != nil {
		return PageResult{}, utils.ProcessError(err)
	}
	filter.ProductID = productID
	page, err := a.inventorySvc.ListMovements(ctx, filter)
	if err != nil {
		return PageResult{}, utils.ProcessError(err)
	}
	items := make([]*InventoryMovementDTO, 0, len(page.Items))
	for _, m := range page.Items {
		items = append(items, toInventoryMovementDTO(m))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// GetClearanceCandidates returns all batches currently on clearance.
func (a *App) GetClearanceCandidates() ([]*InventoryBatchDTO, error) {
	ctx := a.Context()
	now := time.Now().UTC()
	batches, err := a.inventorySvc.GenerateClearanceCandidates(ctx, a.companyID(), now)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	today := valueobjects.Date(now)
	items := make([]*InventoryBatchDTO, 0, len(batches))
	for _, b := range batches {
		items = append(items, toInventoryBatchDTO(today, b))
	}
	return items, nil
}

// ReceiveStockRequest receives stock into a new batch.
type ReceiveStockRequest struct {
	ProductID    string `json:"productId"`
	WarehouseID  string `json:"warehouseId"`
	LotNumber    string `json:"lotNumber"`
	ArrivalDate  string `json:"arrivalDate"`
	Quantity     string `json:"quantity"`
	UnitCost     string `json:"unitCost"`
	CurrencyCode string `json:"currencyCode"`
}

// ReceiveStock persists a new inventory batch.
func (a *App) ReceiveStock(req ReceiveStockRequest) (*InventoryBatchDTO, error) {
	ctx := a.Context()
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	lot, err := valueobjects.NewLotNumber(req.LotNumber)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	arrival, err := time.Parse("2006-01-02", req.ArrivalDate)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	qty, err := valueobjects.QuantityFromString(req.Quantity)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	cost, err := valueobjects.MoneyFromString(req.UnitCost)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	cc, err := valueobjects.NewCurrencyCode(req.CurrencyCode)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	batch, err := a.inventorySvc.Receive(ctx, inventory.ReceiveInput{
		CompanyID:    a.companyID(),
		ProductID:    productID,
		WarehouseID:  warehouseID,
		LotNumber:    lot,
		ArrivalDate:  valueobjects.Date(arrival),
		Quantity:     qty,
		UnitCost:     cost,
		CurrencyCode: cc,
	})
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	today := valueobjects.Date(time.Now().UTC())
	return toInventoryBatchDTO(today, batch), nil
}

// IssueStockRequest removes stock from a batch (sale, write-off...).
type IssueStockRequest struct {
	BatchID  string `json:"batchId"`
	Quantity string `json:"quantity"`
}

// IssueStock records an inventory exit from a batch.
func (a *App) IssueStock(req IssueStockRequest) error {
	ctx := a.Context()
	batchID, err := uuid.Parse(req.BatchID)
	if err != nil {
		return utils.ProcessError(err)
	}
	qty, err := valueobjects.QuantityFromString(req.Quantity)
	if err != nil {
		return utils.ProcessError(err)
	}
	ref, err := valueobjects.NewReference(enums.ReferenceTypeManual, batchID)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.inventorySvc.Issue(ctx, inventory.IssueInput{
		BatchID:   batchID,
		Quantity:  qty,
		Reference: ref,
	}))
}

// AdjustStockRequest corrects the stock level of a batch. Delta is
// signed: positive adds stock, negative removes it.
type AdjustStockRequest struct {
	BatchID string `json:"batchId"`
	Delta   string `json:"delta"`
	Reason  string `json:"reason"`
}

// AdjustStock records a manual inventory adjustment.
func (a *App) AdjustStock(req AdjustStockRequest) error {
	ctx := a.Context()
	batchID, err := uuid.Parse(req.BatchID)
	if err != nil {
		return utils.ProcessError(err)
	}
	delta, err := valueobjects.QuantityFromString(req.Delta)
	if err != nil {
		return utils.ProcessError(err)
	}
	ref, err := valueobjects.NewReference(enums.ReferenceTypeAdjustment, batchID)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.inventorySvc.Adjust(ctx, inventory.AdjustInput{
		BatchID:   batchID,
		Delta:     delta,
		Reason:    req.Reason,
		Reference: ref,
	}))
}

// VoidStockRequest cancels a mistaken stock receipt.
type VoidStockRequest struct {
	BatchID string `json:"batchId"`
	Reason  string `json:"reason"`
}

// VoidStock cancels a mistaken stock receipt. The batch row is kept,
// its remaining quantity is zeroed and its status becomes "voided".
func (a *App) VoidStock(req VoidStockRequest) error {
	ctx := a.Context()
	batchID, err := uuid.Parse(req.BatchID)
	if err != nil {
		return utils.ProcessError(err)
	}
	return utils.ProcessError(a.inventorySvc.Void(ctx, inventory.VoidInput{
		BatchID: batchID,
		Reason:  req.Reason,
	}))
}
