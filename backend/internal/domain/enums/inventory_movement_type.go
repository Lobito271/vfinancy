package enums

// InventoryMovementType classifies a single inventory event.
//
// The signed quantity on the movement is positive for inbound and
// negative for outbound. MovementType is the *reason*; the sign is
// independent and validated by the entity.
type InventoryMovementType string

const (
	MovementTypePurchase      InventoryMovementType = "purchase"
	MovementTypeSale          InventoryMovementType = "sale"
	MovementTypeTransferIn    InventoryMovementType = "transfer_in"
	MovementTypeTransferOut   InventoryMovementType = "transfer_out"
	MovementTypeAdjustmentIn  InventoryMovementType = "adjustment_in"
	MovementTypeAdjustmentOut InventoryMovementType = "adjustment_out"
	MovementTypeReturnIn      InventoryMovementType = "return_in"
	MovementTypeReturnOut     InventoryMovementType = "return_out"
	MovementTypeDamageOut     InventoryMovementType = "damage_out"
)

// IsInbound reports whether the movement increases stock. The caller is
// still expected to use a positive signed quantity.
func (m InventoryMovementType) IsInbound() bool {
	switch m {
	case MovementTypePurchase, MovementTypeTransferIn, MovementTypeAdjustmentIn,
		MovementTypeReturnIn:
		return true
	}
	return false
}

// IsOutbound reports whether the movement decreases stock.
func (m InventoryMovementType) IsOutbound() bool {
	switch m {
	case MovementTypeSale, MovementTypeTransferOut, MovementTypeAdjustmentOut,
		MovementTypeReturnOut, MovementTypeDamageOut:
		return true
	}
	return false
}

func (m InventoryMovementType) Valid() bool {
	switch m {
	case MovementTypePurchase, MovementTypeSale,
		MovementTypeTransferIn, MovementTypeTransferOut,
		MovementTypeAdjustmentIn, MovementTypeAdjustmentOut,
		MovementTypeReturnIn, MovementTypeReturnOut,
		MovementTypeDamageOut:
		return true
	}
	return false
}

func (m InventoryMovementType) String() string { return string(m) }
