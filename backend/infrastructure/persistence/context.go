package persistence

import "context"

type ctxKey int

const txQuerierKey ctxKey = iota

// WithTxQuerier returns a context carrying the given transaction-bound
// Querier. The TxManager stamps the context this way so repositories
// inside a workflow run their SQL on the active transaction.
func WithTxQuerier(parent context.Context, q Querier) context.Context {
	return context.WithValue(parent, txQuerierKey, q)
}

// TxQuerierFromContext returns the transaction-bound Querier stored in
// the context, or nil if none is present.
func TxQuerierFromContext(ctx context.Context) Querier {
	if v, ok := ctx.Value(txQuerierKey).(Querier); ok && v != nil {
		return v
	}
	return nil
}

// Q returns the transaction-bound Querier when the context carries one,
// otherwise it returns the fallback (auto-commit) Querier.
//
// Every repository method must route its SQL through Q(ctx, r.q) so a
// workflow running inside a transaction is never written to the
// connection pool by accident.
func Q(ctx context.Context, fallback Querier) Querier {
	if t := TxQuerierFromContext(ctx); t != nil {
		return t
	}
	return fallback
}
