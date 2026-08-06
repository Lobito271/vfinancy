package repositories

import (
	"context"

	"vfinancy/backend/internal/domain/entities/masterdata"
)

type CountryRepository interface {
	GetByCode(ctx context.Context, code string) (*masterdata.Country, error)
	List(ctx context.Context) ([]*masterdata.Country, error)
	Upsert(ctx context.Context, country *masterdata.Country) error
}
