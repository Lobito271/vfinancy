package identity

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/errors"
)

type UserSession struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Token          string
	IPAddress      string
	UserAgent      string
	Device         string
	IsActive       bool
	LockedAt       *time.Time
	LockedReason   string
	ExpiresAt      time.Time
	LastActivityAt time.Time
	CreatedAt      time.Time
}

func NewUserSession(userID uuid.UUID, token string, ttl time.Duration, ipAddress, userAgent, device string) (*UserSession, error) {
	if userID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("user id is required"))
	}
	if token == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("token is required"))
	}
	if ttl <= 0 {
		return nil, errors.Wrap(errors.ErrOutOfRange, errField("ttl must be greater than zero"))
	}
	now := time.Now()
	return &UserSession{
		ID:             uuid.New(),
		UserID:         userID,
		Token:          token,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		Device:         device,
		IsActive:       true,
		ExpiresAt:      now.Add(ttl),
		LastActivityAt: now,
		CreatedAt:      now,
	}, nil
}

func (s *UserSession) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *UserSession) IsLocked() bool {
	return s.LockedAt != nil
}

func (s *UserSession) IsValid(now time.Time) bool {
	return s.IsActive && !s.IsExpired(now) && !s.IsLocked()
}

func (s *UserSession) RecordActivity(now time.Time) {
	s.LastActivityAt = now
}

func (s *UserSession) Lock(reason string, now time.Time) {
	s.LockedAt = &now
	s.LockedReason = reason
}

func (s *UserSession) Unlock() {
	s.LockedAt = nil
	s.LockedReason = ""
}

func (s *UserSession) Deactivate() {
	s.IsActive = false
}
