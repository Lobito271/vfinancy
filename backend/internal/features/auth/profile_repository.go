package auth

import (
	"context"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *UserProfile) error
	Update(ctx context.Context, profile *UserProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
}
