package auth

import (
	"context"

	"github.com/google/uuid"

)

type SessionRepository interface {
	Create(ctx context.Context, session *UserSession) error
	GetByToken(ctx context.Context, token string) (*UserSession, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*UserSession, error)
	Update(ctx context.Context, session *UserSession) error
	DeactivateAll(ctx context.Context, userID uuid.UUID) error
	Deactivate(ctx context.Context, sessionID uuid.UUID) error
	CleanExpired(ctx context.Context) (int64, error)
}
