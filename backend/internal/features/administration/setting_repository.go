package administration

import "context"

type SettingRepository interface {
	GetByKey(ctx context.Context, key string) (*ApplicationSetting, error)
	Upsert(ctx context.Context, s *ApplicationSetting) error
	List(ctx context.Context) ([]*ApplicationSetting, error)
}
