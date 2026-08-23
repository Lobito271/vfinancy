package administration

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/shared/logger"
)

type AuditService struct {
	repo AuditEventRepository
	log  *logger.Logger
}

func NewAuditService(
	repo AuditEventRepository,
	log *logger.Logger,
) *AuditService {
	if repo == nil {
		panic("administration: nil audit event repository")
	}
	if log == nil {
		panic("administration: nil logger")
	}
	return &AuditService{
		repo: repo,
		log:  log,
	}
}

func (s *AuditService) Record(ctx context.Context, event *AuditEvent) error {
	if event == nil {
		return derrors.New("REQUIRED", "event is required")
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to create audit event: %w", err)
	}

	return nil
}

func (s *AuditService) List(ctx context.Context, companyID uuid.UUID, filter AuditEventFilter) ([]*AuditEvent, int, error) {
	if companyID == uuid.Nil {
		return nil, 0, derrors.New("REQUIRED", "company id is required")
	}

	events, total, err := s.repo.List(ctx, companyID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit events: %w", err)
	}

	return events, total, nil
}
