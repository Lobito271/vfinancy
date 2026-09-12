// Package sales implements the business logic for the sales slice:
// creating a sale (stock or client order) with cost snapshots and
// profit, registering customer payments, and cancellation.
//
// SalesService owns the whole "create a sale" operation: it validates
// the input, reserves stock (FIFO) or builds the client purchase
// order, persists the sale with its lines, and records the resulting
// customer debt and any initial payment — all inside a single
// transaction.
package sales

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/customer"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/internal/features/product"
	"vfinancy/backend/internal/features/purchasing"
	"vfinancy/backend/internal/shared/apperrors"
)

// stockReserver is the narrow inventory contract consumed by the
// sales slice. It is satisfied by *inventory.InventoryService.
type stockReserver interface {
	ReserveForSale(ctx context.Context, in inventory.ReserveForSaleInput) (valueobjects.Money, error)
	ReturnVoidedSale(ctx context.Context, saleID uuid.UUID) error
}

// ClientOrderCreator is the narrow purchasing contract consumed by the
// sales slice. It is satisfied by *purchasing.PurchasingService; the
// line type lives in the purchasing package to keep the dependency
// one-way (sales -> purchasing). rate is the USD->PEN rate snapshotted
// onto the linked import order.
type ClientOrderCreator interface {
	CreateClientOrder(ctx context.Context, customerID, saleID uuid.UUID, rate valueobjects.ExchangeRate, lines []purchasing.ClientOrderLine) error
}

// ClientOrderRateProvider resolves the USD->PEN exchange rate used to
// carry a client order's USD cost into PEN. When unset the rate falls
// back to 1 (no conversion).
type ClientOrderRateProvider func(ctx context.Context) valueobjects.ExchangeRate

// SalesService owns the sales slice.
type SalesService struct {
	sales        SaleRepository
	payments     CustomerPaymentRepository
	customers    *customer.CustomerService
	products     *product.ProductService
	stock        stockReserver
	clientOrders ClientOrderCreator
	clientRate   ClientOrderRateProvider
	txm          repositories.TransactionManager
	log          *logger.Logger
}

// New returns a SalesService ready for use.
func New(sales SaleRepository, payments CustomerPaymentRepository, customers *customer.CustomerService, products *product.ProductService, stock stockReserver, clientOrders ClientOrderCreator, txm repositories.TransactionManager, log *logger.Logger) *SalesService {
	return &SalesService{sales: sales, payments: payments, customers: customers, products: products, stock: stock, clientOrders: clientOrders, txm: txm, log: log}
}

// SetClientOrderRateProvider installs the USD->PEN rate source used for
// client-order cost snapshots. Without it client orders keep their USD
// cost unchanged.
func (s *SalesService) SetClientOrderRateProvider(fn ClientOrderRateProvider) {
	if fn != nil {
		s.clientRate = fn
	}
}

// clientOrderRate returns the configured USD->PEN rate, defaulting to 1.
func (s *SalesService) clientOrderRate(ctx context.Context) valueobjects.ExchangeRate {
	if s.clientRate != nil {
		return s.clientRate(ctx)
	}
	return valueobjects.One()
}

// ItemInput is one requested sale line.
type ItemInput struct {
	ProductID uuid.UUID
	Quantity  valueobjects.Quantity
	UnitPrice valueobjects.Money
}

// CreateInput is the payload for Create. A nil DueDate means a cash
// sale (contado): the full total must be paid up front. A set DueDate
// means a credit sale, with an optional InitialPayment.
type CreateInput struct {
	CustomerID     uuid.UUID
	SaleType       enums.SaleType
	Date           time.Time
	DueDate        *time.Time
	PaymentMethod  enums.PaymentMethod
	Notes          string
	Items          []ItemInput
	InitialPayment valueobjects.Money
}

// CreateResult bundles the persisted sale with its lines.
type CreateResult struct {
	Sale  *Sale
	Items []*SaleItem
}

// validatePaymentMethod defaults an empty method to cash and rejects
// unknown values.
func validatePaymentMethod(m enums.PaymentMethod) (enums.PaymentMethod, error) {
	if m == "" {
		return enums.PaymentMethodCash, nil
	}
	if !m.Valid() {
		return "", apperrors.Errorf(apperrors.ErrValidation, "método de pago inválido")
	}
	return m, nil
}

// resolveInitialPayment returns the up-front payment for the sale.
// Cash sales (no due date) require the full total; credit sales
// accept any amount between zero and the total.
func resolveInitialPayment(contado bool, initial, total valueobjects.Money) (valueobjects.Money, error) {
	if initial.IsNegative() {
		return valueobjects.Zero(), apperrors.Errorf(apperrors.ErrValidation, "el pago inicial no puede ser negativo")
	}
	if !contado {
		if initial.GreaterThan(total) {
			return valueobjects.Zero(), apperrors.Errorf(apperrors.ErrValidation, "el pago inicial no puede superar el total de la venta")
		}
		return initial, nil
	}
	if initial.IsZero() {
		return total, nil
	}
	if !initial.Equals(total) {
		return valueobjects.Zero(), apperrors.Errorf(apperrors.ErrValidation, "una venta al contado requiere el pago completo")
	}
	return initial, nil
}

// Create validates the input, reserves stock or places the client
// order, persists the sale with its lines and records the customer
// debt plus any initial payment — all inside a single transaction.
func (s *SalesService) Create(ctx context.Context, in CreateInput) (*CreateResult, error) {
	if !in.SaleType.Valid() {
		return nil, apperrors.Errorf(apperrors.ErrValidation, "tipo de venta inválido")
	}
	if len(in.Items) == 0 {
		return nil, apperrors.Errorf(apperrors.ErrValidation, "la venta debe tener al menos un producto")
	}
	for _, it := range in.Items {
		if it.ProductID == uuid.Nil {
			return nil, apperrors.Errorf(apperrors.ErrValidation, "cada línea requiere un producto")
		}
		if !it.Quantity.IsPositive() {
			return nil, apperrors.Errorf(apperrors.ErrValidation, "la cantidad debe ser mayor a cero")
		}
		if !it.UnitPrice.IsPositive() {
			return nil, apperrors.Errorf(apperrors.ErrValidation, "el precio unitario debe ser mayor a cero")
		}
	}
	method, err := validatePaymentMethod(in.PaymentMethod)
	if err != nil {
		return nil, err
	}
	if in.DueDate != nil && in.DueDate.Before(in.Date) {
		return nil, apperrors.Errorf(apperrors.ErrValidation, "la fecha de vencimiento no puede ser anterior a la fecha de venta")
	}

	clientRate := s.clientOrderRate(ctx)
	var out *CreateResult
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.customers.GetByID(ctx, in.CustomerID); err != nil {
			return err
		}
		now := time.Now().UTC()
		number, err := s.sales.NextNumber(ctx)
		if err != nil {
			return err
		}
		sale, err := NewSale(now, NewSaleOptions{
			Number:     number,
			CustomerID: in.CustomerID,
			SaleType:   in.SaleType,
			SaleDate:   in.Date,
			DueDate:    in.DueDate,
			Notes:      in.Notes,
		})
		if err != nil {
			return err
		}
		items := make([]*SaleItem, 0, len(in.Items))
		for i, it := range in.Items {
			li, err := NewSaleItem(now, NewSaleItemOptions{
				SaleID:     sale.ID,
				ProductID:  it.ProductID,
				LineNumber: i + 1,
				Quantity:   it.Quantity,
				UnitPrice:  it.UnitPrice,
			})
			if err != nil {
				return err
			}
			items = append(items, li)
		}

		costTotal := valueobjects.Zero()
		lines := make([]purchasing.ClientOrderLine, 0, len(items))
		for _, li := range items {
			switch in.SaleType {
			case enums.SaleTypeStock:
				if s.stock == nil {
					return derrors.New("INTERNAL", "inventory is not configured")
				}
				unitCost, err := s.stock.ReserveForSale(ctx, inventory.ReserveForSaleInput{
					ProductID: li.ProductID,
					Quantity:  li.Quantity,
					SaleID:    sale.ID,
				})
				if err != nil {
					return err
				}
				li.CostSnapshot = unitCost
			case enums.SaleTypeClientOrder:
				prod, err := s.products.GetByID(ctx, li.ProductID)
				if err != nil {
					return err
				}
				li.CostSnapshot = clientRate.Convert(prod.CostUSD)
				lines = append(lines, purchasing.ClientOrderLine{
					ProductID:    li.ProductID,
					Description:  prod.Description,
					Quantity:     li.Quantity,
					SalePricePen: li.UnitPrice,
				})
			}
			costTotal = costTotal.Add(li.CostSnapshot.MulByDecimal(li.Quantity.Decimal()))
		}

		total := valueobjects.Zero()
		for _, li := range items {
			total = total.Add(li.LineTotal)
		}
		sale.Items = items
		sale.Total = total
		sale.CostTotal = costTotal
		sale.Profit = total.Sub(costTotal)

		initial, err := resolveInitialPayment(in.DueDate == nil, in.InitialPayment, total)
		if err != nil {
			return err
		}
		if initial.IsPositive() {
			if _, err := sale.RegisterPayment(initial); err != nil {
				return err
			}
		}
		if err := sale.Validate(); err != nil {
			return err
		}
		if err := s.sales.Create(ctx, sale, items); err != nil {
			return err
		}
		if in.SaleType == enums.SaleTypeClientOrder {
			if s.clientOrders == nil {
				return derrors.New("INTERNAL", "purchasing is not configured")
			}
			if err := s.clientOrders.CreateClientOrder(ctx, in.CustomerID, sale.ID, clientRate, lines); err != nil {
				return err
			}
		}
		if _, err := s.customers.RecordSale(ctx, in.CustomerID, total); err != nil {
			return err
		}
		if initial.IsPositive() {
			if err := s.recordInitialPayment(ctx, sale, initial, method, now); err != nil {
				return err
			}
		}
		out = &CreateResult{Sale: sale, Items: items}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("sale created",
		"sale_id", out.Sale.ID,
		"number", out.Sale.Number,
		"sale_type", out.Sale.SaleType.String(),
		"total", out.Sale.Total.String(),
		"profit", out.Sale.Profit.String(),
		"status", out.Sale.Status.String(),
	)
	return out, nil
}

// recordInitialPayment persists the up-front payment for a sale
// together with its allocation, and reduces the customer debt.
func (s *SalesService) recordInitialPayment(ctx context.Context, sale *Sale, amount valueobjects.Money, method enums.PaymentMethod, now time.Time) error {
	number, err := s.payments.NextNumber(ctx)
	if err != nil {
		return err
	}
	payment, err := NewCustomerPayment(now, NewCustomerPaymentOptions{
		CustomerID:    sale.CustomerID,
		Number:        number,
		PaymentDate:   sale.SaleDate,
		Amount:        amount,
		PaymentMethod: method,
	})
	if err != nil {
		return err
	}
	allocations := []PaymentAllocation{{
		ID:                uuid.New(),
		CustomerPaymentID: payment.ID,
		SaleID:            sale.ID,
		AllocatedAmount:   amount,
		CreatedAt:         now,
	}}
	if err := s.payments.Create(ctx, payment, allocations); err != nil {
		return err
	}
	_, err = s.customers.RecordPayment(ctx, sale.CustomerID, amount)
	return err
}

// CancelInput cancels an existing sale with a reason.
type CancelInput struct {
	ID     uuid.UUID
	Reason string
}

// Cancel marks the sale as cancelled, returns stock for stock sales
// and removes the outstanding balance from the customer debt — all
// inside a single transaction.
func (s *SalesService) Cancel(ctx context.Context, in CancelInput) (*Sale, error) {
	if in.Reason == "" {
		return nil, apperrors.Errorf(apperrors.ErrValidation, "el motivo de anulación es obligatorio")
	}
	var out *Sale
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sale, err := s.sales.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		outstanding := sale.Outstanding()
		if err := sale.Cancel(in.Reason); err != nil {
			return err
		}
		if err := s.sales.Update(ctx, sale); err != nil {
			return err
		}
		if sale.SaleType == enums.SaleTypeStock && s.stock != nil {
			if err := s.stock.ReturnVoidedSale(ctx, sale.ID); err != nil {
				return err
			}
		}
		if outstanding.IsPositive() {
			if _, err := s.customers.AdjustDebt(ctx, sale.CustomerID, outstanding.Neg()); err != nil {
				return err
			}
		}
		out = sale
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("sale cancelled", "sale_id", in.ID, "reason", in.Reason)
	return out, nil
}

// PaymentInput is the payload for ApplyPayment.
type PaymentInput struct {
	Amount        valueobjects.Money
	PaymentMethod enums.PaymentMethod
	Reference     string
	Date          time.Time
}

// ApplyPayment records a customer payment against the sale, updates
// the paid amount and status, and reduces the customer debt. It
// returns the remaining outstanding balance.
func (s *SalesService) ApplyPayment(ctx context.Context, id uuid.UUID, in PaymentInput) (valueobjects.Money, error) {
	remaining, _, err := s.pay(ctx, id, in)
	if err != nil {
		return valueobjects.Money{}, err
	}
	s.log.Info("sale payment applied", "sale_id", id, "amount", in.Amount.String(), "remaining", remaining.String())
	return remaining, nil
}

// MarkAsPaid settles the full outstanding balance as one cash payment
// made today.
func (s *SalesService) MarkAsPaid(ctx context.Context, id uuid.UUID) (*Sale, error) {
	sale, err := s.sales.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sale.Outstanding().IsZero() {
		return nil, derrors.Wrap(derrors.ErrSaleAlreadyPaid, errField("sale is already fully paid"))
	}
	_, out, err := s.pay(ctx, id, PaymentInput{
		Amount:        sale.Outstanding(),
		PaymentMethod: enums.PaymentMethodCash,
		Date:          time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("sale marked as paid", "sale_id", id)
	return out, nil
}

// pay runs the shared payment workflow inside a transaction: it
// validates the amount against the outstanding balance, persists the
// payment with its allocation, updates the sale and reduces the
// customer debt. It returns the remaining balance and the updated
// sale.
func (s *SalesService) pay(ctx context.Context, id uuid.UUID, in PaymentInput) (valueobjects.Money, *Sale, error) {
	if !in.Amount.IsPositive() {
		return valueobjects.Zero(), nil, derrors.Wrap(derrors.ErrInvalidPayment, errField("payment amount must be positive"))
	}
	method, err := validatePaymentMethod(in.PaymentMethod)
	if err != nil {
		return valueobjects.Zero(), nil, err
	}
	var remaining valueobjects.Money
	var updated *Sale
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sale, err := s.sales.GetByID(ctx, id)
		if err != nil {
			return err
		}
		outstanding := sale.Outstanding()
		if outstanding.IsZero() {
			return derrors.Wrap(derrors.ErrSaleAlreadyPaid, errField("sale is already fully paid"))
		}
		if in.Amount.GreaterThan(outstanding) {
			return derrors.Wrap(derrors.ErrPaymentExceedsBalance, errField("payment exceeds outstanding balance"))
		}
		number, err := s.payments.NextNumber(ctx)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		payment, err := NewCustomerPayment(now, NewCustomerPaymentOptions{
			CustomerID:    sale.CustomerID,
			Number:        number,
			PaymentDate:   in.Date,
			Amount:        in.Amount,
			PaymentMethod: method,
			Reference:     in.Reference,
		})
		if err != nil {
			return err
		}
		allocations := []PaymentAllocation{{
			ID:                uuid.New(),
			CustomerPaymentID: payment.ID,
			SaleID:            sale.ID,
			AllocatedAmount:   in.Amount,
			CreatedAt:         now,
		}}
		if err := s.payments.Create(ctx, payment, allocations); err != nil {
			return err
		}
		remaining, err = sale.RegisterPayment(in.Amount)
		if err != nil {
			return err
		}
		if err := s.sales.Update(ctx, sale); err != nil {
			return err
		}
		if _, err := s.customers.RecordPayment(ctx, sale.CustomerID, in.Amount); err != nil {
			return err
		}
		updated = sale
		return nil
	})
	if err != nil {
		return valueobjects.Zero(), nil, err
	}
	return remaining, updated, nil
}

// OutstandingBalance returns the sale's remaining balance (total -
// paid).
func (s *SalesService) OutstandingBalance(ctx context.Context, id uuid.UUID) (valueobjects.Money, error) {
	sale, err := s.sales.GetByID(ctx, id)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return sale.Outstanding(), nil
}

// GetByID returns the sale with its line items.
func (s *SalesService) GetByID(ctx context.Context, id uuid.UUID) (*Sale, error) {
	return s.sales.GetByID(ctx, id)
}

// List returns sales matching the filter.
func (s *SalesService) List(ctx context.Context, filter SaleFilter) (repositories.Page[*Sale], error) {
	return s.sales.List(ctx, filter)
}

// ListItems returns the line items of a sale.
func (s *SalesService) ListItems(ctx context.Context, saleID uuid.UUID) ([]*SaleItem, error) {
	return s.sales.ListItems(ctx, saleID)
}

// ListPayments returns customer payments matching the filter.
func (s *SalesService) ListPayments(ctx context.Context, filter CustomerPaymentFilter) (repositories.Page[*CustomerPayment], error) {
	return s.payments.List(ctx, filter)
}

// ListPaymentsForSale returns the payments allocated to a sale.
func (s *SalesService) ListPaymentsForSale(ctx context.Context, saleID uuid.UUID) ([]*CustomerPayment, error) {
	return s.payments.ListForSale(ctx, saleID)
}

// ListCollections returns the sale allocations of active payments
// received in [from, to), used by dashboard analytics.
func (s *SalesService) ListCollections(ctx context.Context, from, to time.Time) ([]SaleCollection, error) {
	return s.payments.ListCollections(ctx, from, to)
}
