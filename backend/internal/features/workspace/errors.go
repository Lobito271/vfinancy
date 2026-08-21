package workspace

import "errors"

var (
	ErrProfileNotFound = errors.New("workspace: local profile not found")
	ErrInvalidProfile  = errors.New("workspace: invalid local profile")
	ErrInvalidCompany  = errors.New("workspace: invalid company")
	ErrCompanyRequired = errors.New("workspace: active company is required")
	ErrCompanyInactive = errors.New("workspace: company is inactive")
	ErrProfileLocked   = errors.New("workspace: local profile is locked")
	ErrPasswordWrong   = errors.New("workspace: invalid local password")
	ErrPasswordWeak    = errors.New("workspace: local password is too weak")
)
