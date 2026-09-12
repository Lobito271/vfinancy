package enums

// InventoryMovementType classifies a single inventory event.
//
// The signed quantity on the movement is positive for inbound and
// negative for outbound. The type is the reason; the sign is
// independent and validated by the entity.
type InventoryMovementType string

const (
	MovementTypePurchaseReceipt InventoryMovementType = "purchase_receipt"
	MovementTypeSale            InventoryMovementType = "sale"
	MovementTypeVoidSale        InventoryMovementType = "void_sale"
	MovementTypeVoidPurchase    InventoryMovementType = "void_purchase"
	MovementTypeAdjustmentIn    InventoryMovementType = "adjustment_in"
	MovementTypeAdjustmentOut   InventoryMovementType = "adjustment_out"
)

// IsInbound reports whether the movement increases stock. The caller is
// still expected to use a positive signed quantity.
func (m InventoryMovementType) IsInbound() bool {
	switch m {
	case MovementTypePurchaseReceipt, MovementTypeVoidSale, MovementTypeAdjustmentIn:
		return true
	}
	return false
}

// IsOutbound reports whether the movement decreases stock.
func (m InventoryMovementType) IsOutbound() bool {
	switch m {
	case MovementTypeSale, MovementTypeVoidPurchase, MovementTypeAdjustmentOut:
		return true
	}
	return false
}

func (m InventoryMovementType) Valid() bool {
	switch m {
	case MovementTypePurchaseReceipt, MovementTypeSale, MovementTypeVoidSale,
		MovementTypeVoidPurchase, MovementTypeAdjustmentIn, MovementTypeAdjustmentOut:
		return true
	}
	return false
}

func (m InventoryMovementType) String() string { return string(m) }
