package auth

import (
	"vfinancy/backend/internal/domain/repositories"
	"context"

	"github.com/google/uuid"

)

// UserFilter is the input to UserRepository.List. Any non-zero field
// is included in the WHERE clause.
type UserFilter struct {
	CompanyID       *uuid.UUID
	Username        string
	Email           string
	Status          string
	IncludeDeleted  bool
	repositories.PageRequest
}

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, companyID uuid.UUID, username string) (*User, error)
	GetByEmail(ctx context.Context, companyID uuid.UUID, email string) (*User, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, filter UserFilter) (repositories.Page[*User], error)
}
