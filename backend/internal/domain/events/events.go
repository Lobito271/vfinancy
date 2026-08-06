// Package events defines the event types emitted by the domain layer.
//
// Events are value objects: they are produced by entity methods and
// collected by an event bus (not implemented in this phase) which lives
// in the application layer. The domain layer never imports the event
// bus; it only knows how to construct events.
package events

import (
	"time"

	"github.com/google/uuid"
)

// EventType is the type discriminator on the wire.
type EventType string

const (
	EventTypeCustomerActivated        EventType = "customer.activated"
	EventTypeCustomerDeactivated      EventType = "customer.deactivated"
	EventTypeCustomerBlocked          EventType = "customer.blocked"

	EventTypeSaleCreated              EventType = "sale.created"
	EventTypeSalePaid                 EventType = "sale.paid"
	EventTypeSaleCancelled            EventType = "sale.cancelled"

	EventTypePurchaseCreated          EventType = "purchase.created"
	EventTypePurchaseReceived         EventType = "purchase.received"
	EventTypePurchasePaid             EventType = "purchase.paid"
	EventTypePurchaseCancelled        EventType = "purchase.cancelled"

	EventTypeInventoryBelowMinimum    EventType = "inventory.below_minimum"
	EventTypeInventoryClearance       EventType = "inventory.clearance"
	EventTypeInventoryOutOfStock      EventType = "inventory.out_of_stock"

	EventTypeJournalEntryPosted       EventType = "journal_entry.posted"
	EventTypeJournalEntryReversed     EventType = "journal_entry.reversed"
	EventTypeFiscalPeriodClosed       EventType = "fiscal_period.closed"
)

// Event is a single domain event. The OccurredAt is set by the entity
// at the moment of the state change; ID is generated at the same time.
type Event struct {
	ID         uuid.UUID
	Type       EventType
	OccurredAt time.Time
	// AggregateID identifies the entity that produced the event
	// (e.g. the Sale that was cancelled).
	AggregateID uuid.UUID
	// Payload is free-form, domain-defined per event type. Encoding
	// is the responsibility of the application layer (typically JSON).
	Payload map[string]any
}

// NewEvent constructs an event with a freshly generated ID and the
// current timestamp. The caller supplies the event type, the aggregate
// it relates to, and the payload.
func NewEvent(t EventType, aggregateID uuid.UUID, payload map[string]any) Event {
	return Event{
		ID:          uuid.New(),
		Type:        t,
		OccurredAt:  time.Now().UTC(),
		AggregateID: aggregateID,
		Payload:     payload,
	}
}
