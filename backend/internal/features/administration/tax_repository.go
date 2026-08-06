package administration

import (
	"context"

	"github.com/google/uuid"

)

type TaxRepository interface {
	Create(ctx context.Context, tax *Tax) error
	Update(ctx context.Context, tax *Tax) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tax, error)
	GetByCode(ctx context.Context, companyID *uuid.UUID, code string) (*Tax, error)
	List(ctx context.Context, companyID *uuid.UUID) ([]*Tax, error)
}
