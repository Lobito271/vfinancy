package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	derrors "vfinancy/backend/internal/domain/errors"
)

// Logger is the minimal logger interface used by use cases. The
// service layer's *common.Logger satisfies this; tests use a
// *slog.Logger writing to io.Discard.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// TxRunner is the callback signature inside WithinTransaction. The
// application-side TxManager (Tx) provides the same shape.
type TxRunner func(ctx context.Context) error

// Tx is the application-side transaction interface. The service
// layer's services.TxManager satisfies this.
type Tx interface {
	WithinTransaction(ctx context.Context, fn TxRunner) error
}

// Base is a small embeddable struct that gives every use case
// access to the transaction manager and the logger. Use cases embed
// it (composition over inheritance).
type Base struct {
	Tx   Tx
	Log  Logger
	Now  func() time.Time // injectable for tests; defaults to time.Now
}

// NewBase returns a Base with default Now set to time.Now.
func NewBase(tx Tx, log Logger) Base {
	return Base{Tx: tx, Log: log, Now: func() time.Time { return time.Now().UTC() }}
}

// With returns a copy of b with a new Now function. Useful in tests.
func (b Base) With(now func() time.Time) Base {
	b.Now = now
	return b
}

// LogStart logs the start of a workflow.
func (b Base) LogStart(op string, keysAndValues ...any) {
	if b.Log == nil {
		return
	}
	args := append([]any{"op", op, "at", b.Now()}, keysAndValues...)
	b.Log.Info(op+" started", args...)
}

// LogFinish logs the end of a workflow with its duration.
func (b Base) LogFinish(op string, start time.Time, err error, keysAndValues ...any) {
	if b.Log == nil {
		return
	}
	args := append([]any{
		"op", op,
		"at", b.Now(),
		"duration_ms", b.Now().Sub(start).Milliseconds(),
	}, keysAndValues...)
	if err != nil {
		args = append(args, "error", err.Error())
		b.Log.Error(op+" failed", args...)
	} else {
		b.Log.Info(op+" completed", args...)
	}
}

// WrapAsInternal wraps a low-level error as an internal use case
// error. The original is preserved via Wrap so callers can errors.As
// the underlying cause.
func WrapAsInternal(err error) error {
	if err == nil {
		return nil
	}
	return derrors.Wrap(ErrInternal, err)
}

// MapError translates service/domain errors into the appropriate
// application-level sentinel. The use case's Execute method is
// expected to call this on the final return value of the workflow
// before returning to the caller.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) || derrors.IsCode(err, "NOT_FOUND") {
		return ErrNotFound
	}
	if derrors.IsCode(err, "DUPLICATE") {
		return fmt.Errorf("%w: %s", ErrConflict, err.Error())
	}
	if derrors.IsCode(err, "INSUFFICIENT_STOCK") ||
		derrors.IsCode(err, "INVALID_PAYMENT") ||
		derrors.IsCode(err, "NEGATIVE_QUANTITY") ||
		derrors.IsCode(err, "NEGATIVE_MONEY") ||
		derrors.IsCode(err, "PURCHASE_CANCELLED") ||
		derrors.IsCode(err, "SALE_ALREADY_PAID") ||
		derrors.IsCode(err, "ALREADY_RECEIVED") ||
		derrors.IsCode(err, "ALREADY_RECONCILED") {
		return fmt.Errorf("%w: %s", ErrConflict, err.Error())
	}
	if derrors.IsCode(err, "REQUIRED") || derrors.IsCode(err, "INVALID_FORMAT") {
		return fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}
	return WrapAsInternal(err)
}
