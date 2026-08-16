package accounting

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
)

// FiscalPeriod gates journal posting: every journal entry belongs to
// an open fiscal period. Periods progress open → closing → closed.
type FiscalPeriod struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Name        string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Status      string
	ClosedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// NewFiscalPeriodOptions is the input to NewFiscalPeriod.
type NewFiscalPeriodOptions struct {
	CompanyID   uuid.UUID
	Name        string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Status      string
}

// NewFiscalPeriod validates and constructs a fiscal period.
func NewFiscalPeriod(now time.Time, opts NewFiscalPeriodOptions) (*FiscalPeriod, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company is required"))
	}
	if opts.Name == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("name is required"))
	}
	if opts.PeriodStart.IsZero() || opts.PeriodEnd.IsZero() || opts.PeriodEnd.Before(opts.PeriodStart) {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("period range is invalid"))
	}
	switch opts.Status {
	case "open", "closing", "closed":
	default:
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("period status is invalid"))
	}
	return &FiscalPeriod{
		ID:          uuid.New(),
		CompanyID:   opts.CompanyID,
		Name:        opts.Name,
		PeriodStart: opts.PeriodStart,
		PeriodEnd:   opts.PeriodEnd,
		Status:      opts.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
