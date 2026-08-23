package workspace

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetProfile(ctx context.Context) (*LocalProfile, error)
	CreateProfile(ctx context.Context, profile *LocalProfile) error
	UpdateProfile(ctx context.Context, profile *LocalProfile) error
	ListCompanies(ctx context.Context) ([]*Company, error)
	GetCompany(ctx context.Context, id uuid.UUID) (*Company, error)
	CreateCompany(ctx context.Context, company *Company) error
	UpdateCompany(ctx context.Context, company *Company) error
}
