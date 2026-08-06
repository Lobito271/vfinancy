package accounting

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// ChartOfAccount is a single account in the company's chart of
// accounts. The path is a dotted notation (e.g. "1.1.01") and the
// depth is the number of levels.
type ChartOfAccount struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Code        valueobjects.ChartOfAccountsCode
	Name        string
	Type        enums.AccountType
	ParentID    *uuid.UUID
	Path        string
	Depth       int
	IsActive    bool
	AllowsMovement bool
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// NewChartOfAccountOptions is the input to NewChartOfAccount.
type NewChartOfAccountOptions struct {
	CompanyID   uuid.UUID
	Code        valueobjects.ChartOfAccountsCode
	Name        string
	Type        enums.AccountType
	ParentID    *uuid.UUID
	Path        string
	Depth       int
	AllowsMovement bool
	Description string
}

// NewChartOfAccount validates and constructs a chart-of-accounts row.
func NewChartOfAccount(now time.Time, opts NewChartOfAccountOptions) (*ChartOfAccount, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("company id is required"))
	}
	if !opts.Type.Valid() {
		return nil, derrors.Wrap(derrors.ErrInvalidEnum, errField("account type is invalid"))
	}
	if opts.Name == "" {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("account name is required"))
	}
	if opts.Depth < 1 {
		return nil, derrors.Wrap(derrors.ErrOutOfRange, errField("depth must be >= 1"))
	}
	return &ChartOfAccount{
		ID:             uuid.New(),
		CompanyID:      opts.CompanyID,
		Code:           opts.Code,
		Name:           opts.Name,
		Type:           opts.Type,
		ParentID:       opts.ParentID,
		Path:           opts.Path,
		Depth:          opts.Depth,
		IsActive:       true,
		AllowsMovement: opts.AllowsMovement,
		Description:    opts.Description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// NormalBalance returns whether increases to this account are debits
// (assets, expenses) or credits (liabilities, equity, income).
func (a *ChartOfAccount) NormalBalance() enums.NormalBalance {
	return a.Type.NormalBalance()
}

// IsDebitNormal / IsCreditNormal are convenience predicates.
func (a *ChartOfAccount) IsDebitNormal() bool  { return a.NormalBalance() == enums.DebitNormal }
func (a *ChartOfAccount) IsCreditNormal() bool { return a.NormalBalance() == enums.CreditNormal }

// Activate / Deactivate.
func (a *ChartOfAccount) Activate()   { a.IsActive = true }
func (a *ChartOfAccount) Deactivate() { a.IsActive = false }
