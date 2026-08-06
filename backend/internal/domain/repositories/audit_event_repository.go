package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/administration"
)

type AuditEventFilter struct {
	EventType string
	UserID    *uuid.UUID
	From      *time.Time
	To        *time.Time
	PageRequest
}

type AuditEventRepository interface {
	Create(ctx context.Context, event *administration.AuditEvent) error
	List(ctx context.Context, companyID uuid.UUID, filter AuditEventFilter) ([]*administration.AuditEvent, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*administration.AuditEvent, error)
}
