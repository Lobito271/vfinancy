package administration

import (
	"context"

)

type CountryRepository interface {
	GetByCode(ctx context.Context, code string) (*Country, error)
	List(ctx context.Context) ([]*Country, error)
	Upsert(ctx context.Context, country *Country) error
}
