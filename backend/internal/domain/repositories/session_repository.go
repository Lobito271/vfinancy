package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/identity"
)

type SessionRepository interface {
	Create(ctx context.Context, session *identity.UserSession) error
	GetByToken(ctx context.Context, token string) (*identity.UserSession, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*identity.UserSession, error)
	Update(ctx context.Context, session *identity.UserSession) error
	DeactivateAll(ctx context.Context, userID uuid.UUID) error
	Deactivate(ctx context.Context, sessionID uuid.UUID) error
	CleanExpired(ctx context.Context) (int64, error)
}
