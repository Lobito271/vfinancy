package bindings

import (
	"time"
)

// GetClearancePreview returns the batches currently in or near
// clearance (dashboard / inventory badges).
func (a *App) ListClearanceProducts() ([]InventoryBatchDTO, error) {
	batches, err := a.inventorySvc.GenerateClearanceCandidates(a.Context(), time.Now().UTC())
	if err != nil {
		return nil, err
	}
	items := make([]InventoryBatchDTO, 0, len(batches))
	for _, b := range batches {
		dto, err := batchDTO(a, b)
		if err != nil {
			return nil, err
		}
		items = append(items, dto)
	}
	return items, nil
}
