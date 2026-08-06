// Package purchasing implements the business logic for the purchase
// workflow: creation, approval, receipt, payment, AP management.
//
// The PurchasingService does NOT generate the corresponding journal
// entries — that is the AccountingService's job. The application use
// case layer (Phase 1.5) composes the two: it calls
// PurchasingService.ApproveAndReceive and then
// AccountingService.PostPurchaseApproval in the same transaction.
package purchasing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services"
	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/purchasing"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// PurchasingService owns the purchase workflow.
type PurchasingService struct {
	orders   repositories.PurchaseRepository
	payments repositories.SupplierPaymentRepository
	txm      services.TxManager
	log      *common.Logger
}

// New returns a PurchasingService ready for use.
func New(
	orders repositories.PurchaseRepository,
	payments repositories.SupplierPaymentRepository,
	txm services.TxManager,
	log *common.Logger,
) *PurchasingService {
	return &PurchasingService{
		orders:   orders,
		payments: payments,
		txm:      txm,
		log:      log,
	}
}

// CreateInput is the payload for CreatePurchaseOrder. CurrencyCode is
// the transactional currency (often USD for international suppliers).
// Conversion to the company's functional currency is the responsibility
// of the application layer.
type CreateInput struct {
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	Number        string
	SupplierID    uuid.UUID
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	OrderDate     valueobjects.Date
	ExpectedDate  *valueobjects.Date
	Notes         string
	Items         []CreateItemInput
}

// CreateItemInput is one line of the order.
type CreateItemInput struct {
	ProductID       uuid.UUID
	Quantity        valueobjects.Quantity
	UnitPrice       valueobjects.Money
	DiscountPercent valueobjects.Percentage
	DiscountAmount  valueobjects.Money
	TaxRate         valueobjects.Percentage
	TaxAmount       valueobjects.Money
	Description     string
}

// CreatePurchaseOrder validates the input, builds the purchase order
// aggregate, and persists it. The order starts in "pending" status.
func (s *PurchasingService) Create(ctx context.Context, in CreateInput) (*purchasing.PurchaseOrder, error) {
	if len(in.Items) == 0 {
		return nil, services.EnsureError("EMPTY_DOCUMENT", "purchase order must have at least one line")
	}
	var out *purchasing.PurchaseOrder
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		opts := purchasing.NewPurchaseOrderOptions{
			CompanyID:    in.CompanyID,
			BranchID:     in.BranchID,
			Number:       in.Number,
			SupplierID:   in.SupplierID,
			CurrencyCode: in.CurrencyCode,
			ExchangeRate: in.ExchangeRate,
			OrderDate:    in.OrderDate,
			ExpectedDate: in.ExpectedDate,
			Notes:        in.Notes,
		}
		po, err := purchasing.NewPurchaseOrder(time.Now().UTC(), opts)
		if err != nil {
			return err
		}
		for _, it := range in.Items {
			line, err := purchasing.NewPurchaseOrderItem(purchasing.NewPurchaseOrderItemOptions{
				ProductID:       it.ProductID,
				Quantity:        it.Quantity,
				UnitPrice:       it.UnitPrice,
				DiscountPercent: it.DiscountPercent,
				DiscountAmount:  it.DiscountAmount,
				TaxRate:         it.TaxRate,
				TaxAmount:       it.TaxAmount,
				Description:     it.Description,
			})
			if err != nil {
				return err
			}
			if err := po.AddItem(line); err != nil {
				return err
			}
		}
		if err := uow.PurchaseOrders().Create(ctx, po); err != nil {
			return err
		}
		out = po
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("purchase order created",
		"po_id", out.ID,
		"number", out.Number,
		"supplier_id", out.SupplierID,
		"total", out.CalculateTotal(),
	)
	return out, nil
}

// Approve transitions a purchase order from "pending" to "received".
// Approve can only be called once; a second call returns an error.
// The application layer is expected to generate the corresponding
// journal entry in the same transaction via AccountingService.
func (s *PurchasingService) Approve(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		po, err := uow.PurchaseOrders().GetByID(ctx, id)
		if err != nil {
			return err
		}
		if po.Status == enums.PurchaseStatusReceived {
			return services.EnsureError("ALREADY_RECEIVED", "purchase has already been received")
		}
		if po.Status == enums.PurchaseStatusReconciled {
			return services.EnsureError("ALREADY_RECONCILED", "purchase is already reconciled")
		}
		if err := po.Approve(); err != nil {
			return err
		}
		return uow.PurchaseOrders().Update(ctx, po)
	})
	if err != nil {
		return err
	}
	s.log.Info("purchase approved", "po_id", id)
	return nil
}

// MarkAsReceived sets the receipt date. Called by the warehouse module
// once the goods are physically received.
func (s *PurchasingService) MarkAsReceived(ctx context.Context, id uuid.UUID, at valueobjects.Date) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		po, err := uow.PurchaseOrders().GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := po.MarkAsReceived(at); err != nil {
			return err
		}
		return uow.PurchaseOrders().Update(ctx, po)
	})
	if err != nil {
		return err
	}
	s.log.Info("purchase received", "po_id", id, "at", at)
	return nil
}

// Cancel marks the purchase as cancelled with a reason.
func (s *PurchasingService) Cancel(ctx context.Context, id uuid.UUID, reason string) error {
	if reason == "" {
		return services.EnsureError("REQUIRED", "cancel reason is required")
	}
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		po, err := uow.PurchaseOrders().GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := po.Cancel(reason); err != nil {
			return err
		}
		return uow.PurchaseOrders().Update(ctx, po)
	})
	if err != nil {
		return err
	}
	s.log.Info("purchase cancelled", "po_id", id, "reason", reason)
	return nil
}

// PayInput is the payload for RegisterSupplierPayment. The payment is
// allocated to a purchase order via the SupplierPaymentRepository.
type PayInput struct {
	CompanyID      uuid.UUID
	SupplierID     uuid.UUID
	Number         string
	PaymentDate    valueobjects.Date
	Amount         valueobjects.Money
	CurrencyCode   valueobjects.CurrencyCode
	ExchangeRate   valueobjects.ExchangeRate
	Method         enums.PaymentMethod
	BankAccountID  *uuid.UUID
	CashRegisterID *uuid.UUID
	CreditCardID   *uuid.UUID
	Reference      string
	Notes          string
}

// RegisterSupplierPayment creates a new payment and allocates its full
// amount to the given purchase order. Partial allocations are done
// later via ApplyToPurchase.
func (s *PurchasingService) RegisterSupplierPayment(ctx context.Context, in PayInput, purchaseID uuid.UUID) (*purchasing.SupplierPayment, error) {
	var out *purchasing.SupplierPayment
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		sp, err := purchasing.NewSupplierPayment(time.Now().UTC(), purchasing.NewSupplierPaymentOptions{
			CompanyID:      in.CompanyID,
			SupplierID:     in.SupplierID,
			Number:         in.Number,
			PaymentDate:    in.PaymentDate,
			Amount:         in.Amount,
			CurrencyCode:   in.CurrencyCode,
			ExchangeRate:   in.ExchangeRate,
			Method:         in.Method,
			BankAccountID:  in.BankAccountID,
			CashRegisterID: in.CashRegisterID,
			CreditCardID:   in.CreditCardID,
			Reference:      in.Reference,
			Notes:          in.Notes,
		})
		if err != nil {
			return err
		}
		if err := sp.ApplyToPurchase(purchaseID, in.Amount); err != nil {
			return err
		}
		if err := uow.SupplierPayments().Create(ctx, sp); err != nil {
			return err
		}
		out = sp
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("supplier payment registered",
		"payment_id", out.ID,
		"supplier_id", out.SupplierID,
		"purchase_id", purchaseID,
		"amount", out.Amount,
	)
	return out, nil
}

// Reconcile marks a paid purchase as fully reconciled. Terminal state.
func (s *PurchasingService) Reconcile(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		po, err := uow.PurchaseOrders().GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := po.Reconcile(); err != nil {
			return err
		}
		return uow.PurchaseOrders().Update(ctx, po)
	})
	if err != nil {
		return err
	}
	s.log.Info("purchase reconciled", "po_id", id)
	return nil
}

// GetByID returns the purchase order aggregate.
func (s *PurchasingService) GetByID(ctx context.Context, id uuid.UUID) (*purchasing.PurchaseOrder, error) {
	return s.orders.GetByID(ctx, id)
}
