package enums

// JournalStatus is the lifecycle state of a journal entry. Entries are
// append-only: once posted, corrections require a reversing entry.
type JournalStatus string

const (
	// JournalStatusDraft — entry is being built; lines can be added,
	// removed, edited.
	JournalStatusDraft JournalStatus = "draft"
	// JournalStatusPosted — entry is committed to the books. No further
	// mutations are allowed; use a reversing entry to correct.
	JournalStatusPosted JournalStatus = "posted"
	// JournalStatusReversed — a reversing entry has been posted against
	// this entry; this entry is now informational only.
	JournalStatusReversed JournalStatus = "reversed"
)

func (j JournalStatus) Valid() bool {
	switch j {
	case JournalStatusDraft, JournalStatusPosted, JournalStatusReversed:
		return true
	}
	return false
}

func (j JournalStatus) IsEditable() bool {
	return j == JournalStatusDraft
}

func (j JournalStatus) String() string { return string(j) }
