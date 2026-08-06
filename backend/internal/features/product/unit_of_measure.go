package product

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/errors"
)

// UnitOfMeasure is a unit (kg, L, m, und, pack, etc.).
type UnitOfMeasure struct {
	ID              uuid.UUID
	CompanyID       uuid.UUID
	Code            string
	Name            string
	Symbol          string
	AllowsDecimals  bool
	CreatedAt       time.Time
}

// NewUnitOfMeasure validates the code/symbol and constructs a unit.
func NewUnitOfMeasure(now time.Time, companyID uuid.UUID, code, name, symbol string, allowsDecimals bool) (*UnitOfMeasure, error) {
	if companyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if code == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("code is required"))
	}
	if name == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("name is required"))
	}
	return &UnitOfMeasure{
		ID:             uuid.New(),
		CompanyID:      companyID,
		Code:           code,
		Name:           name,
		Symbol:         symbol,
		AllowsDecimals: allowsDecimals,
		CreatedAt:      now,
	}, nil
}
