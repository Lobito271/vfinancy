// Package supplier implements the business logic for the supplier
// aggregate: lifecycle, payment recording, outstanding balance.
package supplier

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/shared/apperrors"
	"vfinancy/backend/internal/shared/logger"
)

// SupplierService owns the supplier slice.
type SupplierService struct {
	repo SupplierRepository
	txm  repositories.TransactionManager
	log  *logger.Logger
}

// New returns a SupplierService ready for use.
func New(repo SupplierRepository, txm repositories.TransactionManager, log *logger.Logger) *SupplierService {
	return &SupplierService{repo: repo, txm: txm, log: log}
}

// CreateInput is the payload for CreateSupplier. DocumentType +
// DocumentNumber are required.
type CreateInput struct {
	CompanyID       uuid.UUID
	DocumentType    enums.DocumentType
	DocumentNumber  string
	BusinessName    string
	TradeName       string
	TaxID           string
	IsInternational bool
	DefaultCurrency valueobjects.CurrencyCode
	PaymentTermDays int
	Email           string
	Phone           string
	Address         string
}

// CreateSupplier validates the input and persists the new supplier.
func (s *SupplierService) Create(ctx context.Context, in CreateInput) (*Supplier, error) {
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
	tradeName, err := valueobjects.NewFullName(in.TradeName)
	if err != nil {
		return nil, err
	}

	var out *Supplier
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sup, err := NewSupplier(time.Now().UTC(), NewSupplierOptions{
			CompanyID:       in.CompanyID,
			Document:        doc,
			BusinessName:    valueobjects.MustFullName(in.BusinessName),
			TradeName:       tradeName,
			TaxID:           in.TaxID,
			IsInternational: in.IsInternational,
			DefaultCurrency: in.DefaultCurrency,
			PaymentTermDays: in.PaymentTermDays,
			Email:           email,
			Phone:           phone,
			Address:         addr,
		})
		if err != nil {
			return err
		}
		if err := s.repo.Create(ctx, sup); err != nil {
			return err
		}
		out = sup
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("supplier created", "supplier_id", out.ID, "company_id", out.CompanyID)
	return out, nil
}

// UpdateInput is the payload for UpdateSupplier. Only the non-zero
// fields are applied.
type UpdateInput struct {
	ID              uuid.UUID
	BusinessName    string
	TradeName       string
	TaxID           string
	PaymentTermDays *int
	Email           string
	Phone           string
	Address         string
}

// UpdateSupplier applies the requested changes. The supplier must be
// active; deactivated suppliers cannot be modified.
func (s *SupplierService) Update(ctx context.Context, in UpdateInput) (*Supplier, error) {
	var out *Supplier
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sup, err := s.repo.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if sup.Status == enums.SupplierStatusInactive {
			return apperrors.ErrSupplierBlocked
		}
		if in.BusinessName != "" {
			sup.BusinessName = valueobjects.MustFullName(in.BusinessName)
		}
		if in.TradeName == "" {
			sup.TradeName = valueobjects.FullName{}
		} else {
			sup.TradeName = valueobjects.MustFullName(in.TradeName)
		}
		if in.TaxID != "" {
			sup.TaxID = in.TaxID
		}
		if in.PaymentTermDays != nil {
			if err := sup.ChangePaymentTerms(*in.PaymentTermDays); err != nil {
				return err
			}
		}
		if in.Email != "" {
			email, err := valueobjects.NewEmail(in.Email)
			if err != nil {
				return err
			}
			sup.Email = email
		}
		if in.Phone != "" {
			p, err := valueobjects.NewPhone(in.Phone)
			if err != nil {
				return err
			}
			sup.Phone = p
		}
		if in.Address != "" {
			addr, err := valueobjects.NewAddress(in.Address)
			if err != nil {
				return err
			}
			sup.Address = addr
		}
		if err := s.repo.Update(ctx, sup); err != nil {
			return err
		}
		out = sup
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("supplier updated", "supplier_id", out.ID)
	return out, nil
}

// Deactivate marks the supplier as inactive. The supplier can no
// longer be the target of new purchases.
func (s *SupplierService) Deactivate(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sup, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		sup.Deactivate()
		return s.repo.Update(ctx, sup)
	})
	if err != nil {
		return err
	}
	s.log.Info("supplier deactivated", "supplier_id", id)
	return nil
}

// RecordPurchase adds a purchase amount to the supplier's debt.
// Called by the purchasing slice.
func (s *SupplierService) RecordPurchase(ctx context.Context, id uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	var balance valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sup, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		balance = sup.RecordPurchase(amount)
		return s.repo.Update(ctx, sup)
	})
	if err != nil {
		return valueobjects.Money{}, err
	}
	return balance, nil
}

// RecordPayment reduces the supplier's debt by a payment. Called by
// the supplier-payments slice.
func (s *SupplierService) RecordPayment(ctx context.Context, id uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	var balance valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sup, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		balance, err = sup.RecordPayment(amount)
		if err != nil {
			return err
		}
		return s.repo.Update(ctx, sup)
	})
	if err != nil {
		return valueobjects.Money{}, err
	}
	return balance, nil
}

// OutstandingBalance returns the supplier's current debt.
func (s *SupplierService) OutstandingBalance(ctx context.Context, id uuid.UUID) (valueobjects.Money, error) {
	sup, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return sup.CurrentDebt, nil
}

// CanPlacePurchase is the guard for the purchasing slice.
func (s *SupplierService) CanPlacePurchase(ctx context.Context, id uuid.UUID, amount valueobjects.Money) error {
	sup, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return sup.CanPlacePurchase(amount)
}

// GetByID returns the supplier aggregate.
func (s *SupplierService) GetByID(ctx context.Context, id uuid.UUID) (*Supplier, error) {
	return s.repo.GetByID(ctx, id)
}
