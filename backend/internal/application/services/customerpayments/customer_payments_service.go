// Package customer_payments owns the customer-payment workflow: register
// a payment, allocate it to one or more sales, and surface the
// outstanding receivables balance. Sales operations are delegated to
// sales.SalesService so the payment flow stays consistent with the
// sale's own state machine.
package customerpayments

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/sales"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CustomerPaymentService handles customer payments and their
// allocations. The service does NOT depend on SalesService directly;
// the application use case calls this service and SalesService
// sequentially, sharing the same transaction.
type CustomerPaymentService struct {
	payments  repositories.CustomerPaymentRepository
	advances  repositories.CustomerAdvanceRepository
	sales     repositories.SalesRepository
	txm       services.TxManager
	log       *common.Logger
}

// New returns a CustomerPaymentService ready for use.
func New(
	payments repositories.CustomerPaymentRepository,
	advances repositories.CustomerAdvanceRepository,
	sales repositories.SalesRepository,
	txm services.TxManager,
	log *common.Logger,
) *CustomerPaymentService {
	return &CustomerPaymentService{
		payments: payments,
		advances: advances,
		sales:    sales,
		txm:      txm,
		log:      log,
	}
}

// PayInput is the payload for Register.
type PayInput struct {
	CompanyID    uuid.UUID
	CustomerID   uuid.UUID
	Number       string
	PaymentDate  valueobjects.Date
	Amount       valueobjects.Money
	CurrencyCode valueobjects.CurrencyCode
	ExchangeRate valueobjects.ExchangeRate
	Method       enums.PaymentMethod
	BankAccountID *uuid.UUID
	CashRegisterID *uuid.UUID
	Reference    string
	Notes        string
}

// Register creates a new customer payment. The payment's allocations
// are populated by the application use case via ApplyToSale; the
// service itself does NOT touch any sale aggregate.
func (s *CustomerPaymentService) Register(ctx context.Context, in PayInput) (*sales.CustomerPayment, error) {
	if in.CompanyID == uuid.Nil || in.CustomerID == uuid.Nil {
		return nil, services.EnsureError("REQUIRED", "company and customer are required")
	}
	if !in.Amount.IsPositive() {
		return nil, services.EnsureError("INVALID_PAYMENT", "payment amount must be positive")
	}
	var out *sales.CustomerPayment
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		cp, err := sales.NewCustomerPayment(time.Now().UTC(), sales.NewCustomerPaymentOptions{
			CompanyID:      in.CompanyID,
			CustomerID:     in.CustomerID,
			Number:         in.Number,
			PaymentDate:    in.PaymentDate,
			Amount:         in.Amount,
			CurrencyCode:   in.CurrencyCode,
			ExchangeRate:   in.ExchangeRate,
			Method:         in.Method,
			BankAccountID:  in.BankAccountID,
			CashRegisterID: in.CashRegisterID,
			Reference:      in.Reference,
			Notes:          in.Notes,
		})
		if err != nil {
			return err
		}
		if err := uow.CustomerPayments().Create(ctx, cp); err != nil {
			return err
		}
		out = cp
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("customer payment registered",
		"payment_id", out.ID,
		"customer_id", out.CustomerID,
		"amount", out.Amount,
	)
	return out, nil
}

// ApplyToSale allocates part of a payment to a specific sale. The sale
// is also updated to reflect the new paid amount, in the same
// transaction. This is the only "join point" between the payment and
// sale flows and is called by the application use case.
func (s *CustomerPaymentService) ApplyToSale(ctx context.Context, paymentID, saleID uuid.UUID, amount valueobjects.Money) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		cp, err := uow.CustomerPayments().GetByID(ctx, paymentID)
		if err != nil {
			return err
		}
		if err := cp.ApplyToSale(saleID, amount); err != nil {
			return err
		}
		if err := uow.CustomerPayments().Update(ctx, cp); err != nil {
			return err
		}
		sale, err := uow.Sales().GetByID(ctx, saleID)
		if err != nil {
			return err
		}
		if _, err := sale.ApplyPayment(amount); err != nil {
			return err
		}
		if err := uow.Sales().Update(ctx, sale); err != nil {
			return err
		}
		return nil
	})
}

// AdvanceInput is the payload for RegisterAdvance.
type AdvanceInput struct {
	CompanyID    uuid.UUID
	CustomerID   uuid.UUID
	Number       string
	AdvanceDate  valueobjects.Date
	Amount       valueobjects.Money
	CurrencyCode valueobjects.CurrencyCode
	ExchangeRate valueobjects.ExchangeRate
	Method       string
	BankAccountID *uuid.UUID
	Notes        string
}

// RegisterAdvance creates a customer advance. The amount is not
// applied to any sale yet; that happens via ApplyToSale on the
// returned advance.
func (s *CustomerPaymentService) RegisterAdvance(ctx context.Context, in AdvanceInput) (*sales.CustomerAdvance, error) {
	if in.CompanyID == uuid.Nil || in.CustomerID == uuid.Nil {
		return nil, services.EnsureError("REQUIRED", "company and customer are required")
	}
	if !in.Amount.IsPositive() {
		return nil, services.EnsureError("INVALID_ADVANCE", "advance amount must be positive")
	}
	var out *sales.CustomerAdvance
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		a, err := sales.NewCustomerAdvance(time.Now().UTC(), sales.NewCustomerAdvanceOptions{
			CompanyID:      in.CompanyID,
			CustomerID:     in.CustomerID,
			Number:         in.Number,
			AdvanceDate:    in.AdvanceDate,
			Amount:         in.Amount,
			CurrencyCode:   in.CurrencyCode,
			ExchangeRate:   in.ExchangeRate,
			Method:         in.Method,
			BankAccountID:  in.BankAccountID,
			Notes:          in.Notes,
		})
		if err != nil {
			return err
		}
		if err := uow.CustomerAdvances().Create(ctx, a); err != nil {
			return err
		}
		out = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("customer advance registered",
		"advance_id", out.ID,
		"customer_id", out.CustomerID,
		"amount", out.Amount,
	)
	return out, nil
}

// ApplyAdvanceToSale applies part of an advance to a sale. The sale's
// paid amount is updated, and the advance's remaining balance is
// reduced. The use case layer is expected to call this and then
// SalesService.ApplyPayment / MarkAsPaid in the same transaction.
func (s *CustomerPaymentService) ApplyAdvanceToSale(ctx context.Context, advanceID, saleID uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	var remaining valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		a, err := uow.CustomerAdvances().GetByID(ctx, advanceID)
		if err != nil {
			return err
		}
		remaining, err = a.ApplyToSale(saleID, amount)
		if err != nil {
			return err
		}
		if err := uow.CustomerAdvances().Update(ctx, a); err != nil {
			return err
		}
		sale, err := uow.Sales().GetByID(ctx, saleID)
		if err != nil {
			return err
		}
		if _, err := sale.ApplyPayment(amount); err != nil {
			return err
		}
		if err := uow.Sales().Update(ctx, sale); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return valueobjects.Money{}, err
	}
	s.log.Info("advance applied to sale",
		"advance_id", advanceID,
		"sale_id", saleID,
		"amount", amount,
		"remaining", remaining,
	)
	return remaining, nil
}

// ReceivableSummary is the AR roll-up exposed to dashboards.
type ReceivableSummary struct {
	Open     valueobjects.Money
	Overdue0to30  valueobjects.Money
	Overdue31to60 valueobjects.Money
	Overdue61to90 valueobjects.Money
	Overdue90plus valueobjects.Money
}

// OutstandingByCustomer returns the open balance for a customer.
func (s *CustomerPaymentService) OutstandingByCustomer(ctx context.Context, customerID uuid.UUID) (valueobjects.Money, error) {
	// The application layer is expected to call the receivables
	// repository for the full aging breakdown. Here we only need the
	// raw customer debt which is maintained on the customer entity
	// itself.
	uow := repositories.UnitOfWorkFromContext(ctx)
	cust, err := uow.Customers().GetByID(ctx, customerID)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return cust.CurrentDebt, nil
}
