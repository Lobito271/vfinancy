package enums

// BatchStatus is the lifecycle state of an inventory batch.
type BatchStatus string

const (
	// BatchStatusActive — batch is available for sale; quantity > 0.
	BatchStatusActive BatchStatus = "active"
	// BatchStatusDepleted — quantity = 0; historical record.
	BatchStatusDepleted BatchStatus = "depleted"
	// BatchStatusVoided — receipt cancelled by the operator; the row
	// is kept for audit and rejects every further stock change.
	BatchStatusVoided BatchStatus = "voided"
)

func (b BatchStatus) Valid() bool {
	switch b {
	case BatchStatusActive, BatchStatusDepleted, BatchStatusVoided:
		return true
	}
	return false
}

func (b BatchStatus) String() string { return string(b) }
