// Package customer implements the business logic for the customer
// aggregate: creation, validation, lifecycle, debt ledger.
package customer

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Customer is a buyer of the business's products. Customers carry a
// running debt balance and an active/inactive status.
type Customer struct {
	ID             uuid.UUID
	DocumentType   enums.DocumentType
	DocumentNumber valueobjects.DocumentNumber
	BusinessName   valueobjects.FullName
	Email          valueobjects.Email
	Phone          valueobjects.Phone
	Address        valueobjects.Address
	CurrentDebt    valueobjects.Money
	Status         enums.CustomerStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	CreatedBy      string
	UpdatedBy      string
}

// NewCustomer validates inputs and constructs a Customer. The customer
// starts active with zero debt.
func NewCustomer(name string, docType enums.DocumentType, docNumber string) (*Customer, error) {
	businessName, err := valueobjects.NewFullName(name)
	if err != nil {
		return nil, err
	}
	if err := valueobjects.ValidateDocument(docType, docNumber); err != nil {
		return nil, err
	}
	doc, err := valueobjects.NewDocumentNumber(docType, docNumber)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Customer{
		ID:             uuid.New(),
		DocumentType:   docType,
		DocumentNumber: doc,
		BusinessName:   businessName,
		CurrentDebt:    valueobjects.Zero(),
		Status:         enums.CustomerStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Touch stamps UpdatedAt with the current UTC time.
func (c *Customer) Touch() { c.UpdatedAt = time.Now().UTC() }

// Validate checks the customer invariants: business name present,
// document pair consistent, status valid, debt non-negative.
func (c *Customer) Validate() error {
	if c.BusinessName.String() == "" {
		return derrors.Wrap(derrors.ErrRequired, errField("business name is required"))
	}
	if err := valueobjects.ValidateDocument(c.DocumentType, c.DocumentNumber.Number()); err != nil {
		return err
	}
	if !c.Status.Valid() {
		return derrors.Wrap(derrors.ErrInvalidEnum, errField("customer status is invalid"))
	}
	if c.CurrentDebt.IsNegative() {
		return derrors.Wrap(derrors.ErrNegativeMoney, errField("current debt cannot be negative"))
	}
	return nil
}

// IsActive reports whether the customer can take part in new operations.
func (c *Customer) IsActive() bool { return c.Status == enums.CustomerStatusActive }
