package purchasing

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// PurchaseFilter is the input to PurchaseRepository.List.
type PurchaseFilter struct {
	Search         string
	Status         string
	CreditCardID   *uuid.UUID
	From           *time.Time
	To             *time.Time
	IncludeDeleted bool
	repositories.PageRequest
}

// PurchaseLineSummary is one order line rendered in the list's
// "Productos" column: a decimal quantity, a product name and the
// ordered / sold quantities used to flag fully sold orders.
type PurchaseLineSummary struct {
	Quantity string
	Name     string
	Ordered  string
	Sold     string
}

// parseDecimal reads a stored decimal string, defaulting to zero.
func parseDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(strings.TrimSpace(s))
	if err != nil {
		return decimal.Zero
	}
	return d
}

// SummarizeLines renders the first line of an order as the "3× Mouse"
// summary used by the list's products column. Only the first product is
// shown there; the full line detail lives in the order drawer.
func SummarizeLines(lines []PurchaseLineSummary) string {
	if len(lines) == 0 {
		return ""
	}
	q := parseDecimal(lines[0].Quantity)
	if q.Equal(q.Truncate(0)) {
		q = q.Truncate(0)
	}
	return q.String() + "× " + lines[0].Name
}

// AllLinesSold reports whether every unit of every line of an order has
// been sold. An order without lines is never considered sold.
func AllLinesSold(lines []PurchaseLineSummary) bool {
	if len(lines) == 0 {
		return false
	}
	for _, l := range lines {
		if parseDecimal(l.Sold).LessThan(parseDecimal(l.Ordered)) {
			return false
		}
	}
	return true
}

// PurchaseRepository persists purchase orders and their line items.
type PurchaseRepository interface {
	// Create inserts the order and all of its items.
	Create(ctx context.Context, po *PurchaseOrder, items []*PurchaseOrderItem) error
	// Update persists the mutable order fields.
	Update(ctx context.Context, po *PurchaseOrder) error
	// SoftDelete marks the order as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// GetByID loads a single order without its items.
	GetByID(ctx context.Context, id uuid.UUID) (*PurchaseOrder, error)
	// ListItems returns the lines of an order ordered by line number.
	ListItems(ctx context.Context, purchaseOrderID uuid.UUID) ([]*PurchaseOrderItem, error)
	// UpdateItemReceipt records the received quantity of a line.
	UpdateItemReceipt(ctx context.Context, itemID uuid.UUID, received valueobjects.Quantity) error
	// NextNumber returns the next "PO-" zero-padded sequence number.
	NextNumber(ctx context.Context) (string, error)
	// List returns the orders matching the filter.
	List(ctx context.Context, filter PurchaseFilter) (repositories.Page[*PurchaseOrder], error)
	// ListLineSummaries returns the product lines of the given orders,
	// keyed by order id and ordered by line number, for the list's
	// products column.
	ListLineSummaries(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]PurchaseLineSummary, error)
	// ListExtraCosts returns the extra costs of an order, newest first.
	ListExtraCosts(ctx context.Context, purchaseOrderID uuid.UUID) ([]*ExtraCost, error)
	// CreateExtraCost inserts one extra cost against an order.
	CreateExtraCost(ctx context.Context, ec *ExtraCost) error
	// UpdateExtraCost persists the mutable fields of an extra cost
	// scoped to its parent order.
	UpdateExtraCost(ctx context.Context, ec *ExtraCost) error
	// DeleteExtraCost removes one extra cost scoped to its parent order.
	DeleteExtraCost(ctx context.Context, id, purchaseOrderID uuid.UUID) error
}
