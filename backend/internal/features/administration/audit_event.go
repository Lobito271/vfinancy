package administration

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditEventType string

const (
	EventConfigUpdate   AuditEventType = "CONFIG_UPDATE"
	EventBackupCreate   AuditEventType = "BACKUP_CREATE"
	EventExportData     AuditEventType = "EXPORT_DATA"
)

type AuditEvent struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	UserID      *uuid.UUID
	SessionID   *uuid.UUID
	EventType   AuditEventType
	TargetType  string
	TargetID    *uuid.UUID
	Description string
	Metadata    json.RawMessage
	IPAddress   string
	Device      string
	OccurredAt  time.Time
}

func NewAuditEvent(companyID uuid.UUID, eventType AuditEventType, description string) *AuditEvent {
	return &AuditEvent{
		ID:          uuid.New(),
		CompanyID:   companyID,
		EventType:   eventType,
		Description: description,
		OccurredAt:  time.Now(),
	}
}

func (e *AuditEvent) WithUser(userID uuid.UUID) *AuditEvent {
	e.UserID = &userID
	return e
}

func (e *AuditEvent) WithSession(sessionID uuid.UUID) *AuditEvent {
	e.SessionID = &sessionID
	return e
}

func (e *AuditEvent) WithTarget(targetType string, targetID uuid.UUID) *AuditEvent {
	e.TargetType = targetType
	e.TargetID = &targetID
	return e
}

func (e *AuditEvent) WithMetadata(metadata json.RawMessage) *AuditEvent {
	e.Metadata = metadata
	return e
}

func (e *AuditEvent) WithIPAddress(ip string) *AuditEvent {
	e.IPAddress = ip
	return e
}

func (e *AuditEvent) WithDevice(device string) *AuditEvent {
	e.Device = device
	return e
}
