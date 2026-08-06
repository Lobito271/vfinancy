// Package common holds shared helpers for the service layer. The
// logger is intentionally minimal: a thin wrapper over log/slog
// from the standard library. It exposes only the methods the
// services need (Info, Error, With).
package logger

import (
	"context"
	"log/slog"
)

// Logger is the application-layer logger. It is a value type around
// *slog.Logger that gives the services a stable, mockable surface
// while staying cheap.
type Logger struct {
	*slog.Logger
}

// NewLogger builds a logger from a *slog.Logger. The default text
// handler is used; production code wires JSON via slog.NewJSONHandler.
func NewLogger(l *slog.Logger) *Logger {
	if l == nil {
		l = slog.Default()
	}
	return &Logger{Logger: l}
}

// With returns a new logger with the given key/value pairs attached.
// The application layer never calls this directly; it is exposed
// for tests.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithContext returns a logger that includes the request id from ctx
// in every line. The base implementation is a no-op; the application
// container may install a richer one.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if id, ok := ctx.Value(reqIDKey{}).(string); ok && id != "" {
		return l.With("request_id", id)
	}
	return l
}

type reqIDKey struct{}
