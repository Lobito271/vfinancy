package services_test

import (
	"context"

	"vfinancy/backend/internal/application/services"
)

// fakeTxManager is an in-memory implementation of services.TxManager
// for unit tests. It simply invokes the callback in the same goroutine
// without touching a database. Each test gets its own fakeTxManager
// so they cannot accidentally share state.
type fakeTxManager struct {
	commitCount   int
	rollbackCount int
}

func newFakeTxManager() *fakeTxManager {
	return &fakeTxManager{}
}

// WithinTransaction runs fn immediately. If fn returns nil,
// commitCount is incremented. If it returns an error, rollbackCount
// is and the error is propagated.
func (f *fakeTxManager) WithinTransaction(ctx context.Context, fn services.TxRunner) error {
	if err := fn(ctx); err != nil {
		f.rollbackCount++
		return err
	}
	f.commitCount++
	return nil
}
