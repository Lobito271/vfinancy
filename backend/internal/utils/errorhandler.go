package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	derrors "vfinancy/backend/internal/domain/errors"
)

const errorLogFile = "app_errors_log.json"

var (
	logFileOnce sync.Once
	logFile     *os.File
	logFileMu   sync.Mutex
	logFileErr  error
)

type errorLogEntry struct {
	Timestamp string `json:"timestamp"`
	Error     string `json:"error"`
}

func openLogFile() {
	logFileOnce.Do(func() {
		f, err := os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logFileErr = err
			return
		}
		logFile = f
	})
}

func appendErrorLog(entry errorLogEntry) {
	openLogFile()
	if logFile == nil {
		return
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	data = append(data, '\n')
	logFileMu.Lock()
	defer logFileMu.Unlock()
	_, _ = logFile.Write(data)
}

type errorMapping struct {
	contains string
	message  string
}

var errorMappings = []errorMapping{
	{contains: "UNIQUE constraint failed", message: "Este registro ya existe en el sistema."},
	{contains: "FOREIGN KEY constraint failed", message: "No se puede eliminar este registro porque está en uso por otro módulo."},
	{contains: "syntax error", message: "El texto ingresado contiene caracteres no válidos para la búsqueda."},
	{contains: "unrecognized token", message: "El texto ingresado contiene caracteres no válidos para la búsqueda."},
	{contains: "invalid input syntax", message: "El texto ingresado contiene caracteres no válidos para la búsqueda."},
	{contains: "database is locked", message: "El sistema está ocupado. Intente nuevamente en un instante."},
	{contains: "disk I/O error", message: "No se pudo escribir en el disco. Verifique el espacio disponible."},
	{contains: "no space left", message: "No se pudo escribir en el disco. Verifique el espacio disponible."},
	{contains: "connection refused", message: "No se pudo conectar con la base de datos."},
	{contains: "dial tcp", message: "No se pudo conectar con la base de datos."},
	{contains: "no rows in result set", message: "No se encontró el registro solicitado."},
	{contains: "context deadline exceeded", message: "La operación tardó demasiado. Intente nuevamente."},
	{contains: "timeout", message: "La operación tardó demasiado. Intente nuevamente."},
	{contains: "no such table", message: "Ocurrió un error interno. Por favor, revise el archivo de registro."},
	{contains: "no such column", message: "Ocurrió un error interno. Por favor, revise el archivo de registro."},
}

const fallbackMessage = "Ocurrió un error interno. Por favor, revise el archivo de registro."

func mapError(raw string) string {
	lower := strings.ToLower(raw)
	for _, m := range errorMappings {
		if strings.Contains(lower, strings.ToLower(m.contains)) {
			return m.message
		}
	}
	return fallbackMessage
}

var knownDomainCodes = []string{
	"VALIDATION", "NOT_FOUND", "CONFLICT", "REQUIRED", "INVALID_FORMAT",
	"DUPLICATE", "INSUFFICIENT_STOCK", "INVALID_PAYMENT", "NEGATIVE_QUANTITY",
	"NEGATIVE_MONEY", "PURCHASE_CANCELLED", "SALE_ALREADY_PAID",
	"ALREADY_RECEIVED", "ALREADY_RECONCILED", "CUSTOMER_INACTIVE",
	"SUPPLIER_INACTIVE", "OUT_OF_RANGE",
}

func isDomainError(err error) bool {
	for _, code := range knownDomainCodes {
		if derrors.IsCode(err, code) {
			return true
		}
	}
	return false
}

func ProcessError(err error) error {
	if err == nil {
		return nil
	}
	if isDomainError(err) {
		return err
	}
	raw := err.Error()
	log.Printf("[ERROR] %s", raw)
	appendErrorLog(errorLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Error:     raw,
	})
	return errors.New(mapError(raw))
}
