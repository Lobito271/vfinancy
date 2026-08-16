package accounting

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// FiscalPeriodFilter is the input to FiscalPeriodRepository.List.
type FiscalPeriodFilter struct {
	CompanyID *uuid.UUID
	Status    string
	repositories.PageRequest
}

// FiscalPeriodRepository persists fiscal periods and resolves the open
// period that covers a given date.
type FiscalPeriodRepository interface {
	Create(ctx context.Context, p *FiscalPeriod) error
	GetByID(ctx context.Context, id uuid.UUID) (*FiscalPeriod, error)
	GetOpenForDate(ctx context.Context, companyID uuid.UUID, date time.Time) (*FiscalPeriod, error)
	List(ctx context.Context, filter FiscalPeriodFilter) (repositories.Page[*FiscalPeriod], error)
}
