// Package identity contains the entities that model the company's
// identity, branches, users, roles and permissions. These are the
// root of the multi-tenant model: every business entity references a
// Company (directly or transitively) and may carry an optional Branch
// scope.
package identity

import (
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Company is the multi-tenant root. A single installation may host
// many companies; the application layer guarantees that a session
// only sees one company at a time.
type Company struct {
	ID                    uuid.UUID
	Code                  valueobjects.ShortCode
	LegalName             valueobjects.FullName
	TradeName             valueobjects.FullName
	TaxID                 string
	CountryCode           string
	FunctionalCurrency    valueobjects.CurrencyCode
	Timezone              string
	FiscalYearStartMonth  int
	IsActive              bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
	CreatedBy             *uuid.UUID
	UpdatedBy             *uuid.UUID
}

// NewCompanyOptions carries the inputs for NewCompany. All fields are
// required unless marked optional.
type NewCompanyOptions struct {
	Code                 valueobjects.ShortCode
	LegalName            valueobjects.FullName
	TradeName            valueobjects.FullName  // optional
	TaxID                string
	CountryCode          string
	FunctionalCurrency   valueobjects.CurrencyCode
	Timezone             string                  // optional, defaults applied
	FiscalYearStartMonth int                      // 1..12
}

// NewCompany validates inputs and returns a new Company. The ID is
// generated; created/updated timestamps are set to now.
func NewCompany(now time.Time, opts NewCompanyOptions) (*Company, error) {
	if opts.FiscalYearStartMonth < 1 || opts.FiscalYearStartMonth > 12 {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("fiscal year start month must be 1..12"))
	}
	if opts.CountryCode == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("country code is required"))
	}
	if opts.TaxID == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("tax id is required"))
	}
	if opts.Timezone == "" {
		opts.Timezone = "UTC"
	}
	return &Company{
		ID:                   uuid.New(),
		Code:                 opts.Code,
		LegalName:            opts.LegalName,
		TradeName:            opts.TradeName,
		TaxID:                opts.TaxID,
		CountryCode:          opts.CountryCode,
		FunctionalCurrency:   opts.FunctionalCurrency,
		Timezone:             opts.Timezone,
		FiscalYearStartMonth: opts.FiscalYearStartMonth,
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// Activate marks the company as active. Idempotent.
func (c *Company) Activate() { c.IsActive = true }

// Deactivate marks the company as inactive. New sessions cannot be
// opened for an inactive company (enforced by the application layer).
func (c *Company) Deactivate() { c.IsActive = false }

// ChangeFunctionalCurrency updates the company's functional currency.
// This is a structural change with downstream impact on every
// historical exchange rate snapshot; the application layer is
// responsible for asking the operator to confirm.
func (c *Company) ChangeFunctionalCurrency(code valueobjects.CurrencyCode) {
	c.FunctionalCurrency = code
}

// ChangeFiscalYearStart updates the start month of the fiscal year.
// Cannot be applied to a company with already-closed periods without
// further accounting work; the application layer is responsible for
// that gate.
func (c *Company) ChangeFiscalYearStart(month int) error {
	if month < 1 || month > 12 {
		return derrors.Wrap(derrors.ErrOutOfRange, errField("fiscal year start month must be 1..12"))
	}
	c.FiscalYearStartMonth = month
	return nil
}

// SoftDelete marks the company as deleted. The company is preserved
// for historical reporting. The application layer must verify that no
// active users / branches / sales exist before allowing this.
func (c *Company) SoftDelete(at time.Time, by uuid.UUID) {
	now := at
	c.DeletedAt = &now
	c.UpdatedAt = at
	c.UpdatedBy = &by
	c.IsActive = false
}

// IsDeleted reports whether the company has been soft-deleted.
func (c *Company) IsDeleted() bool { return c.DeletedAt != nil }

// errField wraps a context message around the derrors helpers.
func errField(s string) error { return errs{s} }

type errs struct{ s string }

func (e errs) Error() string { return e.s }
