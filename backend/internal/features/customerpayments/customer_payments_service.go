// Package customer_payments owns the customer-payment slice: register
// a payment, allocate it to one or more sales, and surface the
// outstanding receivables balance. Sales operations are delegated to
// sales.SalesService so the payment flow stays consistent with the
// sale's own state machine.
package customerpayments

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/customer"
	"vfinancy/backend/internal/features/sales"
	"vfinancy/backend/internal/shared/logger"
)

// CustomerPaymentService handles customer payments and their
// allocations. Multi-step payment flows compose this service with
// sales.SalesService sharing the same transaction.
type CustomerPaymentService struct {
	payments  sales.CustomerPaymentRepository
	advances  sales.CustomerAdvanceRepository
	sales     sales.SalesRepository
	customers customer.CustomerRepository
	txm       repositories.TransactionManager
	log       *logger.Logger
}

// New returns a CustomerPaymentService ready for use.
func New(
	payments sales.CustomerPaymentRepository,
	advances sales.CustomerAdvanceRepository,
	sales sales.SalesRepository,
	customers customer.CustomerRepository,
	txm repositories.TransactionManager,
	log *logger.Logger,
) *CustomerPaymentService {
	return &CustomerPaymentService{
		payments:  payments,
		advances:  advances,
		sales:     sales,
		customers: customers,
		txm:       txm,
		log:       log,
	}
}

// PayInput is the payload for Register.
type PayInput struct {
	CompanyID      uuid.UUID
	CustomerID     uuid.UUID
	Number         string
	PaymentDate    valueobjects.Date
	Amount         valueobjects.Money
	CurrencyCode   valueobjects.CurrencyCode
	ExchangeRate   valueobjects.ExchangeRate
	Method         enums.PaymentMethod
	BankAccountID  *uuid.UUID
	CashRegisterID *uuid.UUID
	Reference      string
	Notes          string
}

// Register creates a new customer payment. The payment's allocations
// are populated via ApplyToSale; the service itself does NOT touch any
// sale aggregate.
func (s *CustomerPaymentService) Register(ctx context.Context, in PayInput) (*sales.CustomerPayment, error) {
	if in.CompanyID == uuid.Nil || in.CustomerID == uuid.Nil {
		return nil, derrors.New("REQUIRED", "company and customer are required")
	}
	if !in.Amount.IsPositive() {
		return nil, derrors.New("INVALID_PAYMENT", "payment amount must be positive")
	}
	var out *sales.CustomerPayment
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
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
		if err := s.payments.Create(ctx, cp); err != nil {
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

// MarkPaidInput is the payload for MarkPaid.
type MarkPaidInput struct {
	CompanyID   uuid.UUID
	PaymentDate valueobjects.Date
	Method      enums.PaymentMethod
	Reference   string
	Notes       string
}

// MarkPaid records a customer payment covering the sale's full
// outstanding balance, applies it to the sale and reduces the
// customer's debt — all in a single transaction.
func (s *CustomerPaymentService) MarkPaid(ctx context.Context, saleID uuid.UUID, in MarkPaidInput) (*sales.Sale, error) {
	var out *sales.Sale
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sale, err := s.sales.GetByID(ctx, saleID)
		if err != nil {
			return err
		}
		if sale.Status == enums.SaleStatusCancelled {
			return derrors.New("INVALID_STATE_TRANSITION", "cannot collect a cancelled sale")
		}
		if sale.Status == enums.SaleStatusPaid {
			return derrors.New("SALE_ALREADY_PAID", "sale is already fully paid")
		}
		balance := sale.Balance()
		if !balance.IsPositive() {
			return derrors.New("EMPTY_DOCUMENT", "sale has no outstanding balance")
		}
		method := in.Method
		if !method.Valid() {
			method = enums.PaymentMethodCash
		}
		number, err := s.payments.GetNextNumber(ctx, in.CompanyID)
		if err != nil {
			return err
		}
		cp, err := sales.NewCustomerPayment(time.Now().UTC(), sales.NewCustomerPaymentOptions{
			CompanyID:    in.CompanyID,
			CustomerID:   sale.CustomerID,
			Number:       number,
			PaymentDate:  in.PaymentDate,
			Amount:       balance,
			CurrencyCode: sale.CurrencyCode,
			ExchangeRate: sale.ExchangeRate,
			Method:       method,
			Reference:    in.Reference,
			Notes:        in.Notes,
		})
		if err != nil {
			return err
		}
		if err := cp.ApplyToSale(saleID, balance); err != nil {
			return err
		}
		if err := s.payments.Create(ctx, cp); err != nil {
			return err
		}
		if _, err := sale.ApplyPayment(balance); err != nil {
			return err
		}
		if err := s.sales.Update(ctx, sale); err != nil {
			return err
		}
		customer, err := s.customers.GetByID(ctx, sale.CustomerID)
		if err != nil {
			return err
		}
		if _, err := customer.RecordPayment(balance); err != nil {
			return err
		}
		if err := s.customers.Update(ctx, customer); err != nil {
			return err
		}
		out = sale
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("sale marked as paid",
		"sale_id", saleID,
		"number", out.Number,
		"amount", out.Paid,
	)
	return out, nil
}

// ApplyToSale allocates part of a payment to a specific sale. The sale
// is also updated to reflect the new paid amount, in the same
// transaction. This is the only "join point" between the payment and
// sale flows.
func (s *CustomerPaymentService) ApplyToSale(ctx context.Context, paymentID, saleID uuid.UUID, amount valueobjects.Money) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		cp, err := s.payments.GetByID(ctx, paymentID)
		if err != nil {
			return err
		}
		if err := cp.ApplyToSale(saleID, amount); err != nil {
			return err
		}
		if err := s.payments.Update(ctx, cp); err != nil {
			return err
		}
		sale, err := s.sales.GetByID(ctx, saleID)
		if err != nil {
			return err
		}
		if _, err := sale.ApplyPayment(amount); err != nil {
			return err
		}
		if err := s.sales.Update(ctx, sale); err != nil {
			return err
		}
		return nil
	})
}

// AdvanceInput is the payload for RegisterAdvance.
type AdvanceInput struct {
	CompanyID     uuid.UUID
	CustomerID    uuid.UUID
	Number        string
	AdvanceDate   valueobjects.Date
	Amount        valueobjects.Money
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	Method        enums.PaymentMethod
	BankAccountID *uuid.UUID
	Notes         string
}

// RegisterAdvance creates a customer advance. The amount is not
// applied to any sale yet; that happens via ApplyToSale on the
// returned advance.
func (s *CustomerPaymentService) RegisterAdvance(ctx context.Context, in AdvanceInput) (*sales.CustomerAdvance, error) {
	if in.CompanyID == uuid.Nil || in.CustomerID == uuid.Nil {
		return nil, derrors.New("REQUIRED", "company and customer are required")
	}
	if !in.Amount.IsPositive() {
		return nil, derrors.New("INVALID_ADVANCE", "advance amount must be positive")
	}
	var out *sales.CustomerAdvance
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		a, err := sales.NewCustomerAdvance(time.Now().UTC(), sales.NewCustomerAdvanceOptions{
			CompanyID:     in.CompanyID,
			CustomerID:    in.CustomerID,
			Number:        in.Number,
			AdvanceDate:   in.AdvanceDate,
			Amount:        in.Amount,
			CurrencyCode:  in.CurrencyCode,
			ExchangeRate:  in.ExchangeRate,
			Method:        in.Method,
			BankAccountID: in.BankAccountID,
			Notes:         in.Notes,
		})
		if err != nil {
			return err
		}
		if err := s.advances.Create(ctx, a); err != nil {
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
// reduced. Callers compose this with sales.SalesService.ApplyPayment /
// MarkAsPaid in the same transaction.
func (s *CustomerPaymentService) ApplyAdvanceToSale(ctx context.Context, advanceID, saleID uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	var remaining valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		a, err := s.advances.GetByID(ctx, advanceID)
		if err != nil {
			return err
		}
		remaining, err = a.ApplyToSale(saleID, amount)
		if err != nil {
			return err
		}
		if err := s.advances.Update(ctx, a); err != nil {
			return err
		}
		sale, err := s.sales.GetByID(ctx, saleID)
		if err != nil {
			return err
		}
		if _, err := sale.ApplyPayment(amount); err != nil {
			return err
		}
		if err := s.sales.Update(ctx, sale); err != nil {
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
	Open          valueobjects.Money
	Overdue0to30  valueobjects.Money
	Overdue31to60 valueobjects.Money
	Overdue61to90 valueobjects.Money
	Overdue90plus valueobjects.Money
}

// OutstandingByCustomer returns the open balance for a customer. The
// full aging breakdown comes from the receivables repository; here we
// return the raw customer debt which is maintained on the customer
// entity itself.
func (s *CustomerPaymentService) OutstandingByCustomer(ctx context.Context, customerID uuid.UUID) (valueobjects.Money, error) {
	cust, err := s.customers.GetByID(ctx, customerID)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return cust.CurrentDebt, nil
}

// ListPayments returns customer payments matching the filter.
func (s *CustomerPaymentService) ListPayments(ctx context.Context, filter sales.CustomerPaymentFilter) (repositories.Page[*sales.CustomerPayment], error) {
	return s.payments.List(ctx, filter)
}

// ListAdvances returns the customer advances for a given customer.
func (s *CustomerPaymentService) ListAdvances(ctx context.Context, customerID uuid.UUID) ([]*sales.CustomerAdvance, error) {
	return s.advances.ListByCustomer(ctx, customerID)
}
