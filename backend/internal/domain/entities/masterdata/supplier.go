package masterdata

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Supplier is a vendor of goods or services. Suppliers carry debt
// (what we owe them) and payment terms.
type Supplier struct {
	ID              uuid.UUID
	CompanyID       uuid.UUID
	Document        valueobjects.DocumentNumber
	BusinessName    valueobjects.FullName
	TradeName       valueobjects.FullName
	TaxID            string
	IsInternational bool
	DefaultCurrency valueobjects.CurrencyCode
	CurrentDebt     valueobjects.Money
	PaymentTermDays int
	Status          enums.SupplierStatus
	Email           valueobjects.Email
	Phone           valueobjects.Phone
	Address         valueobjects.Address
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	CreatedBy       *uuid.UUID
	UpdatedBy       *uuid.UUID
}

// NewSupplierOptions is the input to NewSupplier.
type NewSupplierOptions struct {
	CompanyID       uuid.UUID
	Document        valueobjects.DocumentNumber
	BusinessName    valueobjects.FullName
	TradeName       valueobjects.FullName
	TaxID            string
	IsInternational bool
	DefaultCurrency valueobjects.CurrencyCode
	PaymentTermDays int
	Email           valueobjects.Email
	Phone           valueobjects.Phone
	Address         valueobjects.Address
}

// NewSupplier validates inputs and constructs a Supplier.
func NewSupplier(now time.Time, opts NewSupplierOptions) (*Supplier, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if opts.PaymentTermDays < 0 {
		return nil, errors.Wrap(errors.ErrOutOfRange, errField("payment term days cannot be negative"))
	}
	return &Supplier{
		ID:              uuid.New(),
		CompanyID:       opts.CompanyID,
		Document:        opts.Document,
		BusinessName:    opts.BusinessName,
		TradeName:       opts.TradeName,
		TaxID:            opts.TaxID,
		IsInternational: opts.IsInternational,
		DefaultCurrency: opts.DefaultCurrency,
		CurrentDebt:     valueobjects.Zero(),
		PaymentTermDays: opts.PaymentTermDays,
		Status:          enums.SupplierStatusActive,
		Email:           opts.Email,
		Phone:           opts.Phone,
		Address:         opts.Address,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Activate / Deactivate.
func (s *Supplier) Activate()   { s.Status = enums.SupplierStatusActive }
func (s *Supplier) Deactivate() { s.Status = enums.SupplierStatusInactive }

// RecordPurchase adds a new purchase amount to the supplier's debt.
func (s *Supplier) RecordPurchase(amount valueobjects.Money) valueobjects.Money {
	s.CurrentDebt = s.CurrentDebt.Add(amount)
	return s.CurrentDebt
}

// RecordPayment reduces the supplier's debt. Returns the resulting balance.
func (s *Supplier) RecordPayment(amount valueobjects.Money) (valueobjects.Money, error) {
	if amount.IsNegative() || amount.IsZero() {
		return s.CurrentDebt, errors.Wrap(errors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	s.CurrentDebt = s.CurrentDebt.Sub(amount)
	if s.CurrentDebt.IsNegative() {
		s.CurrentDebt = valueobjects.Zero()
	}
	return s.CurrentDebt, nil
}

// CanPlacePurchase reports whether a new purchase can be placed. A
// purchase is allowed if the supplier is active.
func (s *Supplier) CanPlacePurchase(amount valueobjects.Money) error {
	if s.Status == enums.SupplierStatusInactive {
		return errors.Wrap(errors.ErrSupplierInactive, errField("supplier is inactive"))
	}
	if amount.IsNegative() || amount.IsZero() {
		return errors.Wrap(errors.ErrInvalidPayment, errField("purchase amount must be positive"))
	}
	return nil
}

// UpdateContact updates the email / phone / address.
func (s *Supplier) UpdateContact(email valueobjects.Email, phone valueobjects.Phone, address valueobjects.Address) {
	s.Email = email
	s.Phone = phone
	s.Address = address
}

// ChangePaymentTerms updates the net-N days.
func (s *Supplier) ChangePaymentTerms(days int) error {
	if days < 0 {
		return errors.Wrap(errors.ErrOutOfRange, errField("payment term days cannot be negative"))
	}
	s.PaymentTermDays = days
	return nil
}
