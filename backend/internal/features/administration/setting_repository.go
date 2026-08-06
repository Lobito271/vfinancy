package administration

import (
	"context"

	"github.com/google/uuid"

)

type SettingRepository interface {
	Upsert(ctx context.Context, setting *ApplicationSetting) error
	GetByKey(ctx context.Context, companyID uuid.UUID, key string) (*ApplicationSetting, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID) ([]*ApplicationSetting, error)
	ListByCategory(ctx context.Context, companyID uuid.UUID, category string) ([]*ApplicationSetting, error)
	Delete(ctx context.Context, companyID uuid.UUID, key string) error
}
