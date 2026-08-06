package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New(level, format, output string) *Logger {
	return NewWith(os.Stdout, level, format, output)
}

func NewWith(w io.Writer, level, format, output string) *Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "text":
		handler = slog.NewTextHandler(resolveWriter(w, output), opts)
	default:
		handler = slog.NewJSONHandler(resolveWriter(w, output), opts)
	}

	return &Logger{Logger: slog.New(handler)}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	if reqID, ok := ctx.Value(requestIDKey{}).(string); ok && reqID != "" {
		return l.With("request_id", reqID)
	}
	return l
}

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

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}
