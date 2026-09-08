package repositories

import (
	"time"

	"github.com/google/uuid"
)

// PageRequest carries the page size and offset for a paginated query.
// A zero PageRequest means "no pagination, return everything up to a
// sensible internal cap". Use MaxPageSize on the repository side to
// clamp the limit.
type PageRequest struct {
	Limit  int
	Offset int
}

// Page is the paginated result envelope. Items is always non-nil.
type Page[T any] struct {
	Items []T
	Total int
	Limit int
	Offset int
}

// HasMore reports whether there is at least one more page after the
// current one.
func (p Page[T]) HasMore() bool {
	return p.Offset+len(p.Items) < p.Total
}

// SortDirection is the order direction for a sort.
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// Sort is a single sort directive. Repositories that support sorting
// accept []Sort in their filter types.
type Sort struct {
	Field     string
	Direction SortDirection
}

// TimeRange is an inclusive [From, To] range filter. Either end may
// be the zero value to mean "unbounded".
type TimeRange struct {
	From time.Time
	To   time.Time
}

// IsZero reports whether the range is empty (no constraint).
func (r TimeRange) IsZero() bool {
	return r.From.IsZero() && r.To.IsZero()
}

// UUIDSlice is a small helper to make filter structs more readable
// than []uuid.UUID.
type UUIDSlice []uuid.UUID
