// Package errors defines the typed errors used across the domain layer.
//
// Domain errors are explicit and named so the application layer can map
// them to UI messages, API responses, or retry policies without inspecting
// message strings.
package errors

import (
	"errors"
	"fmt"
)

// DomainError is the base interface for all domain-layer errors.
// It exposes a stable Code() identifier and supports Unwrap() so the
// application layer can use errors.Is / errors.As transparently.
type DomainError interface {
	error
	Code() string
	Unwrap() error
}

// base is the common implementation. Concrete error types in other
// files embed it (or compose it) to inherit Code() and Unwrap().
type base struct {
	code    string
	message string
	cause   error
}

func (b *base) Error() string {
	if b.cause != nil {
		return fmt.Sprintf("%s: %s: %v", b.code, b.message, b.cause)
	}
	return fmt.Sprintf("%s: %s", b.code, b.message)
}

func (b *base) Code() string     { return b.code }
func (b *base) Unwrap() error   { return b.cause }

// New constructs a DomainError with the given code and message.
// Use this from concrete error helpers in this package.
func New(code, message string) DomainError {
	return &base{code: code, message: message}
}

// Wrap attaches a cause to an existing DomainError while preserving
// the same code. Useful when bubbling up a low-level error.
func Wrap(err DomainError, cause error) DomainError {
	return &base{code: err.Code(), message: err.Error(), cause: cause}
}

// IsCode reports whether err (or anything in its Unwrap chain) is a
// DomainError with the given code.
func IsCode(err error, code string) bool {
	var de DomainError
	if errors.As(err, &de) {
		return de.Code() == code
	}
	return false
}

// IsAnyCode reports whether err matches any of the given codes.
func IsAnyCode(err error, codes ...string) bool {
	for _, c := range codes {
		if IsCode(err, c) {
			return true
		}
	}
	return false
}

type FieldErr string

func (e FieldErr) Error() string { return string(e) }

func ErrField(s string) error { return FieldErr(s) }
