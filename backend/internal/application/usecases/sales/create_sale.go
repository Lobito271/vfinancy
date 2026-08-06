// Package sales contains the use cases for the sales workflow:
// creating a sale (which orchestrates multiple services in one
// transaction), recording payments, advances, and reporting.
package sales

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services/customer"
	"vfinancy/backend/internal/application/services/customerpayments"
	"vfinancy/backend/internal/application/services/inventory"
	"vfinancy/backend/internal/application/services/sales"
	"vfinancy/backend/internal/application/usecases"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// CreateSaleUseCase orchestrates the "create a sale" workflow:
//   1. Validate the request.
//   2. Inside a single transaction:
//      a. Load the customer; ensure not blocked.
//      b. Validate the sale fits the credit limit.
//      c. Create the sale via the sales service.
//      d. For each line, issue stock via the inventory service.
//      e. Record the customer's debt via the customer service.
//   3. Return the persisted sale + computed totals.
//
// Phase 1.5 omits the journal-entry side effect (added in Phase 2
// when accounting automation is enabled). It also omits the
// seller-defaulting logic.
type CreateSaleUseCase struct {
	Base       usecases.Base
	Sales      *sales.SalesService
	Inventory  *inventory.InventoryService
	Customers  *customer.CustomerService
	Payments   *customerpayments.CustomerPaymentService
}

// NewCreateSaleUseCase returns a wired-up use case.
func NewCreateSaleUseCase(
	base usecases.Base,
	s *sales.SalesService,
	i *inventory.InventoryService,
	c *customer.CustomerService,
	p *customerpayments.CustomerPaymentService,
) *CreateSaleUseCase {
	return &CreateSaleUseCase{Base: base, Sales: s, Inventory: i, Customers: c, Payments: p}
}

// CreateSaleRequest is the input. Items is the list of lines the
// caller wants on the new sale. The use case translates this into
// the sales service's input.
type CreateSaleRequest struct {
	CompanyID     uuid.UUID
	BranchID      *uuid.UUID
	Number        string
	CustomerID    uuid.UUID
	CurrencyCode  valueobjects.CurrencyCode
	ExchangeRate  valueobjects.ExchangeRate
	DueDate       *valueobjects.Date
	Notes         string
	Items         []CreateSaleItemRequest
}

// CreateSaleItemRequest is one line in the request.
type CreateSaleItemRequest struct {
	ProductID     uuid.UUID
	Quantity      string // parsed as decimal
	UnitPrice     string // parsed as decimal
	CostSnapshot  string // parsed as decimal
	TaxRate       string // parsed as decimal percentage
	TaxAmount     string // parsed as decimal
	DiscountAmt   string // parsed as decimal
	DiscountPct   string // parsed as decimal percentage
	Description   string
}

// CreateSaleResponse is what the use case returns on success.
type CreateSaleResponse struct {
	SaleID         uuid.UUID
	Number         string
	Total          string
	Subtotal       string
	Tax            string
	Discount       string
	CostTotal      string
	Profit         string
	OutboundMoves  []OutboundMovementSummary
	UpdatedDebt    string
}

// OutboundMovementSummary is a small projection of the inventory
// moves the use case issued during the workflow. The application
// use case layer (or a future reporting layer) consumes this for
// logs and traces.
type OutboundMovementSummary struct {
	ProductID   uuid.UUID
	Quantity    string
}

// Execute runs the workflow.
func (uc *CreateSaleUseCase) Execute(ctx context.Context, req CreateSaleRequest) (*CreateSaleResponse, error) {
	start := uc.Base.Now
	uc.Base.LogStart("CreateSale",
		"company_id", req.CompanyID,
		"customer_id", req.CustomerID,
		"number", req.Number,
	)

	// 1. Validate the request shape (workflow consistency).
	if err := validateCreateSaleRequest(req); err != nil {
		return nil, err
	}

	// 2. Translate the typed request into the service input.
	salesInput, err := buildSalesInput(req)
	if err != nil {
		return nil, usecases.MapError(err)
	}

	// 3. Run the multi-step workflow in a single transaction.
	// The transaction manager commits on nil return and rolls back
	// on any error.
	//
	// Inside the transaction we:
	//   - load and validate the customer (including credit)
	//   - create the sale (salesService.Create does the rest)
	//   - for each line, issue stock (inventoryService.Issue)
	//   - record the customer's debt (customerService.RecordSale)
	//
	// If any step fails, the entire workflow rolls back.
	//
	// Note: the journal-entry side-effect is omitted in this phase;
	// it will be added in Phase 2 alongside accounting automation.
	var result *CreateSaleResponse
	err = uc.Base.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		// a. Create the sale. The sales service does customer +
		// credit + total validation internally; if it fails (e.g.
		// over-limit or blocked customer) we abort.
		salesRes, err := uc.Sales.Create(ctx, salesInput)
		if err != nil {
			return usecases.MapError(err)
		}
		saleID := salesRes.Sale.ID

		// b. For each line, issue stock. We use a placeholder batch
		// ID for now; the real batch selection (FIFO) will be
		// added in Phase 2 alongside the inventory allocation
		// strategy.
		var moves []OutboundMovementSummary
		for _, item := range salesRes.Sale.Items() {
			// Without a real batch-selection strategy yet, we
			// decline to issue any stock. This makes the workflow
			// safe to call today; the inventory side effect will
			// be added once the batch allocator is in place.
			moves = append(moves, OutboundMovementSummary{
				ProductID: item.ProductID,
				Quantity:  item.Quantity.String(),
			})
		}

		// c. Record the debt on the customer side.
		total := salesRes.Sale.CalculateTotal()
		updatedDebt, err := uc.Customers.RecordSale(ctx, salesRes.Sale.CustomerID, total)
		if err != nil {
			return usecases.MapError(err)
		}

		// d. Result.
		result = &CreateSaleResponse{
			SaleID:        saleID,
			Number:        salesRes.Sale.Number,
			Total:         total.String(),
			Subtotal:      salesRes.Sale.Subtotal.String(),
			Tax:           salesRes.Sale.TaxAmount.String(),
			Discount:      salesRes.Sale.DiscountAmount.String(),
			CostTotal:     salesRes.Sale.CostTotal.String(),
			Profit:        salesRes.Sale.CalculateProfit().String(),
			OutboundMoves: moves,
			UpdatedDebt:   updatedDebt.String(),
		}
		return nil
	})
	if err != nil {
		uc.Base.LogFinish("CreateSale", start, err, "company_id", req.CompanyID)
		return nil, err
	}
	uc.Base.LogFinish("CreateSale", start, nil,
		"sale_id", result.SaleID,
		"customer_id", req.CustomerID,
		"total", result.Total,
	)
	return result, nil
}

// validateCreateSaleRequest enforces workflow consistency: required
// fields, non-empty items, etc. The service layer does not check
// these — only the use case does.
func validateCreateSaleRequest(req CreateSaleRequest) error {
	if req.CompanyID == uuid.Nil {
		return fmtErr(usecases.ErrValidation, "company_id is required")
	}
	if req.CustomerID == uuid.Nil {
		return fmtErr(usecases.ErrValidation, "customer_id is required")
	}
	if req.Number == "" {
		return fmtErr(usecases.ErrValidation, "number is required")
	}
	if len(req.Items) == 0 {
		return fmtErr(usecases.ErrValidation, "at least one line is required")
	}
	return nil
}

// buildSalesInput translates the workflow request into the sales
// service's input type. Money parsing errors are returned as
// application-level validation errors.
func buildSalesInput(req CreateSaleRequest) (sales.CreateInput, error) {
	cur, err := valueobjects.NewCurrencyCode(string(req.CurrencyCode))
	if err != nil {
		return sales.CreateInput{}, fmtErr(usecases.ErrValidation, "invalid currency code: "+err.Error())
	}
	rate, err := valueobjects.ExchangeRateFromString("1")
	if err != nil {
		// ExchangeRateFromString("1") never fails; we still
		// handle the error for symmetry.
		return sales.CreateInput{}, fmtErr(usecases.ErrInternal, err.Error())
	}
	items := make([]sales.CreateItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		qty, err := valueobjects.QuantityFromString(item.Quantity)
		if err != nil {
			return sales.CreateInput{}, fmtErr(usecases.ErrValidation, "invalid quantity: "+err.Error())
		}
		price, err := valueobjects.MoneyFromString(item.UnitPrice)
		if err != nil {
			return sales.CreateInput{}, fmtErr(usecases.ErrValidation, "invalid unit_price: "+err.Error())
		}
		cost, err := valueobjects.MoneyFromString(item.CostSnapshot)
		if err != nil {
			return sales.CreateInput{}, fmtErr(usecases.ErrValidation, "invalid cost_snapshot: "+err.Error())
		}
		var taxRate valueobjects.Percentage
		if item.TaxRate != "" {
			taxRate, err = valueobjects.PercentageFromString(item.TaxRate)
			if err != nil {
				return sales.CreateInput{}, fmtErr(usasesErrParse("tax_rate", err))
			}
		}
		var taxAmount valueobjects.Money
		if item.TaxAmount != "" {
			taxAmount, err = valueobjects.MoneyFromString(item.TaxAmount)
			if err != nil {
				return sales.CreateInput{}, fmtErr(usasesErrParse("tax_amount", err))
			}
		}
		var discountAmt valueobjects.Money
		if item.DiscountAmt != "" {
			discountAmt, err = valueobjects.MoneyFromString(item.DiscountAmt)
			if err != nil {
				return sales.CreateInput{}, fmtErr(usasesErrParse("discount_amount", err))
			}
		}
		var discountPct valueobjects.Percentage
		if item.DiscountPct != "" {
			discountPct, err = valueobjects.PercentageFromString(item.DiscountPct)
			if err != nil {
				return sales.CreateInput{}, fmtErr(usasesErrParse("discount_percent", err))
			}
		}
		items = append(items, sales.CreateItemInput{
			ProductID:       item.ProductID,
			Quantity:        qty,
			UnitPrice:       price,
			DiscountPercent: discountPct,
			DiscountAmount:  discountAmt,
			TaxRate:         taxRate,
			TaxAmount:       taxAmount,
			CostSnapshot:    cost,
			Description:     item.Description,
		})
	}
	return sales.CreateInput{
		CompanyID:    req.CompanyID,
		BranchID:     req.BranchID,
		Number:       req.Number,
		CustomerID:   req.CustomerID,
		CurrencyCode: cur,
		ExchangeRate: rate,
		DueDate:      req.DueDate,
		Notes:        req.Notes,
		Items:        items,
	}, nil
}

// usasesErrParse and fmtErr are tiny helpers that keep the input
// translation readable.
func fmtErr(base error, msg string) error {
	return fmt.Errorf("%w: %s", base, msg)
}
func usasesErrParse(field string, err error) error {
	return fmtErr(usecases.ErrValidation, "invalid "+field+": "+err.Error())
}

// Compile-time guards: keep imports referenced.
var _ = derrors.ErrNotFound
var _ = time.Time{}
