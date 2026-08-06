package masterdata

import (
	"time"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Currency is the currency catalog (ISO 4217). It is a global lookup;
// the same set of currencies is available to every tenant.
type Currency struct {
	Code          valueobjects.CurrencyCode
	Symbol        string
	Name          string
	DecimalPlaces int
	Type          enums.CurrencyType
	IsActive      bool
	CreatedAt     time.Time
}

// NewCurrency validates the inputs and constructs a Currency.
func NewCurrency(now time.Time, code valueobjects.CurrencyCode, symbol, name string, decimalPlaces int, kind enums.CurrencyType) (*Currency, error) {
	if !kind.Valid() {
		return nil, errors.Wrap(errors.ErrInvalidEnum, errField("currency type is invalid"))
	}
	if decimalPlaces < 0 || decimalPlaces > 6 {
		return nil, errors.Wrap(errors.ErrOutOfRange, errField("decimal places must be in 0..6"))
	}
	return &Currency{
		Code:          code,
		Symbol:        symbol,
		Name:          name,
		DecimalPlaces: decimalPlaces,
		Type:          kind,
		IsActive:      true,
		CreatedAt:     now,
	}, nil
}

// Activate / Deactivate.
func (c *Currency) Activate()   { c.IsActive = true }
func (c *Currency) Deactivate() { c.IsActive = false }
