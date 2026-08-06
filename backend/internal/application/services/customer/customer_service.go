// Package customer implements the business logic for the customer
// aggregate: creation, validation, lifecycle, credit management.
package customer

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/masterdata"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CustomerService is the entry point for all customer-related
// business operations. It owns no state of its own; every method is
// a self-contained transaction or single query.
type CustomerService struct {
	repo repositories.CustomerRepository
	txm  services.TxManager
	log  *common.Logger
}

// New returns a CustomerService ready for use.
func New(repo repositories.CustomerRepository, txm services.TxManager, log *common.Logger) *CustomerService {
	if repo == nil {
		panic("customer: nil repo")
	}
	if txm == nil {
		panic("customer: nil txm")
	}
	if log == nil {
		panic("customer: nil logger")
	}
	return &CustomerService{repo: repo, txm: txm, log: log}
}

// CreateInput is the payload for CreateCustomer. DocumentType and
// DocumentNumber are required; the rest has sensible defaults.
type CreateInput struct {
	CompanyID        uuid.UUID
	DocumentType     enums.DocumentType
	DocumentNumber   string
	BusinessName     string
	TradeName        string  // optional
	TaxCategory       enums.TaxCategory
	CreditLimit      valueobjects.Money
	PaymentTermDays  int
	Email            string
	Phone            string
	Address          string
	BranchID         *uuid.UUID
}

// CreateInput also has a "now" for testability.
func (in CreateInput) now() time.Time { return time.Now().UTC() }

// CreateCustomer validates the input, constructs a Customer entity
// via the domain constructor, and persists it. The whole operation
// runs inside a transaction.
func (s *CustomerService) Create(ctx context.Context, in CreateInput) (*masterdata.Customer, error) {
	if in.CompanyID == uuid.Nil {
		return nil, services.EnsureError("REQUIRED", "company id is required")
	}
	doc, err := valueobjects.NewDocumentNumber(in.DocumentType, in.DocumentNumber)
	if err != nil {
		return nil, err
	}
	email, err := valueobjects.NewEmail(in.Email)
	if err != nil {
		return nil, err
	}
	addr, err := valueobjects.NewAddress(in.Address)
	if err != nil {
		return nil, err
	}
	var phone valueobjects.Phone
	if in.Phone != "" {
		p, err := valueobjects.NewPhone(in.Phone)
		if err != nil {
			return nil, err
		}
		phone = p
	}
	var tradeName valueobjects.FullName
	if in.TradeName != "" {
		tn, err := valueobjects.NewFullName(in.TradeName)
		if err != nil {
			return nil, err
		}
		tradeName = tn
	}

	var out *masterdata.Customer
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := masterdata.NewCustomer(in.now(), masterdata.NewCustomerOptions{
			CompanyID:       in.CompanyID,
			Document:        doc,
			BusinessName:    valueobjects.MustFullName(in.BusinessName),
			TradeName:       tradeName,
			TaxCategory:     in.TaxCategory,
			CreditLimit:     in.CreditLimit,
			PaymentTermDays: in.PaymentTermDays,
			Email:           email,
			Phone:           phone,
			Address:         addr,
			BranchID:        in.BranchID,
		})
		if err != nil {
			return err
		}
		if err := uow.Customers().Create(ctx, c); err != nil {
			return err
		}
		out = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("customer created",
		"customer_id", out.ID,
		"company_id", out.CompanyID,
		"document_number", out.Document.Number(),
	)
	return out, nil
}

// UpdateInput is the payload for UpdateCustomer. All fields except
// ID are optional — only the non-zero ones are applied.
type UpdateInput struct {
	ID             uuid.UUID
	BusinessName   string  // optional
	TradeName      string  // optional, "" means clear
	TaxCategory     enums.TaxCategory
	CreditLimit    *valueobjects.Money  // nil = unchanged
	PaymentTermDays *int                 // nil = unchanged
	Email          string
	Phone          string
	Address        string
}

// UpdateCustomer loads the customer, applies the requested changes,
// persists the updated entity. The customer must be active.
func (s *CustomerService) Update(ctx context.Context, in UpdateInput) (*masterdata.Customer, error) {
	if in.ID == uuid.Nil {
		return nil, services.EnsureError("REQUIRED", "customer id is required")
	}
	var out *masterdata.Customer
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := uow.Customers().GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if c.Status != enums.CustomerStatusActive {
			return services.ErrCustomerBlocked
		}
		if in.BusinessName != "" {
			name, err := valueobjects.NewFullName(in.BusinessName)
			if err != nil {
				return err
			}
			c.BusinessName = name
		}
		if in.TradeName == "" {
			c.TradeName = valueobjects.FullName{}
		} else {
			tn, err := valueobjects.NewFullName(in.TradeName)
			if err != nil {
				return err
			}
			c.TradeName = tn
		}
		if in.TaxCategory != "" {
			c.TaxCategory = in.TaxCategory
		}
		if in.CreditLimit != nil {
			if err := c.UpdateCreditLimit(*in.CreditLimit); err != nil {
				return err
			}
		}
		if in.PaymentTermDays != nil {
			if err := c.ChangePaymentTerms(*in.PaymentTermDays); err != nil {
				return err
			}
		}
		if in.Email != "" {
			email, err := valueobjects.NewEmail(in.Email)
			if err != nil {
				return err
			}
			c.Email = email
		}
		if in.Phone != "" {
			p, err := valueobjects.NewPhone(in.Phone)
			if err != nil {
				return err
			}
			c.Phone = p
		}
		if in.Address != "" {
			addr, err := valueobjects.NewAddress(in.Address)
			if err != nil {
				return err
			}
			c.Address = addr
		}
		if err := uow.Customers().Update(ctx, c); err != nil {
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

// Deactivate marks the customer as inactive. Already-inactive or
// already-deleted customers are no-ops.
func (s *CustomerService) Deactivate(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := uow.Customers().GetByID(ctx, id)
		if err != nil {
			return err
		}
		c.Deactivate()
		return uow.Customers().Update(ctx, c)
	})
	if err != nil {
		return err
	}
	s.log.Info("customer deactivated", "customer_id", id)
	return nil
}

// Block marks the customer as blocked with a reason. The reason is
// mandatory and free-form (typically a payment-related note).
func (s *CustomerService) Block(ctx context.Context, id uuid.UUID, reason string) error {
	if reason == "" {
		return services.EnsureError("REQUIRED", "block reason is required")
	}
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := uow.Customers().GetByID(ctx, id)
		if err != nil {
			return err
		}
		c.Block(reason)
		return uow.Customers().Update(ctx, c)
	})
	if err != nil {
		return err
	}
	s.log.Info("customer blocked", "customer_id", id, "reason", reason)
	return nil
}

// Unblock clears the blocked state, returning the customer to active.
func (s *CustomerService) Unblock(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := uow.Customers().GetByID(ctx, id)
		if err != nil {
			return err
		}
		c.Unblock()
		return uow.Customers().Update(ctx, c)
	})
	if err != nil {
		return err
	}
	s.log.Info("customer unblocked", "customer_id", id)
	return nil
}

// UpdateCreditLimit changes the customer's credit limit. The new limit
// must be non-negative. This is a thin wrapper around the entity
// method that loads + validates + saves the customer.
func (s *CustomerService) UpdateCreditLimit(ctx context.Context, id uuid.UUID, limit valueobjects.Money) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := uow.Customers().GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := c.UpdateCreditLimit(limit); err != nil {
			return err
		}
		return uow.Customers().Update(ctx, c)
	})
	if err != nil {
		return err
	}
	s.log.Info("customer credit limit updated", "customer_id", id, "new_limit", limit)
	return nil
}

// RecordSale adds a sale amount to the customer's current debt. Called
// by the sales workflow when a sale is finalized.
func (s *CustomerService) RecordSale(ctx context.Context, id uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	var balance valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := uow.Customers().GetByID(ctx, id)
		if err != nil {
			return err
		}
		balance = c.RecordSale(amount)
		return uow.Customers().Update(ctx, c)
	})
	if err != nil {
		return valueobjects.Money{}, err
	}
	return balance, nil
}

// RecordPayment reduces the customer's debt by a payment. Used by the
// payments workflow.
func (s *CustomerService) RecordPayment(ctx context.Context, id uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	var balance valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		c, err := uow.Customers().GetByID(ctx, id)
		if err != nil {
			return err
		}
		balance, err = c.RecordPayment(amount)
		if err != nil {
			return err
		}
		return uow.Customers().Update(ctx, c)
	})
	if err != nil {
		return valueobjects.Money{}, err
	}
	return balance, nil
}

// OutstandingBalance returns the customer's current debt. It is a
// convenience over the repo that does not require a transaction.
func (s *CustomerService) OutstandingBalance(ctx context.Context, id uuid.UUID) (valueobjects.Money, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return c.CurrentDebt, nil
}

// AvailableCredit returns credit_limit - current_debt, clamped to zero.
// Used by sales UI to decide whether a new sale is allowed.
func (s *CustomerService) AvailableCredit(ctx context.Context, id uuid.UUID) (valueobjects.Money, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return c.AvailableCredit(), nil
}

// IsOverLimit returns whether current_debt exceeds credit_limit.
func (s *CustomerService) IsOverLimit(ctx context.Context, id uuid.UUID) (bool, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return c.IsOverLimit(), nil
}

// CanPlaceSale is a guard: returns nil if a new sale of the given
// amount is allowed, or a domain error if not (blocked customer,
// inactive, would exceed credit limit).
func (s *CustomerService) CanPlaceSale(ctx context.Context, id uuid.UUID, amount valueobjects.Money) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return c.CanPlaceSale(amount)
}

// GetByID is a convenience that returns the customer aggregate.
func (s *CustomerService) GetByID(ctx context.Context, id uuid.UUID) (*masterdata.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByDocument looks up a customer by document type + number.
func (s *CustomerService) GetByDocument(ctx context.Context, companyID uuid.UUID, docType enums.DocumentType, docNum string) (*masterdata.Customer, error) {
	return s.repo.GetByDocument(ctx, companyID, string(docType), docNum)
}
