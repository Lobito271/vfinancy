package administration

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"
	"time"

	"github.com/google/uuid"

)

type AuditEventFilter struct {
	EventType string
	UserID    *uuid.UUID
	From      *time.Time
	To        *time.Time
	repositories.PageRequest
}

type AuditEventRepository interface {
	Create(ctx context.Context, event *AuditEvent) error
	List(ctx context.Context, companyID uuid.UUID, filter AuditEventFilter) ([]*AuditEvent, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error)
}
