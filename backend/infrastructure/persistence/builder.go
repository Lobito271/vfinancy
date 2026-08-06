package persistence

import "strings"

// joinClauses joins WHERE clauses with AND. The slice must be
// non-empty; callers handle the empty case (no filters) at the
// construction site.
func JoinClauses(clauses []string) string {
	return strings.Join(clauses, " AND ")
}

// inColumn returns "column IN (?, ?, ?, ...)" with the right number
// of placeholders for the given list. Returns "column IN (NULL)"
// for an empty list (always false).
func InColumn(column string, n int) string {
	return column + " IN (" + InClause(n) + ")"
}
