package bindings

import (
	"time"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/administration"
	"vfinancy/backend/internal/utils"
)

type AuditEventDTO struct {
	ID          string `json:"id"`
	EventType   string `json:"eventType"`
	UserID      string `json:"userId"`
	Description string `json:"description"`
	IPAddress   string `json:"ipAddress"`
	Device      string `json:"device"`
	OccurredAt  string `json:"occurredAt"`
}

type AuditLogResult struct {
	Events []AuditEventDTO `json:"events"`
	Total  int             `json:"total"`
}

func (a *App) GetAuditLog(page, pageSize int, eventType string) (*AuditLogResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	filter := administration.AuditEventFilter{EventType: eventType, PageRequest: repositories.PageRequest{
		Limit: pageSize, Offset: (page - 1) * pageSize,
	}}
	events, total, err := a.auditSvc.List(a.Context(), a.companyID(), filter)
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	result := &AuditLogResult{Events: make([]AuditEventDTO, len(events)), Total: total}
	for i, event := range events {
		result.Events[i] = AuditEventDTO{ID: event.ID.String(), EventType: string(event.EventType),
			Description: event.Description, IPAddress: event.IPAddress, Device: event.Device,
			OccurredAt: event.OccurredAt.Format(time.RFC3339)}
		if event.UserID != nil {
			result.Events[i].UserID = event.UserID.String()
		}
	}
	return result, nil
}
