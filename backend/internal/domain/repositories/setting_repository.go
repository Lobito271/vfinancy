package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/administration"
)

type SettingRepository interface {
	Upsert(ctx context.Context, setting *administration.ApplicationSetting) error
	GetByKey(ctx context.Context, companyID uuid.UUID, key string) (*administration.ApplicationSetting, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID) ([]*administration.ApplicationSetting, error)
	ListByCategory(ctx context.Context, companyID uuid.UUID, category string) ([]*administration.ApplicationSetting, error)
	Delete(ctx context.Context, companyID uuid.UUID, key string) error
}
