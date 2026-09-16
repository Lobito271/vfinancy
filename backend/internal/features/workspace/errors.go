package workspace

import "errors"

var (
	ErrProfileNotFound         = errors.New("workspace: local profile not found")
	ErrInvalidProfile          = errors.New("workspace: invalid local profile")
	ErrProfileExists           = errors.New("workspace: local profile already exists")
	ErrProfileLocked           = errors.New("workspace: local profile is locked")
	ErrPasswordWrong           = errors.New("workspace: invalid local password")
	ErrPasswordWeak            = errors.New("workspace: local password is too weak")
	ErrSecurityQuestionRequired = errors.New("workspace: no security question configured")
	ErrSecurityAnswerInvalid   = errors.New("workspace: invalid security answer")
	ErrSecurityAnswerTooShort  = errors.New("workspace: security answer is too short")
)
