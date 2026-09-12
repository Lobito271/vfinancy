// Package customer implements the business logic for the customer
// aggregate: creation, validation, lifecycle, debt ledger.
package customer

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/shared/apperrors"
)

// CustomerService is the entry point for all customer-related
// business operations. It owns no state of its own; every method is
// a self-contained transaction or single query.
type CustomerService struct {
	repo CustomerRepository
	txm  repositories.TransactionManager
	log  *logger.Logger
}

// NewService returns a CustomerService ready for use.
func NewService(repo CustomerRepository, txm repositories.TransactionManager, log *logger.Logger) *CustomerService {
	return &CustomerService{repo: repo, txm: txm, log: log}
}

// CreateInput is the payload for Create. DocType is "" (no document),
// "DNI" or "RUC"; Email, Phone and Address are optional.
type CreateInput struct {
	BusinessName string
	DocType      string
	DocNumber    string
	Email        string
	Phone        string
	Address      string
}

// parseDocType converts a raw string to a DocumentType, rejecting
// unknown values.
func parseDocType(s string) (enums.DocumentType, error) {
	dt := enums.DocumentType(s)
	if !dt.Valid() {
		return enums.TypeNone, derrors.Wrap(derrors.ErrInvalidEnum, errField("document type is invalid: "+s))
	}
	return dt, nil
}

// optional builds an optional value object: empty input yields the
// zero value, anything else is validated.
func optional[T any](s string, build func(string) (T, error)) (T, error) {
	var zero T
	if s == "" {
		return zero, nil
	}
	return build(s)
}

// conflict is the user-facing error for a duplicated document pair.
func conflict() error {
	return apperrors.Errorf(apperrors.ErrConflict, "ya existe un cliente con ese tipo y número de documento")
}

// Create validates the input, constructs a Customer and persists it
// inside a transaction. A duplicated document pair is rejected.
func (s *CustomerService) Create(ctx context.Context, in CreateInput) (*Customer, error) {
	docType, err := parseDocType(in.DocType)
	if err != nil {
		return nil, err
	}
	if err := valueobjects.ValidateDocument(docType, in.DocNumber); err != nil {
		return nil, err
	}
	email, err := optional(in.Email, valueobjects.NewEmail)
	if err != nil {
		return nil, err
	}
	phone, err := optional(in.Phone, valueobjects.NewPhone)
	if err != nil {
		return nil, err
	}
	addr, err := optional(in.Address, func(s string) (valueobjects.Address, error) {
		return valueobjects.NewAddress(s)
	})
	if err != nil {
		return nil, err
	}
	c, err := NewCustomer(in.BusinessName, docType, in.DocNumber)
	if err != nil {
		return nil, err
	}
	c.Email = email
	c.Phone = phone
	c.Address = addr

	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		if in.DocNumber != "" {
			_, err := s.repo.GetByDocument(ctx, in.DocType, in.DocNumber)
			if err == nil {
				return conflict()
			}
			if !errors.Is(err, repositories.ErrNotFound) {
				return err
			}
		}
		if err := s.repo.Create(ctx, c); err != nil {
			if errors.Is(err, repositories.ErrDuplicate) {
				return conflict()
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("customer created", "customer_id", c.ID)
	return c, nil
}

// UpdateInput is the payload for Update. A non-empty BusinessName is
// applied; DocType "" clears the document pair (empty DocNumber with a
// type is rejected by validation); Email, Phone and Address are
// cleared when empty and validated otherwise; Status "" keeps the
// current value, anything else must be a valid CustomerStatus.
type UpdateInput struct {
	ID           uuid.UUID
	BusinessName string
	DocType      string
	DocNumber    string
	Email        string
	Phone        string
	Address      string
	Status       string
}

// Update loads the customer, applies the requested changes and
// persists the entity inside a transaction.
func (s *CustomerService) Update(ctx context.Context, in UpdateInput) (*Customer, error) {
	if in.ID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("customer id is required"))
	}
	var out *Customer
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.repo.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if in.BusinessName != "" {
			name, err := valueobjects.NewFullName(in.BusinessName)
			if err != nil {
				return err
			}
			c.BusinessName = name
		}
		docType, err := parseDocType(in.DocType)
		if err != nil {
			return err
		}
		doc, err := valueobjects.NewDocumentNumber(docType, in.DocNumber)
		if err != nil {
			return err
		}
		c.DocumentType = docType
		c.DocumentNumber = doc
		email, err := optional(in.Email, valueobjects.NewEmail)
		if err != nil {
			return err
		}
		c.Email = email
		phone, err := optional(in.Phone, valueobjects.NewPhone)
		if err != nil {
			return err
		}
		c.Phone = phone
		addr, err := optional(in.Address, func(s string) (valueobjects.Address, error) {
			return valueobjects.NewAddress(s)
		})
		if err != nil {
			return err
		}
		c.Address = addr
		if in.Status != "" {
			st := enums.CustomerStatus(in.Status)
			if !st.Valid() {
				return derrors.Wrap(derrors.ErrInvalidEnum, errField("customer status is invalid: "+in.Status))
			}
			c.Status = st
		}
		c.Touch()
		if err := s.repo.Update(ctx, c); err != nil {
			if errors.Is(err, repositories.ErrDuplicate) {
				return conflict()
			}
			return err
		}
		out = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("customer updated", "customer_id", out.ID)
	return out, nil
}

// Delete soft-deletes the customer. Historical references are
// preserved.
func (s *CustomerService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.log.Info("customer deleted", "customer_id", id)
	return nil
}

// GetByID returns a single customer.
func (s *CustomerService) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByDocument looks up a customer by document type + number.
func (s *CustomerService) GetByDocument(ctx context.Context, docType, docNumber string) (*Customer, error) {
	return s.repo.GetByDocument(ctx, docType, docNumber)
}

// List returns customers matching the filter.
func (s *CustomerService) List(ctx context.Context, filter CustomerFilter) (repositories.Page[*Customer], error) {
	return s.repo.List(ctx, filter)
}

// Options returns active customers for the sale form selects.
func (s *CustomerService) Options(ctx context.Context) ([]*Customer, error) {
	page, err := s.repo.List(ctx, CustomerFilter{
		Status:      string(enums.CustomerStatusActive),
		PageRequest: repositories.PageRequest{Limit: 1000},
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// mutateDebt applies a signed delta to the customer's debt, with a
// floor of zero, and persists it in a transaction.
func (s *CustomerService) mutateDebt(ctx context.Context, id uuid.UUID, delta valueobjects.Money) (valueobjects.Money, error) {
	var out valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		debt := c.CurrentDebt.Add(delta)
		if debt.IsNegative() {
			debt = valueobjects.Zero()
		}
		c.CurrentDebt = debt
		c.Touch()
		if err := s.repo.Update(ctx, c); err != nil {
			return err
		}
		out = debt
		return nil
	})
	if err != nil {
		return valueobjects.Money{}, err
	}
	return out, nil
}

// RecordSale adds a sale amount to the customer's debt and returns the
// new balance. Called when a sale is finalized.
func (s *CustomerService) RecordSale(ctx context.Context, id uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	return s.mutateDebt(ctx, id, amount)
}

// RecordPayment reduces the customer's debt by a payment amount, with
// a floor of zero, and returns the new balance.
func (s *CustomerService) RecordPayment(ctx context.Context, id uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	return s.mutateDebt(ctx, id, amount.Neg())
}

// AdjustDebt applies a signed delta to the customer's debt, with a
// floor of zero, and returns the new balance. Used by sale
// cancellation.
func (s *CustomerService) AdjustDebt(ctx context.Context, id uuid.UUID, delta valueobjects.Money) (valueobjects.Money, error) {
	return s.mutateDebt(ctx, id, delta)
}

// OutstandingBalance returns the customer's stored current debt.
func (s *CustomerService) OutstandingBalance(ctx context.Context, id uuid.UUID) (valueobjects.Money, error) {
	raw, err := s.repo.GetOutstandingBalance(ctx, id)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return valueobjects.MoneyFromString(raw)
}
