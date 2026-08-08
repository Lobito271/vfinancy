// Package logger is a thin wrapper over log/slog from the standard
// library. It exposes only what the app needs: one constructor from
// config (New), one from an existing *slog.Logger (NewLogger), and
// With/WithContext. It lives in infrastructure (not internal) because
// the root package must be able to import it.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger is the application logger. It is a value type around
// *slog.Logger that gives services a stable, mockable surface while
// staying cheap.
type Logger struct {
	*slog.Logger
}

// New builds a logger from config values.
func New(level, format, output string) *Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	var handler slog.Handler
	switch strings.ToLower(format) {
	case "text":
		handler = slog.NewTextHandler(resolveWriter(os.Stdout, output), opts)
	default:
		handler = slog.NewJSONHandler(resolveWriter(os.Stdout, output), opts)
	}
	return &Logger{Logger: slog.New(handler)}
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

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func resolveWriter(defaultW io.Writer, output string) io.Writer {
	switch strings.ToLower(output) {
	case "stderr":
		return os.Stderr
	case "stdout", "":
		return defaultW
	default:
		if f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			return f
		}
		return defaultW
	}
}
