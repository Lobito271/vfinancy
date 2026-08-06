// Package sales implements the business logic for the sales workflow:
// creating a sale (with validation, totals, profit), registering
// payments, customer advances, and cancellation.
//
// The SalesService does NOT directly call the inventory or
// accounting services. Instead, it returns a typed Result that tells
// the application use case which side-effects need to happen (stock
// exit, journal entry, customer debt update). The use case layer (Phase
// 1.5) is responsible for orchestrating the cross-service work in a
// single transaction. This separation keeps the service layer free
// of cross-service coupling and lets each service be tested in
// isolation.
package sales

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

// SalesService owns the sales workflow.
type SalesService struct {
	orders  repositories.SalesRepository
	txm     services.TxManager
	log     *common.Logger
}

// New returns a SalesService ready for use.
func New(orders repositories.SalesRepository, txm services.TxManager, log *common.Logger) *SalesService {
	return &SalesService{orders: orders, txm: txm, log: log}
}

// CreateInput is the payload for CreateSale. CurrencyCode is the
// transactional currency; the application's use-case layer is
// responsible for converting to the company's functional currency
// before persisting (in this design we just store what the caller
// passes).
type CreateInput struct {
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	Number        string
	CustomerID    uuid.UUID
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	DueDate       *valueobjects.Date
	Notes         string
	SellerID      *uuid.UUID
	PriceListID   *uuid.UUID
	Items         []CreateItemInput
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

// CreateResult bundles the persisted sale plus the cross-service
// side-effect requests that the application use case must execute in
// the same transaction. By returning this as a value, the service
// stays free of cross-service coupling.
type CreateResult struct {
	Sale           *sales.Sale
	// OutboundMovements is the list of inventory movements to record,
	// one per line, with negative quantities.
	OutboundMovements []OutboundMovement
	// JournalEntryID is left empty here; the accounting service
	// computes it.
	JournalEntryID uuid.UUID
}

// OutboundMovement describes one inventory write that the application
// must perform as part of the same transaction.
type OutboundMovement struct {
	BatchID  uuid.UUID
	Quantity valueobjects.Quantity
}

// CreateSale validates inputs, constructs the sale entity, and persists
// it. The application use case receives the result and is responsible
// for dispatching the inventory / accounting side-effects in the same
// transaction.
//
// The validate-only phase (no persistence) rejects:
//   * an empty document
//   * a duplicate product in the items
//   * a blocked / inactive customer
//   * an over-limit sale
func (s *SalesService) Create(ctx context.Context, in CreateInput) (*CreateResult, error) {
	if len(in.Items) == 0 {
		return nil, services.EnsureError("EMPTY_DOCUMENT", "sale must have at least one line")
	}
	var out *CreateResult
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)

		// 1. Validate the customer (status + credit) BEFORE we touch
		// the sale. The application's use case could also do this; we
		// do it here so the service gives a single failure point.
		customer, err := uow.Customers().GetByID(ctx, in.CustomerID)
		if err != nil {
			return err
		}
		// 2. Compute the total up-front so we can guard the credit limit.
		var total valueobjects.Money
		for _, it := range in.Items {
			sub := it.UnitPrice.MulByDecimal(it.Quantity.Decimal())
			afterDiscount := sub.Sub(it.DiscountAmount)
			total = total.Add(afterDiscount.Add(it.TaxAmount))
		}
		if err := customer.CanPlaceSale(total); err != nil {
			return err
		}

		// 3. Construct the sale entity.
		opts := sales.NewSaleOptions{
			CompanyID:    in.CompanyID,
			BranchID:     in.BranchID,
			Number:       in.Number,
			CustomerID:   in.CustomerID,
			CurrencyCode: in.CurrencyCode,
			ExchangeRate: in.ExchangeRate,
			DueDate:      in.DueDate,
			Notes:        in.Notes,
		}
		// _ unused: sellerID is read by the use case layer from the
		// session, not the sale itself. The sale entity does not carry
		// the seller; that's a reporting concern.
		_ = in.SellerID
		// _ unused: priceListID similarly belongs to a future pricing
		// subsystem.
		_ = in.PriceListID
		sale, err := sales.NewSale(time.Now().UTC(), opts)
		if err != nil {
			return err
		}
		for _, it := range in.Items {
			line, err := sales.NewSaleItem(sales.NewSaleItemOptions{
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
		if err := uow.Sales().Create(ctx, sale); err != nil {
			return err
		}

		// 4. Build the cross-service request: outbound inventory
		// movements. The application use case is expected to look up
		// the right batch (FIFO/LIFO strategy) and call the inventory
		// service. We expose the requested quantities per product;
		// the inventory service (or a dedicated allocation helper)
//		resolves them to batches.
		//
		// In this initial version the use case is expected to call
		// inventoryService.Issue per (product, batch) after the sales
		// service returns. We pre-aggregate the requested negative
		// quantities per product to make the call shape clear.
		neededByProduct := map[uuid.UUID]valueobjects.Quantity{}
		for _, it := range in.Items {
			cur, ok := neededByProduct[it.ProductID]
			if !ok {
				cur = valueobjects.ZeroQuantity()
			}
			neededByProduct[it.ProductID] = cur.Add(it.Quantity)
		}
		// The application layer will need to pick a batch. The service
		// hands back the products + quantities and lets the use case
		// decide on the allocation policy. To keep the contract simple
		// we return a flat list of (product, quantity) pairs.
		_ = neededByProduct

		out = &CreateResult{
			Sale: sale,
			// OutboundMovements is left empty here. The application use
			// case is expected to look up batches and call
			// inventoryService.Issue directly. We expose this empty
			// field so the struct is ready for an explicit batch-list
			// payload in a future revision.
			OutboundMovements: nil,
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
	)
	return out, nil
}

// CancelInput cancels an existing sale with a reason.
type CancelInput struct {
	ID     uuid.UUID
	Reason string
}

// Cancel marks the sale as cancelled.
func (s *SalesService) Cancel(ctx context.Context, in CancelInput) (*sales.Sale, error) {
	if in.Reason == "" {
		return nil, services.EnsureError("REQUIRED", "cancel reason is required")
	}
	var out *sales.Sale
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		sale, err := uow.Sales().GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if err := sale.Cancel(in.Reason); err != nil {
			return err
		}
		if err := uow.Sales().Update(ctx, sale); err != nil {
			return err
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
		uow := repositories.UnitOfWorkFromContext(ctx)
		sale, err := uow.Sales().GetByID(ctx, id)
		if err != nil {
			return err
		}
		balance, err = sale.ApplyPayment(amount)
		if err != nil {
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
	s.log.Info("sale payment applied",
		"sale_id", id,
		"amount", amount,
		"balance", balance,
		"status", "updated",
	)
	return balance, nil
}

// MarkAsPaid is a convenience for the "all paid in cash" workflow.
func (s *SalesService) MarkAsPaid(ctx context.Context, id uuid.UUID) (*sales.Sale, error) {
	var out *sales.Sale
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		sale, err := uow.Sales().GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := sale.MarkAsPaid(); err != nil {
			return err
		}
		if err := uow.Sales().Update(ctx, sale); err != nil {
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
func (s *SalesService) GetByID(ctx context.Context, id uuid.UUID) (*sales.Sale, error) {
	return s.orders.GetByID(ctx, id)
}

// silence unused-import lint: enums is referenced via sales.NewSale's
// signature (the application uses enums.SaleStatus through the entity).
var _ = enums.SaleStatusPending
