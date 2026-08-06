package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/identity"
)

// UserFilter is the input to UserRepository.List. Any non-zero field
// is included in the WHERE clause.
type UserFilter struct {
	CompanyID       *uuid.UUID
	Username        string
	Email           string
	Status          string
	IncludeDeleted  bool
	PageRequest
}

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, user *identity.User) error
	Update(ctx context.Context, user *identity.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error)
	GetByUsername(ctx context.Context, companyID uuid.UUID, username string) (*identity.User, error)
	GetByEmail(ctx context.Context, companyID uuid.UUID, email string) (*identity.User, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter UserFilter) (Page[*identity.User], error)
}
