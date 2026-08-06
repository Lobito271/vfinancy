package enums

// PeriodStatus is the lifecycle state of a fiscal period.
type PeriodStatus string

const (
	// PeriodStatusOpen — entries can be posted freely.
	PeriodStatusOpen PeriodStatus = "open"
	// PeriodStatusClosing — closing entries are being recorded; no new
	// regular entries accepted.
	PeriodStatusClosing PeriodStatus = "closing"
	// PeriodStatusClosed — no further entries; only reversing entries
	// referencing a posted entry in this period.
	PeriodStatusClosed PeriodStatus = "closed"
)

func (p PeriodStatus) Valid() bool {
	switch p {
	case PeriodStatusOpen, PeriodStatusClosing, PeriodStatusClosed:
		return true
	}
	return false
}

func (p PeriodStatus) AcceptsNewEntries() bool {
	return p == PeriodStatusOpen
}

func (p PeriodStatus) String() string { return string(p) }
