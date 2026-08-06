package administration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/administration"
	"vfinancy/backend/internal/domain/repositories"
)

type AuditService struct {
	repo repositories.AuditEventRepository
	log  *common.Logger
}

func NewAuditService(
	repo repositories.AuditEventRepository,
	log *common.Logger,
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

func (s *AuditService) Record(ctx context.Context, event *administration.AuditEvent) error {
	if event == nil {
		return fmt.Errorf("REQUIRED: event is required")
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to create audit event: %w", err)
	}

	return nil
}

func (s *AuditService) RecordLogin(ctx context.Context, companyID, userID, sessionID uuid.UUID, ipAddress, device string) error {
	if companyID == uuid.Nil {
		return fmt.Errorf("REQUIRED: company id is required")
	}
	if userID == uuid.Nil {
		return fmt.Errorf("REQUIRED: user id is required")
	}

	event := administration.NewAuditEvent(companyID, administration.EventLogin, "user logged in").
		WithUser(userID).
		WithIPAddress(ipAddress).
		WithDevice(device)

	if sessionID != uuid.Nil {
		event.WithSession(sessionID)
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to record login: %w", err)
	}

	s.log.InfoContext(ctx, "login recorded", "company_id", companyID, "user_id", userID)
	return nil
}

func (s *AuditService) RecordLogout(ctx context.Context, companyID, userID, sessionID uuid.UUID) error {
	if companyID == uuid.Nil {
		return fmt.Errorf("REQUIRED: company id is required")
	}
	if userID == uuid.Nil {
		return fmt.Errorf("REQUIRED: user id is required")
	}

	event := administration.NewAuditEvent(companyID, administration.EventLogout, "user logged out").
		WithUser(userID)

	if sessionID != uuid.Nil {
		event.WithSession(sessionID)
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to record logout: %w", err)
	}

	s.log.InfoContext(ctx, "logout recorded", "company_id", companyID, "user_id", userID)
	return nil
}

func (s *AuditService) RecordLoginFailed(ctx context.Context, companyID uuid.UUID, username, ipAddress, device, reason string) error {
	if companyID == uuid.Nil {
		return fmt.Errorf("REQUIRED: company id is required")
	}
	if username == "" {
		return fmt.Errorf("REQUIRED: username is required")
	}

	metadata, _ := json.Marshal(map[string]string{
		"username": username,
		"reason":   reason,
	})

	event := administration.NewAuditEvent(companyID, administration.EventLoginFailed, "login attempt failed").
		WithIPAddress(ipAddress).
		WithDevice(device).
		WithMetadata(metadata)

	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to record failed login: %w", err)
	}

	s.log.InfoContext(ctx, "failed login recorded", "company_id", companyID, "username", username)
	return nil
}

func (s *AuditService) RecordPasswordChange(ctx context.Context, companyID, userID uuid.UUID) error {
	if companyID == uuid.Nil {
		return fmt.Errorf("REQUIRED: company id is required")
	}
	if userID == uuid.Nil {
		return fmt.Errorf("REQUIRED: user id is required")
	}

	event := administration.NewAuditEvent(companyID, administration.EventPasswordChange, "password changed").
		WithUser(userID)

	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to record password change: %w", err)
	}

	s.log.InfoContext(ctx, "password change recorded", "company_id", companyID, "user_id", userID)
	return nil
}

func (s *AuditService) RecordConfigUpdate(ctx context.Context, companyID, userID uuid.UUID, key string) error {
	if companyID == uuid.Nil {
		return fmt.Errorf("REQUIRED: company id is required")
	}
	if userID == uuid.Nil {
		return fmt.Errorf("REQUIRED: user id is required")
	}
	if key == "" {
		return fmt.Errorf("REQUIRED: key is required")
	}

	metadata, _ := json.Marshal(map[string]string{
		"key": key,
	})

	event := administration.NewAuditEvent(companyID, administration.EventConfigUpdate, "configuration updated").
		WithUser(userID).
		WithTarget("setting", uuid.Nil).
		WithMetadata(metadata)

	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to record config update: %w", err)
	}

	s.log.InfoContext(ctx, "config update recorded", "company_id", companyID, "user_id", userID, "key", key)
	return nil
}

func (s *AuditService) List(ctx context.Context, companyID uuid.UUID, filter repositories.AuditEventFilter) ([]*administration.AuditEvent, int, error) {
	if companyID == uuid.Nil {
		return nil, 0, fmt.Errorf("REQUIRED: company id is required")
	}

	events, total, err := s.repo.List(ctx, companyID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit events: %w", err)
	}

	return events, total, nil
}
