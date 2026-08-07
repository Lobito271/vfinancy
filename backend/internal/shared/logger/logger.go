// Package logger re-exports the shared application logger. Services
// import it from here so infrastructure stays out of their import
// graph; the implementation lives in infrastructure/logger because the
// root package must be able to import it.
package logger

import (
	"log/slog"

	ilogger "vfinancy/backend/infrastructure/logger"
)

type Logger = ilogger.Logger

func New(level, format, output string) *Logger { return ilogger.New(level, format, output) }

func NewLogger(l *slog.Logger) *Logger { return ilogger.NewLogger(l) }
