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

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/internal/shared/logger"
)

// StockLedger is the narrow inventory contract consumed by the purchase
// slice. It is satisfied by *inventory.InventoryService and lets the
// purchasing service inject batch stock on receipt and deduct it on
// cancel without depending on the full inventory surface.
type StockLedger interface {
	ReceiveFromPurchase(ctx context.Context, in inventory.ReceiveFromPurchaseInput) (*inventory.InventoryBatch, error)
	VoidPurchaseReceipt(ctx context.Context, companyID uuid.UUID, purchaseLineIDs []uuid.UUID) error
}

// PurchasingService owns the purchase slice.
type PurchasingService struct {
	orders   PurchaseRepository
	payments SupplierPaymentRepository
	stock    StockLedger
	txm      repositories.TransactionManager
	log      *logger.Logger
}

// New returns a PurchasingService ready for use.
func New(
	orders PurchaseRepository,
	payments SupplierPaymentRepository,
	stock StockLedger,
	txm repositories.TransactionManager,
	log *logger.Logger,
) *PurchasingService {
	return &PurchasingService{
		orders:   orders,
		payments: payments,
		stock:    stock,
		txm:      txm,
		log:      log,
	}
}

// CreateInput is the payload for CreatePurchaseOrder. CurrencyCode is
// the transactional currency (often USD for international suppliers).
// Conversion to the company's functional currency is the responsibility
// of the application layer.
type CreateInput struct {
	CompanyID    uuid.UUID
	BranchID     *uuid.UUID
	Number       string
	SupplierID   uuid.UUID
	CurrencyCode valueobjects.CurrencyCode
	ExchangeRate valueobjects.ExchangeRate
	OrderDate    valueobjects.Date
	ExpectedDate *valueobjects.Date
	Notes        string
	Items        []CreateItemInput
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

// receiveStock injects the received goods as inventory batches. It is
// idempotent per line (a line whose batch already exists is skipped),
// so it can be called safely from Create, Approve and MarkAsReceived.
// The lot is the purchase number and the arrival date is the receipt
// date (falling back to the order date).
func (s *PurchasingService) receiveStock(ctx context.Context, po *PurchaseOrder) error {
	if s.stock == nil || len(po.Items) == 0 {
		return nil
	}
	arrival := po.OrderDate
	if po.ReceivedDate != nil {
		arrival = *po.ReceivedDate
	}
	for _, li := range po.Items {
		_, err := s.stock.ReceiveFromPurchase(ctx, inventory.ReceiveFromPurchaseInput{
			CompanyID:      po.CompanyID,
			SupplierID:     &po.SupplierID,
			ProductID:      li.ProductID,
			PurchaseLineID: li.ID,
			LotNumber:      valueobjects.LotNumber(po.Number),
			ArrivalDate:    arrival,
			Quantity:       li.Quantity,
			UnitCost:       li.UnitCostNet(),
			CurrencyCode:   po.CurrencyCode,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// CreatePurchaseOrder validates the input, builds the purchase order
// aggregate, and persists it. The order starts in "pending" status and
// its goods are injected into inventory immediately (idempotently).
func (s *PurchasingService) Create(ctx context.Context, in CreateInput) (*PurchaseOrder, error) {
	if len(in.Items) == 0 {
		return nil, derrors.New("EMPTY_DOCUMENT", "purchase order must have at least one line")
	}
	var out *PurchaseOrder
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		number := in.Number
		if number == "" {
			n, err := s.orders.GetNextNumber(ctx, in.CompanyID)
			if err != nil {
				return err
			}
			number = n
		}
		opts := NewPurchaseOrderOptions{
			CompanyID:    in.CompanyID,
			BranchID:     in.BranchID,
			Number:       number,
			SupplierID:   in.SupplierID,
			CurrencyCode: in.CurrencyCode,
			ExchangeRate: in.ExchangeRate,
			OrderDate:    in.OrderDate,
			ExpectedDate: in.ExpectedDate,
			Notes:        in.Notes,
		}
		po, err := NewPurchaseOrder(time.Now().UTC(), opts)
		if err != nil {
			return err
		}
		for _, it := range in.Items {
			line, err := NewPurchaseOrderItem(NewPurchaseOrderItemOptions{
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
		if err := s.orders.Create(ctx, po); err != nil {
			return err
		}
		if err := s.receiveStock(ctx, po); err != nil {
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
		po, err := s.orders.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if po.Status == enums.PurchaseStatusReceived {
			return derrors.New("ALREADY_RECEIVED", "purchase has already been received")
		}
		if po.Status == enums.PurchaseStatusReconciled {
			return derrors.New("ALREADY_RECONCILED", "purchase is already reconciled")
		}
		if err := po.Approve(); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, po); err != nil {
			return err
		}
		return s.receiveStock(ctx, po)
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
		po, err := s.orders.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := po.MarkAsReceived(at); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, po); err != nil {
			return err
		}
		return s.receiveStock(ctx, po)
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
		return derrors.New("REQUIRED", "cancel reason is required")
	}
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		po, err := s.orders.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := po.Cancel(reason); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, po); err != nil {
			return err
		}
		if s.stock != nil {
			lineIDs := make([]uuid.UUID, 0, len(po.Items))
			for _, li := range po.Items {
				lineIDs = append(lineIDs, li.ID)
			}
			if err := s.stock.VoidPurchaseReceipt(ctx, po.CompanyID, lineIDs); err != nil {
				return err
			}
		}
		return nil
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
func (s *PurchasingService) RegisterSupplierPayment(ctx context.Context, in PayInput, purchaseID uuid.UUID) (*SupplierPayment, error) {
	var out *SupplierPayment
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sp, err := NewSupplierPayment(time.Now().UTC(), NewSupplierPaymentOptions{
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
		if err := s.payments.Create(ctx, sp); err != nil {
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

// MarkPaidInput is the payload for MarkPaid.
type MarkPaidInput struct {
	CompanyID   uuid.UUID
	PaymentDate valueobjects.Date
	Method      enums.PaymentMethod
	Reference   string
	Notes       string
}

// MarkPaid records a supplier payment covering the purchase's full
// outstanding balance and transitions the purchase to "paid", all in
// a single transaction. The payment and its allocation are persisted
// through SupplierPaymentRepository.
func (s *PurchasingService) MarkPaid(ctx context.Context, purchaseID uuid.UUID, in MarkPaidInput) (*PurchaseOrder, error) {
	var out *PurchaseOrder
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		po, err := s.orders.GetByID(ctx, purchaseID)
		if err != nil {
			return err
		}
		if po.IsCancelled() {
			return derrors.Wrap(derrors.ErrPurchaseCancelled, errField("cannot pay a cancelled purchase"))
		}
		if po.Status == enums.PurchaseStatusPaid || po.Status == enums.PurchaseStatusReconciled {
			return derrors.Wrap(derrors.ErrInvalidStateTransition, errField("purchase is already paid"))
		}
		balance := po.Balance()
		if !balance.IsPositive() {
			return derrors.Wrap(derrors.ErrEmptyDocument, errField("purchase has no outstanding balance"))
		}
		method := in.Method
		if !method.Valid() {
			method = enums.PaymentMethodCash
		}
		number, err := s.payments.GetNextNumber(ctx, in.CompanyID)
		if err != nil {
			return err
		}
		sp, err := NewSupplierPayment(time.Now().UTC(), NewSupplierPaymentOptions{
			CompanyID:    in.CompanyID,
			SupplierID:   po.SupplierID,
			Number:       number,
			PaymentDate:  in.PaymentDate,
			Amount:       balance,
			CurrencyCode: po.CurrencyCode,
			ExchangeRate: po.ExchangeRate,
			Method:       method,
			Reference:    in.Reference,
			Notes:        in.Notes,
		})
		if err != nil {
			return err
		}
		if err := sp.ApplyToPurchase(purchaseID, balance); err != nil {
			return err
		}
		if err := s.payments.Create(ctx, sp); err != nil {
			return err
		}
		if _, err := po.ApplyPayment(balance); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, po); err != nil {
			return err
		}
		out = po
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("purchase marked as paid",
		"po_id", purchaseID,
		"number", out.Number,
		"amount", out.Paid,
	)
	return out, nil
}

// Reconcile marks a paid purchase as fully reconciled. Terminal state.
func (s *PurchasingService) Reconcile(ctx context.Context, id uuid.UUID) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		po, err := s.orders.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := po.Reconcile(); err != nil {
			return err
		}
		return s.orders.Update(ctx, po)
	})
	if err != nil {
		return err
	}
	s.log.Info("purchase reconciled", "po_id", id)
	return nil
}

// GetByID returns the purchase order aggregate.
func (s *PurchasingService) GetByID(ctx context.Context, id uuid.UUID) (*PurchaseOrder, error) {
	return s.orders.GetByID(ctx, id)
}

// List returns purchase orders matching the filter.
func (s *PurchasingService) List(ctx context.Context, filter PurchaseFilter) (repositories.Page[*PurchaseOrder], error) {
	return s.orders.List(ctx, filter)
}
