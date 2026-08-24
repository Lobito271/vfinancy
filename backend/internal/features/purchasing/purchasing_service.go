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
	"github.com/shopspring/decimal"

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

// importSurcharge is the fixed per-dollar overhead (customs, freight,
// logistics) added to the official exchange rate when computing the
// real landed cost of an import in PEN.
const importSurcharge = 0.07

// RealCostPEN computes the landed cost in PEN for an order bought in
// USD: cost_usd * (exchange_rate + importSurcharge).
func RealCostPEN(costUSD valueobjects.Money, rate valueobjects.ExchangeRate) valueobjects.Money {
	m, _ := valueobjects.MoneyFromDecimal(
		costUSD.Decimal().Mul(rate.Decimal().Add(decimal.NewFromFloat(importSurcharge))),
	)
	return m
}

// ProjectedProfitPEN is the expected profit for a customer order:
// sale_price_pen - real_cost_pen.
func ProjectedProfitPEN(salePricePEN, realCostPEN valueobjects.Money) valueobjects.Money {
	return salePricePEN.Sub(realCostPEN)
}

// CreateInput is the payload for CreatePurchaseOrder. CurrencyCode is
// the transactional currency (often USD for international suppliers).
// CostUSD is the order cost in dollars and SalePricePEN the expected
// PEN sale price; both are used for the landed-cost / profit math on
// customer orders.
type CreateInput struct {
	CompanyID           uuid.UUID
	BranchID            *uuid.UUID
	Number              string
	SupplierID          uuid.UUID
	CustomerID          *uuid.UUID
	CreditCardID        *uuid.UUID
	OrderType           enums.OrderType
	CurrencyCode        valueobjects.CurrencyCode
	ExchangeRate        valueobjects.ExchangeRate
	OrderDate           valueobjects.Date
	ExpectedDate        *valueobjects.Date
	ArrivalDate         *valueobjects.Date
	SupplierOrderNumber string
	CostUSD             valueobjects.Money
	SalePricePEN        valueobjects.Money
	Anticipo            valueobjects.Money
	AnticipoDate        *valueobjects.Date
	Notes               string
	Items               []CreateItemInput
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

// paySupplierWithCard pays the full order total with the credit card
// recorded on the purchase order. It persists the supplier payment, its
// allocation to the order, and advances the order's paid amount — all
// in the caller's transaction.
func (s *PurchasingService) paySupplierWithCard(ctx context.Context, po *PurchaseOrder) error {
	number, err := s.payments.GetNextNumber(ctx, po.CompanyID)
	if err != nil {
		return err
	}
	amount := po.CalculateTotal()
	if !amount.IsPositive() {
		return nil
	}
	sp, err := NewSupplierPayment(time.Now().UTC(), NewSupplierPaymentOptions{
		CompanyID:    po.CompanyID,
		SupplierID:   po.SupplierID,
		Number:       number,
		PaymentDate:  po.OrderDate,
		Amount:       amount,
		CurrencyCode: po.CurrencyCode,
		ExchangeRate: po.ExchangeRate,
		Method:       enums.PaymentMethodCard,
		CreditCardID: po.CreditCardID,
		Reference:    "Pago automático con tarjeta",
		Notes:        "Pago automático del pedido " + po.Number,
	})
	if err != nil {
		return err
	}
	if err := sp.ApplyToPurchase(po.ID, amount); err != nil {
		return err
	}
	if err := s.payments.Create(ctx, sp); err != nil {
		return err
	}
	_, err = po.ApplyPayment(amount)
	return err
}

// CreatePurchaseOrder validates the input, builds the purchase order
// aggregate, and persists it. The order starts in "pending" status and
// the supplier is paid in full with the credit card in the same
// transaction (no receivables or inventory movements are created here).
// Customer orders also record the initial anticipo (down payment) as
// the first row of the customer-payment ledger.
func (s *PurchasingService) Create(ctx context.Context, in CreateInput) (*PurchaseOrder, error) {
	if len(in.Items) == 0 {
		return nil, derrors.New("EMPTY_DOCUMENT", "purchase order must have at least one line")
	}
	orderType := in.OrderType
	if orderType == "" {
		orderType = enums.OrderTypeGeneral
	}
	realCost := RealCostPEN(in.CostUSD, in.ExchangeRate)
	profit := ProjectedProfitPEN(in.SalePricePEN, realCost)
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
			CompanyID:           in.CompanyID,
			BranchID:            in.BranchID,
			Number:              number,
			SupplierID:          in.SupplierID,
			CurrencyCode:        in.CurrencyCode,
			ExchangeRate:        in.ExchangeRate,
			OrderDate:           in.OrderDate,
			ExpectedDate:        in.ExpectedDate,
			ArrivalDate:         in.ArrivalDate,
			Notes:               in.Notes,
			OrderType:           orderType,
			CustomerID:          in.CustomerID,
			CreditCardID:        in.CreditCardID,
			SupplierOrderNumber: in.SupplierOrderNumber,
			CostUSD:             in.CostUSD,
			SalePricePEN:       in.SalePricePEN,
			RealCostPEN:        realCost,
			ProjectedProfitPEN: profit,
			Anticipo:           in.Anticipo,
			AnticipoDate:       in.AnticipoDate,
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
		if po.Anticipo.IsPositive() && po.CustomerID != nil {
			number, err := s.orders.GetNextCustomerPaymentNumber(ctx, po.CompanyID)
			if err != nil {
				return err
			}
			pm, err := NewCustomerOrderPayment(time.Now().UTC(), NewCustomerOrderPaymentOptions{
				CompanyID:       po.CompanyID,
				PurchaseOrderID: po.ID,
				CustomerID:      *po.CustomerID,
				Number:          number,
				PaymentDate:     *po.AnticipoDate,
				Amount:          po.Anticipo,
				Method:          enums.PaymentMethodCash,
				CurrencyCode:    valueobjects.PEN,
				ExchangeRate:    valueobjects.One(),
				Reference:       "Anticipo inicial",
				Notes:           "Anticipo inicial del pedido",
			})
			if err != nil {
				return err
			}
			po.CustomerPayments = append(po.CustomerPayments, pm)
		}
		if err := s.orders.Create(ctx, po); err != nil {
			return err
		}
		if err := s.paySupplierWithCard(ctx, po); err != nil {
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
		"order_type", out.OrderType,
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

// MarkAsReceived sets the receipt date (and the arrival date used for
// the inventory aging rule). Called by the warehouse module once the
// goods are physically received; this is what injects the goods into
// inventory as batches.
func (s *PurchasingService) MarkAsReceived(ctx context.Context, id uuid.UUID, at valueobjects.Date) error {
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		po, err := s.orders.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := po.MarkAsReceived(at); err != nil {
			return err
		}
		po.ArrivalDate = &at
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

// Cancel marks the purchase as cancelled with a reason. Inventory is
// restored (compensating movements) and, for customer orders, every
// recorded down payment is refunded automatically.
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
		if s.stock != nil {
			lineIDs := make([]uuid.UUID, 0, len(po.Items))
			for _, li := range po.Items {
				lineIDs = append(lineIDs, li.ID)
			}
			if err := s.stock.VoidPurchaseReceipt(ctx, po.CompanyID, lineIDs); err != nil {
				return err
			}
		}
		if po.OrderType == enums.OrderTypeCustomer {
			refunded, err := s.refundCustomerPayments(ctx, po)
			if err != nil {
				return err
			}
			po.RefundedAmount = po.RefundedAmount.Add(refunded)
		}
		if err := s.orders.Update(ctx, po); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.log.Info("purchase cancelled", "po_id", id, "reason", reason)
	return nil
}

// FaultyInput is the payload for MarkFaulty.
type FaultyInput struct {
	ID          uuid.UUID
	ArrivalDate valueobjects.Date
	Reason      string
}

// MarkFaulty runs the "Llegó en mal estado" workflow: it records the
// arrival date, voids the order (restoring inventory) and automatically
// registers a 100% refund of every down payment made by the customer.
func (s *PurchasingService) MarkFaulty(ctx context.Context, in FaultyInput) (*PurchaseOrder, error) {
	if in.ID == uuid.Nil {
		return nil, derrors.New("REQUIRED", "purchase id is required")
	}
	var out *PurchaseOrder
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		po, err := s.orders.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if err := po.MarkFaulty(in.ArrivalDate, in.Reason); err != nil {
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
		refunded, err := s.refundCustomerPayments(ctx, po)
		if err != nil {
			return err
		}
		po.RefundedAmount = po.RefundedAmount.Add(refunded)
		if err := s.orders.Update(ctx, po); err != nil {
			return err
		}
		out = po
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("purchase marked faulty",
		"po_id", in.ID,
		"arrival_date", in.ArrivalDate,
		"refunded", out.RefundedAmount,
	)
	return out, nil
}

// refundCustomerPayments marks every active customer down payment of
// the order as refunded and returns the total refunded amount.
func (s *PurchasingService) refundCustomerPayments(ctx context.Context, po *PurchaseOrder) (valueobjects.Money, error) {
	var total valueobjects.Money
	now := time.Now().UTC()
	for _, pm := range po.CustomerPayments {
		if pm.IsRefunded() {
			continue
		}
		if err := pm.MarkRefunded(now, "Devolución por anulación del pedido"); err != nil {
			return total, err
		}
		if err := s.orders.UpdateCustomerPayment(ctx, pm); err != nil {
			return total, err
		}
		total = total.Add(pm.Amount)
	}
	return total, nil
}

// CustomerPaymentInput is the payload for RegisterCustomerOrderPayment.
type CustomerPaymentInput struct {
	CompanyID    uuid.UUID
	Number       string
	PaymentDate  valueobjects.Date
	Amount       valueobjects.Money
	CurrencyCode valueobjects.CurrencyCode
	ExchangeRate valueobjects.ExchangeRate
	Method       enums.PaymentMethod
	Reference    string
	Notes        string
}

// RegisterCustomerOrderPayment records a partial down payment
// (anticipo) from the customer against a customer order. The payment is
// persisted to the customer-payment ledger and the order's anticipo is
// advanced in the same transaction.
func (s *PurchasingService) RegisterCustomerOrderPayment(ctx context.Context, purchaseID uuid.UUID, in CustomerPaymentInput) (*CustomerOrderPayment, error) {
	if in.CompanyID == uuid.Nil {
		return nil, derrors.New("REQUIRED", "company is required")
	}
	var out *CustomerOrderPayment
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		po, err := s.orders.GetByID(ctx, purchaseID)
		if err != nil {
			return err
		}
		if po.IsCancelled() {
			return derrors.Wrap(derrors.ErrPurchaseCancelled, errField("cannot pay a cancelled purchase"))
		}
		if err := po.RecordCustomerPayment(in.Amount); err != nil {
			return err
		}
		number := in.Number
		if number == "" {
			n, err := s.orders.GetNextCustomerPaymentNumber(ctx, in.CompanyID)
			if err != nil {
				return err
			}
			number = n
		}
		method := in.Method
		if !method.Valid() {
			method = enums.PaymentMethodCash
		}
		pm, err := NewCustomerOrderPayment(time.Now().UTC(), NewCustomerOrderPaymentOptions{
			CompanyID:       in.CompanyID,
			PurchaseOrderID: po.ID,
			CustomerID:      *po.CustomerID,
			Number:          number,
			PaymentDate:     in.PaymentDate,
			Amount:          in.Amount,
			Method:          method,
			CurrencyCode:    in.CurrencyCode,
			ExchangeRate:    in.ExchangeRate,
			Reference:       in.Reference,
			Notes:           in.Notes,
		})
		if err != nil {
			return err
		}
		if err := s.orders.SaveCustomerPayment(ctx, pm); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, po); err != nil {
			return err
		}
		out = pm
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("customer order payment registered",
		"payment_id", out.ID,
		"purchase_id", purchaseID,
		"amount", out.Amount,
	)
	return out, nil
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
