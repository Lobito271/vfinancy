package bindings

import (
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/features/shipment"
)

type ShipmentDTO struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	SaleID      string `json:"saleId"`
	CustomerID  string `json:"customerId"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func shipmentDTO(s *shipment.Shipment) ShipmentDTO {
	return ShipmentDTO{
		ID:          s.ID.String(),
		Code:        s.Code,
		SaleID:      uuidPtrString(s.SaleID),
		CustomerID:  uuidPtrString(s.CustomerID),
		Description: s.Description,
		Notes:       s.Notes,
		Status:      s.Status.String(),
		CreatedAt:   dayStr(s.CreatedAt),
		UpdatedAt:   dayStr(s.UpdatedAt),
	}
}

type SaveShipmentRequest struct {
	ID          string `json:"id"`
	SaleID      string `json:"saleId"`
	CustomerID  string `json:"customerId"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	Status      string `json:"status"`
}

func (a *App) shipmentInput(req SaveShipmentRequest) (shipment.CreateInput, error) {
	customerID, err := parseOptionalUUID(req.CustomerID)
	if err != nil {
		return shipment.CreateInput{}, err
	}
	saleID, err := parseOptionalUUID(req.SaleID)
	if err != nil {
		return shipment.CreateInput{}, err
	}
	return shipment.CreateInput{
		CustomerID:  customerID,
		SaleID:      saleID,
		Description: req.Description,
		Notes:       req.Notes,
	}, nil
}

// CreateShipment registers a new Envío with an auto-generated four-digit
// code that serves as the reference to share with the carrier.
func (a *App) CreateShipment(req SaveShipmentRequest) (ShipmentDTO, error) {
	in, err := a.shipmentInput(req)
	if err != nil {
		return ShipmentDTO{}, err
	}
	s, err := a.shipmentSvc.Create(a.Context(), in)
	if err != nil {
		return ShipmentDTO{}, err
	}
	return shipmentDTO(s), nil
}

// UpdateShipment edits the mutable shipment fields. The code is
// immutable after creation.
func (a *App) UpdateShipment(req SaveShipmentRequest) (ShipmentDTO, error) {
	id, err := parseUUID(req.ID)
	if err != nil {
		return ShipmentDTO{}, err
	}
	customerID, err := parseOptionalUUID(req.CustomerID)
	if err != nil {
		return ShipmentDTO{}, err
	}
	saleID, err := parseOptionalUUID(req.SaleID)
	if err != nil {
		return ShipmentDTO{}, err
	}
	description, notes := req.Description, req.Notes
	status := enums.ShipmentStatus(req.Status)
	s, err := a.shipmentSvc.Update(a.Context(), shipment.UpdateInput{
		ID:          id,
		CustomerID:  customerID,
		SaleID:      saleID,
		Description: &description,
		Notes:       &notes,
		Status:      &status,
	})
	if err != nil {
		return ShipmentDTO{}, err
	}
	return shipmentDTO(s), nil
}

// DeleteShipment soft-deletes an Envío preserving its history.
func (a *App) DeleteShipment(id string) error {
	sid, err := parseUUID(id)
	if err != nil {
		return err
	}
	return a.shipmentSvc.Delete(a.Context(), sid)
}

// GetShipment returns a single Envío.
func (a *App) GetShipment(id string) (ShipmentDTO, error) {
	sid, err := parseUUID(id)
	if err != nil {
		return ShipmentDTO{}, err
	}
	s, err := a.shipmentSvc.GetByID(a.Context(), sid)
	if err != nil {
		return ShipmentDTO{}, err
	}
	return shipmentDTO(s), nil
}

// ListShipments returns the Envíos matching the search/status filters.
func (a *App) ListShipments(req PaginationRequest, search string, status string) (PageResult, error) {
	page, err := a.shipmentSvc.List(a.Context(), shipment.ShipmentFilter{
		Search:      search,
		Status:      status,
		PageRequest: req.toPageRequest(),
	})
	if err != nil {
		return PageResult{}, err
	}
	items := make([]ShipmentDTO, 0, len(page.Items))
	for _, s := range page.Items {
		items = append(items, shipmentDTO(s))
	}
	return PageResult{Items: items, Total: page.Total, Page: req.Page, PageSize: req.PageSize}, nil
}
