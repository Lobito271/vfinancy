package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/identity"
	"vfinancy/backend/internal/domain/repositories"
)

type SessionService struct {
	sessions repositories.SessionRepository
	ttl      time.Duration
	log      *common.Logger
}

func NewSessionService(sessions repositories.SessionRepository, ttl time.Duration, log *common.Logger) *SessionService {
	if sessions == nil {
		panic("auth: nil sessions repository")
	}
	if log == nil {
		panic("auth: nil logger")
	}
	return &SessionService{
		sessions: sessions,
		ttl:      ttl,
		log:      log,
	}
}

func (s *SessionService) Create(ctx context.Context, userID uuid.UUID, ipAddress, userAgent, device string) (*identity.UserSession, error) {
	tokenBytes := make([]byte, 64)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("SESSION_FAILED: error generando token de sesión")
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	session, err := identity.NewUserSession(userID, token, s.ttl, ipAddress, userAgent, device)
	if err != nil {
		return nil, fmt.Errorf("SESSION_FAILED: %w", err)
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("SESSION_FAILED: error creando sesión")
	}

	s.log.WithContext(ctx).Info("session created", "session_id", session.ID, "user_id", userID)
	return session, nil
}

func (s *SessionService) Validate(ctx context.Context, token string) (*identity.UserSession, error) {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("SESSION_NOT_FOUND: sesión no encontrada")
	}

	now := time.Now().UTC()
	if !session.IsValid(now) {
		s.log.WithContext(ctx).Warn("invalid session", "session_id", session.ID, "active", session.IsActive, "expired", session.IsExpired(now), "locked", session.IsLocked())
		return nil, fmt.Errorf("SESSION_INVALID: la sesión no es válida")
	}

	session.RecordActivity(now)
	if err := s.sessions.Update(ctx, session); err != nil {
		s.log.WithContext(ctx).Error("failed to record session activity", "session_id", session.ID, "error", err)
	}

	return session, nil
}

func (s *SessionService) Destroy(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessions.Deactivate(ctx, sessionID); err != nil {
		return fmt.Errorf("SESSION_FAILED: error desactivando sesión")
	}
	s.log.WithContext(ctx).Info("session destroyed", "session_id", sessionID)
	return nil
}

func (s *SessionService) DestroyAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessions.DeactivateAll(ctx, userID); err != nil {
		return fmt.Errorf("SESSION_FAILED: error desactivando sesiones del usuario")
	}
	s.log.WithContext(ctx).Info("all sessions destroyed", "user_id", userID)
	return nil
}

func (s *SessionService) Lock(ctx context.Context, token string, reason string) error {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("SESSION_NOT_FOUND: sesión no encontrada")
	}

	now := time.Now().UTC()
	session.Lock(reason, now)
	if err := s.sessions.Update(ctx, session); err != nil {
		return fmt.Errorf("SESSION_FAILED: error bloqueando sesión")
	}

	s.log.WithContext(ctx).Info("session locked", "session_id", session.ID, "reason", reason)
	return nil
}

func (s *SessionService) CleanExpired(ctx context.Context) (int64, error) {
	count, err := s.sessions.CleanExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("SESSION_FAILED: error limpiando sesiones expiradas")
	}
	if count > 0 {
		s.log.WithContext(ctx).Info("expired sessions cleaned", "count", count)
	}
	return count, nil
}
