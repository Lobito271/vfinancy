package masterdata

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Tax is the tax catalog. Examples: IGV 18%, ISR 29.5%, IVAP 4%, EXEMPT 0%.
// Rates are versioned separately (see valueobjects.Percentage / a future
// TaxRate entity) and applied per document line via a snapshot.
type Tax struct {
	ID           uuid.UUID
	Code         string
	Name         string
	ShortName    string
	CountryCode  string
	DefaultRate  valueobjects.Percentage
	IsInclusive  bool
	IsPercentage bool
	Category     enums.TaxCategory
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewTax validates and constructs a Tax catalog row.
func NewTax(now time.Time, code, name, shortName, countryCode string, defaultRate valueobjects.Percentage, inclusive, isPercentage bool, category enums.TaxCategory) (*Tax, error) {
	if !category.Valid() {
		return nil, errors.Wrap(errors.ErrInvalidEnum, errField("tax category is invalid"))
	}
	if code == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("tax code is required"))
	}
	if name == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("tax name is required"))
	}
	return &Tax{
		ID:           uuid.New(),
		Code:         code,
		Name:         name,
		ShortName:    shortName,
		CountryCode:  countryCode,
		DefaultRate:  defaultRate,
		IsInclusive:  inclusive,
		IsPercentage: isPercentage,
		Category:     category,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Activate / Deactivate.
func (t *Tax) Activate()   { t.IsActive = true }
func (t *Tax) Deactivate() { t.IsActive = false }

// ChangeDefaultRate updates the default rate. The application layer is
// expected to record the old rate in a history table for audit.
func (t *Tax) ChangeDefaultRate(rate valueobjects.Percentage) {
	t.DefaultRate = rate
}
