package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/shared/logger"
)

type SessionService struct {
	sessions SessionRepository
	ttl      time.Duration
	log      *logger.Logger
}

func NewSessionService(sessions SessionRepository, ttl time.Duration, log *logger.Logger) *SessionService {
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

func (s *SessionService) Create(ctx context.Context, userID uuid.UUID, ipAddress, userAgent, device string) (*UserSession, error) {
	tokenBytes := make([]byte, 64)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, ErrSessionFailed
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	session, err := NewUserSession(userID, token, s.ttl, ipAddress, userAgent, device)
	if err != nil {
		return nil, derrors.Wrap(ErrSessionFailed, err)
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, ErrSessionFailed
	}

	s.log.WithContext(ctx).Info("session created", "session_id", session.ID, "user_id", userID)
	return session, nil
}

func (s *SessionService) Validate(ctx context.Context, token string) (*UserSession, error) {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	now := time.Now().UTC()
	if !session.IsValid(now) {
		s.log.WithContext(ctx).Warn("invalid session", "session_id", session.ID, "active", session.IsActive, "expired", session.IsExpired(now), "locked", session.IsLocked())
		return nil, ErrSessionInvalid
	}

	session.RecordActivity(now)
	if err := s.sessions.Update(ctx, session); err != nil {
		s.log.WithContext(ctx).Error("failed to record session activity", "session_id", session.ID, "error", err)
	}

	return session, nil
}

func (s *SessionService) Destroy(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessions.Deactivate(ctx, sessionID); err != nil {
		return ErrSessionFailed
	}
	s.log.WithContext(ctx).Info("session destroyed", "session_id", sessionID)
	return nil
}

func (s *SessionService) DestroyAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessions.DeactivateAll(ctx, userID); err != nil {
		return ErrSessionFailed
	}
	s.log.WithContext(ctx).Info("all sessions destroyed", "user_id", userID)
	return nil
}

func (s *SessionService) Lock(ctx context.Context, token string, reason string) error {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return ErrSessionNotFound
	}

	now := time.Now().UTC()
	session.Lock(reason, now)
	if err := s.sessions.Update(ctx, session); err != nil {
		return ErrSessionFailed
	}

	s.log.WithContext(ctx).Info("session locked", "session_id", session.ID, "reason", reason)
	return nil
}

func (s *SessionService) CleanExpired(ctx context.Context) (int64, error) {
	count, err := s.sessions.CleanExpired(ctx)
	if err != nil {
		return 0, ErrSessionFailed
	}
	if count > 0 {
		s.log.WithContext(ctx).Info("expired sessions cleaned", "count", count)
	}
	return count, nil
}
