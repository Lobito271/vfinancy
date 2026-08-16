package inventory

import (
	derrors "vfinancy/backend/internal/domain/errors"
)

// ErrBatchAlreadyVoided indicates the batch was already cancelled.
var ErrBatchAlreadyVoided = derrors.New("BATCH_ALREADY_VOIDED", "batch is already voided")

// errField is a small helper that returns a context-tagged error.
func errField(s string) error { return fieldErr(s) }

type fieldErr string

func (e fieldErr) Error() string { return string(e) }
