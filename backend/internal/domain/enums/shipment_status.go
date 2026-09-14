package enums

// ShipmentStatus is the lifecycle state of a shipment:
// pending → shipped → delivered.
type ShipmentStatus string

const (
	// ShipmentStatusPending — created, not yet dispatched.
	ShipmentStatusPending ShipmentStatus = "pending"
	// ShipmentStatusShipped — handed to the carrier, in transit.
	ShipmentStatusShipped ShipmentStatus = "shipped"
	// ShipmentStatusDelivered — received by the customer.
	ShipmentStatusDelivered ShipmentStatus = "delivered"
)

// AllShipmentStatuses returns every valid shipment status.
func AllShipmentStatuses() []ShipmentStatus {
	return []ShipmentStatus{
		ShipmentStatusPending,
		ShipmentStatusShipped,
		ShipmentStatusDelivered,
	}
}

// Valid reports whether the status is a known member.
func (s ShipmentStatus) Valid() bool {
	switch s {
	case ShipmentStatusPending, ShipmentStatusShipped, ShipmentStatusDelivered:
		return true
	}
	return false
}

// String returns the canonical string form.
func (s ShipmentStatus) String() string { return string(s) }