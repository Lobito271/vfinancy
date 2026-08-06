package administration

import (
	"context"
)

// CurrencyRepository persists the global currency catalog.
type CurrencyRepository interface {
	GetByCode(ctx context.Context, code string) (*Currency, error)
	List(ctx context.Context) ([]*Currency, error)
	Upsert(ctx context.Context, c *Currency) error
}
