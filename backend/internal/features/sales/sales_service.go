// Package sales implements the business logic for the sales slice:
// creating a sale (with validation, totals, profit), registering
// payments, customer advances, and cancellation.
//
// SalesService owns the whole "create a sale" operation: it validates
// the customer and credit limit, persists the sale with its lines,
// and records the resulting debt on the customer — all inside a single
// transaction. Inventory and accounting side-effects are added here in
// later phases using the same transaction pattern.
package sales

import (
	"context"
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/customer"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/internal/shared/logger"
)

// StockLedger is the narrow inventory contract consumed by the sales
// slice. It is satisfied by *inventory.InventoryService and lets the
// sales service trigger FIFO stock deduction and void restocking
// without depending on the full inventory surface.
type StockLedger interface {
	ReserveForSale(ctx context.Context, in inventory.ReserveForSaleInput) (valueobjects.Money, error)
	ReturnVoidedSale(ctx context.Context, companyID, saleID uuid.UUID) error
}

// SalesService owns the sales slice.
type SalesService struct {
	orders    SalesRepository
	customers customer.CustomerRepository
	stock     StockLedger
	products  inventory.ProductClassifier
	txm       repositories.TransactionManager
	log       *logger.Logger
}

// New returns a SalesService ready for use.
func New(orders SalesRepository, customers customer.CustomerRepository, stock StockLedger, products inventory.ProductClassifier, txm repositories.TransactionManager, log *logger.Logger) *SalesService {
	return &SalesService{orders: orders, customers: customers, stock: stock, products: products, txm: txm, log: log}
}

// CreateInput is the payload for CreateSale. CurrencyCode is the
// transactional currency; converting to the company's functional
// currency before persisting is the caller's concern.
type CreateInput struct {
	CompanyID    uuid.UUID
	BranchID     *uuid.UUID
	Number       string
	CustomerID   uuid.UUID
	CurrencyCode valueobjects.CurrencyCode
	ExchangeRate valueobjects.ExchangeRate
	DueDate      *valueobjects.Date
	Notes        string
	SellerID     *uuid.UUID
	PriceListID  *uuid.UUID
	Items        []CreateItemInput
}

// CreateItemInput is one line. CostSnapshot is the cost-per-unit at the
// moment of sale, used to compute profit.
type CreateItemInput struct {
	ProductID       uuid.UUID
	Quantity        valueobjects.Quantity
	UnitPrice       valueobjects.Money
	DiscountPercent valueobjects.Percentage
	DiscountAmount  valueobjects.Money
	TaxRate         valueobjects.Percentage
	TaxAmount       valueobjects.Money
	CostSnapshot    valueobjects.Money
	Description     string
}

// endOfDay interprets a date-only due date as the last moment of its
// calendar day (UTC). The sales table enforces due_date >= sale_date,
// and sale_date is the sale's creation timestamp; without this
// normalization a sale due "today" would compare its midnight due date
// against the current time and fail the ck_sales_dates CHECK constraint.
func endOfDay(d *valueobjects.Date) *valueobjects.Date {
	if d == nil {
		return nil
	}
	t := d.UTC()
	eod := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999000, time.UTC)
	out := valueobjects.Date(eod)
	return &out
}

// CreateResult bundles the persisted sale plus the customer's updated
// debt after the sale.
type CreateResult struct {
	Sale        *Sale
	UpdatedDebt valueobjects.Money
}

// Create validates the customer, constructs the sale entity with its
// lines, persists everything and records the sale on the customer —
// all inside a single transaction.
//
// The validation rejects:
//   - an empty document
//   - a duplicate product in the items
//   - a blocked / inactive customer
//   - an over-limit sale
func (s *SalesService) Create(ctx context.Context, in CreateInput) (*CreateResult, error) {
	if len(in.Items) == 0 {
		return nil, derrors.New("EMPTY_DOCUMENT", "sale must have at least one line")
	}
	var out *CreateResult
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {

		customer, err := s.customers.GetByID(ctx, in.CustomerID)
		if err != nil {
			return err
		}

		var total valueobjects.Money
		for _, it := range in.Items {
			sub := it.UnitPrice.MulByDecimal(it.Quantity.Decimal())
			afterDiscount := sub.Sub(it.DiscountAmount)
			total = total.Add(afterDiscount.Add(it.TaxAmount))
		}
		if err := customer.CanPlaceSale(total); err != nil {
			return err
		}
		if in.DueDate == nil {
			due := valueobjects.Date(time.Now().UTC().AddDate(0, 0, customer.PaymentTermDays))
			in.DueDate = &due
		}

		opts := NewSaleOptions{
			CompanyID:    in.CompanyID,
			BranchID:     in.BranchID,
			Number:       in.Number,
			CustomerID:   in.CustomerID,
			CurrencyCode: in.CurrencyCode,
			ExchangeRate: in.ExchangeRate,
			DueDate:      endOfDay(in.DueDate),
			Notes:        in.Notes,
		}
		if opts.Number == "" {
			n, err := s.orders.GetNextNumber(ctx, in.CompanyID)
			if err != nil {
				return err
			}
			opts.Number = n
		}
		sale, err := NewSale(time.Now().UTC(), opts)
		if err != nil {
			return err
		}
		for _, it := range in.Items {
			line, err := NewSaleItem(NewSaleItemOptions{
				ProductID:       it.ProductID,
				Quantity:        it.Quantity,
				UnitPrice:       it.UnitPrice,
				DiscountPercent: it.DiscountPercent,
				DiscountAmount:  it.DiscountAmount,
				TaxRate:         it.TaxRate,
				TaxAmount:       it.TaxAmount,
				CostSnapshot:    it.CostSnapshot,
				Description:     it.Description,
			})
			if err != nil {
				return err
			}
			if err := sale.AddItem(line); err != nil {
				return err
			}
		}
		// Deduct stock (FIFO) for physical products and snapshot the
		// weighted average cost per line. Services are not stocked and
		// keep a zero cost snapshot. All movements happen on the same
		// transaction; a shortfall aborts the whole sale.
		if s.stock != nil {
			for _, li := range sale.Items {
				isService, err := s.products.IsService(ctx, li.ProductID)
				if err != nil {
					return err
				}
				if isService {
					continue
				}
				cost, err := s.stock.ReserveForSale(ctx, inventory.ReserveForSaleInput{
					CompanyID: sale.CompanyID,
					ProductID: li.ProductID,
					Quantity:  li.Quantity,
					SaleID:    sale.ID,
				})
				if err != nil {
					return err
				}
				li.CostSnapshot = cost
			}
			if err := sale.Recalculate(); err != nil {
				return err
			}
		}
		if err := s.orders.Create(ctx, sale); err != nil {
			return err
		}

		updatedDebt := customer.RecordSale(sale.CalculateTotal())
		if err := s.customers.Update(ctx, customer); err != nil {
			return err
		}

		out = &CreateResult{
			Sale:        sale,
			UpdatedDebt: updatedDebt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("sale created",
		"sale_id", out.Sale.ID,
		"number", out.Sale.Number,
		"customer_id", out.Sale.CustomerID,
		"total", out.Sale.CalculateTotal(),
		"profit", out.Sale.CalculateProfit(),
		"updated_debt", out.UpdatedDebt,
	)
	return out, nil
}

// CancelInput cancels an existing sale with a reason.
type CancelInput struct {
	ID     uuid.UUID
	Reason string
}

// Cancel marks the sale as cancelled.
func (s *SalesService) Cancel(ctx context.Context, in CancelInput) (*Sale, error) {
	if in.Reason == "" {
		return nil, derrors.New("REQUIRED", "cancel reason is required")
	}
	var out *Sale
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sale, err := s.orders.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		outstanding := sale.Balance()
		if err := sale.Cancel(in.Reason); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, sale); err != nil {
			return err
		}
		if s.stock != nil {
			if err := s.stock.ReturnVoidedSale(ctx, sale.CompanyID, sale.ID); err != nil {
				return err
			}
		}
		if outstanding.IsPositive() {
			customer, err := s.customers.GetByID(ctx, sale.CustomerID)
			if err != nil {
				return err
			}
			if _, err := customer.RecordPayment(outstanding); err != nil {
				return err
			}
			if err := s.customers.Update(ctx, customer); err != nil {
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

// ApplyPayment reduces the sale's paid amount. The sale's status is
// updated automatically (pending / partial / paid).
func (s *SalesService) ApplyPayment(ctx context.Context, id uuid.UUID, amount valueobjects.Money) (valueobjects.Money, error) {
	var balance valueobjects.Money
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sale, err := s.orders.GetByID(ctx, id)
		if err != nil {
			return err
		}
		balance, err = sale.ApplyPayment(amount)
		if err != nil {
			return err
		}
		if err := s.orders.Update(ctx, sale); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return valueobjects.Money{}, err
	}
	s.log.Info("sale payment applied",
		"sale_id", id,
		"amount", amount,
		"balance", balance,
		"status", "updated",
	)
	return balance, nil
}

// MarkAsPaid is a convenience for the "all paid in cash" workflow.
func (s *SalesService) MarkAsPaid(ctx context.Context, id uuid.UUID) (*Sale, error) {
	var out *Sale
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		sale, err := s.orders.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := sale.MarkAsPaid(); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, sale); err != nil {
			return err
		}
		out = sale
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("sale marked as paid", "sale_id", id)
	return out, nil
}

// OutstandingBalance returns the sale's remaining balance (total -
// paid). Used by the receivables service.
func (s *SalesService) OutstandingBalance(ctx context.Context, id uuid.UUID) (valueobjects.Money, error) {
	sale, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return sale.Balance(), nil
}

// GetByID returns the sale aggregate.
func (s *SalesService) GetByID(ctx context.Context, id uuid.UUID) (*Sale, error) {
	return s.orders.GetByID(ctx, id)
}

// List returns sales matching the filter.
func (s *SalesService) List(ctx context.Context, filter SaleFilter) (repositories.Page[*Sale], error) {
	return s.orders.List(ctx, filter)
}
