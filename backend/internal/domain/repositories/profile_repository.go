package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/identity"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *identity.UserProfile) error
	Update(ctx context.Context, profile *identity.UserProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*identity.UserProfile, error)
}
