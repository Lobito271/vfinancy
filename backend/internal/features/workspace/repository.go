package workspace

import "context"

type Repository interface {
	GetProfile(ctx context.Context) (*LocalProfile, error)
	CreateProfile(ctx context.Context, p *LocalProfile) error
	UpdateProfile(ctx context.Context, p *LocalProfile) error
}
