package inventory

import derrors "vfinancy/backend/internal/domain/errors"

var ErrBatchAlreadyVoided = derrors.New("BATCH_ALREADY_VOIDED", "batch is already voided")

var errField = derrors.ErrField
