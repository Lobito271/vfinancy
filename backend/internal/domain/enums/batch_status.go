package enums

// BatchStatus is the lifecycle state of an inventory batch.
type BatchStatus string

const (
	// BatchStatusActive — batch is available for sale; quantity > 0.
	BatchStatusActive BatchStatus = "active"
	// BatchStatusDepleted — quantity = 0; historical record.
	BatchStatusDepleted BatchStatus = "depleted"
	// BatchStatusWrittenOff — quantity discarded (damage, expiry);
	// generates an adjustment movement.
	BatchStatusWrittenOff BatchStatus = "written_off"
)

func (b BatchStatus) Valid() bool {
	switch b {
	case BatchStatusActive, BatchStatusDepleted, BatchStatusWrittenOff:
		return true
	}
	return false
}

func (b BatchStatus) String() string { return string(b) }
