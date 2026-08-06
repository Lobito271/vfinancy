package repositories

import (
	"context"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/entities/masterdata"
)

type TaxRepository interface {
	Create(ctx context.Context, tax *masterdata.Tax) error
	Update(ctx context.Context, tax *masterdata.Tax) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Tax, error)
	GetByCode(ctx context.Context, companyID *uuid.UUID, code string) (*masterdata.Tax, error)
	List(ctx context.Context, companyID *uuid.UUID) ([]*masterdata.Tax, error)
}
