package bindings

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/accounting"
)

// AccountDTO is the serializable view of a chart-of-accounts entry.
type AccountDTO struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ParentID      string `json:"parentId"`
	Path          string `json:"path"`
	Depth         int    `json:"depth"`
	IsActive      bool   `json:"isActive"`
	AllowsMovement bool  `json:"allowsMovement"`
	Description   string `json:"description"`
}

// JournalEntryLineDTO is a serializable journal line.
type JournalEntryLineDTO struct {
	ID           string `json:"id"`
	LineNumber   int    `json:"lineNumber"`
	AccountID    string `json:"accountId"`
	Description  string `json:"description"`
	Debit        string `json:"debit"`
	Credit       string `json:"credit"`
}

// JournalEntryDTO is the serializable view of a journal entry.
type JournalEntryDTO struct {
	ID          string                 `json:"id"`
	Number      string                 `json:"number"`
	EntryDate   string                 `json:"entryDate"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	SourceID    string                 `json:"sourceId"`
	Status      string                 `json:"status"`
	Lines       []*JournalEntryLineDTO `json:"lines"`
	CreatedAt   string                 `json:"createdAt"`
}

// FiscalPeriodDTO is the serializable view of a fiscal period.
type FiscalPeriodDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PeriodStart string `json:"periodStart"`
	PeriodEnd   string `json:"periodEnd"`
	Status      string `json:"status"`
}

func toFiscalPeriodDTO(p *accounting.FiscalPeriod) *FiscalPeriodDTO {
	return &FiscalPeriodDTO{
		ID:          p.ID.String(),
		Name:        p.Name,
		PeriodStart: p.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   p.PeriodEnd.Format("2006-01-02"),
		Status:      p.Status,
	}
}

func toAccountDTO(a *accounting.ChartOfAccount) *AccountDTO {
	parentID := ""
	if a.ParentID != nil {
		parentID = a.ParentID.String()
	}
	return &AccountDTO{
		ID:             a.ID.String(),
		Code:           a.Code.String(),
		Name:           a.Name,
		Type:           a.Type.String(),
		ParentID:       parentID,
		Path:           a.Path,
		Depth:          a.Depth,
		IsActive:       a.IsActive,
		AllowsMovement: a.AllowsMovement,
		Description:    a.Description,
	}
}

func toJournalEntryDTO(e *accounting.JournalEntry) *JournalEntryDTO {
	sourceID := ""
	if e.SourceID != nil {
		sourceID = e.SourceID.String()
	}
	lines := make([]*JournalEntryLineDTO, 0, len(e.Lines))
	for _, l := range e.Lines {
		lines = append(lines, &JournalEntryLineDTO{
			ID:          l.ID.String(),
			LineNumber:  l.LineNumber,
			AccountID:   l.AccountID.String(),
			Description: l.Description,
			Debit:       l.Debit.String(),
			Credit:      l.Credit.String(),
		})
	}
	return &JournalEntryDTO{
		ID:          e.ID.String(),
		Number:      e.Number,
		EntryDate:   e.EntryDate.Format("2006-01-02"),
		Description: e.Description,
		Source:      e.Source.String(),
		SourceID:    sourceID,
		Status:      e.Status.String(),
		Lines:       lines,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
	}
}

// ListChartOfAccounts returns the full chart of accounts.
func (a *App) ListChartOfAccounts() ([]*AccountDTO, error) {
	ctx := a.Context()
	accounts, err := a.accountingSvc.ListChartOfAccounts(ctx, demoCompanyID)
	if err != nil {
		return nil, err
	}
	items := make([]*AccountDTO, 0, len(accounts))
	for _, acc := range accounts {
		items = append(items, toAccountDTO(acc))
	}
	return items, nil
}

// CreateChartOfAccountRequest creates a chart-of-accounts entry.
type CreateChartOfAccountRequest struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	ParentID       string `json:"parentId"`
	AllowsMovement bool   `json:"allowsMovement"`
	Description    string `json:"description"`
}

// CreateChartOfAccount persists a new account, computing its path and
// depth from the parent (if any).
func (a *App) CreateChartOfAccount(req CreateChartOfAccountRequest) (*AccountDTO, error) {
	ctx := a.Context()
	code, err := valueobjects.NewChartOfAccountsCode(req.Code)
	if err != nil {
		return nil, err
	}
	parent, err := a.resolveChartParent(ctx, req.ParentID)
	if err != nil {
		return nil, err
	}
	var pid *uuid.UUID
	if parent != nil {
		pid = &parent.ID
	}
	depth := code.Depth()
	path := code.String()
	if parent != nil {
		depth = parent.Depth + 1
		path = accountPath(parent, code.String())
	}
	acc, err := a.accountingSvc.CreateChartOfAccounts(ctx, accounting.CreateChartOfAccountsInput{
		CompanyID:      demoCompanyID,
		Code:           code,
		Name:           req.Name,
		Type:           enums.AccountType(req.Type),
		ParentID:       pid,
		Path:           path,
		Depth:          depth,
		AllowsMovement: req.AllowsMovement,
		Description:    req.Description,
	})
	if err != nil {
		return nil, err
	}
	return toAccountDTO(acc), nil
}

// UpdateChartOfAccountRequest updates a chart-of-accounts entry.
type UpdateChartOfAccountRequest struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	ParentID       string `json:"parentId"`
	AllowsMovement bool   `json:"allowsMovement"`
	IsActive       bool   `json:"isActive"`
	Description    string `json:"description"`
}

// UpdateChartOfAccount persists changes to a chart-of-accounts entry.
func (a *App) UpdateChartOfAccount(req UpdateChartOfAccountRequest) (*AccountDTO, error) {
	ctx := a.Context()
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	parent, err := a.resolveChartParent(ctx, req.ParentID)
	if err != nil {
		return nil, err
	}
	var pid *uuid.UUID
	if parent != nil {
		pid = &parent.ID
	}
	updateCode, err := valueobjects.NewChartOfAccountsCode(req.Code)
	if err != nil {
		return nil, err
	}
	depth := updateCode.Depth()
	path := req.Code
	if parent != nil {
		depth = parent.Depth + 1
		path = accountPath(parent, req.Code)
	}
	acc, err := a.accountingSvc.UpdateChartOfAccount(ctx, accounting.UpdateChartOfAccountInput{
		ID:             id,
		Name:           req.Name,
		Type:           enums.AccountType(req.Type),
		ParentID:       pid,
		Code:           req.Code,
		Path:           path,
		Depth:          depth,
		AllowsMovement: req.AllowsMovement,
		IsActive:       req.IsActive,
		Description:    req.Description,
	})
	if err != nil {
		return nil, err
	}
	return toAccountDTO(acc), nil
}

// DeleteChartOfAccount deactivates a chart-of-accounts entry.
func (a *App) DeleteChartOfAccount(id string) error {
	aid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return a.accountingSvc.DeleteChartOfAccount(a.Context(), aid)
}

// ListFiscalPeriods returns the fiscal periods of the demo company.
func (a *App) ListFiscalPeriods() ([]*FiscalPeriodDTO, error) {
	ctx := a.Context()
	periods, err := a.accountingSvc.ListFiscalPeriods(ctx, demoCompanyID)
	if err != nil {
		return nil, err
	}
	items := make([]*FiscalPeriodDTO, 0, len(periods))
	for _, p := range periods {
		items = append(items, toFiscalPeriodDTO(p))
	}
	return items, nil
}

// resolveChartParent returns the parent account for a given parent ID
// (nil when the ID is empty). Lookups are scoped to the demo company.
func (a *App) resolveChartParent(ctx context.Context, parentID string) (*accounting.ChartOfAccount, error) {
	if parentID == "" {
		return nil, nil
	}
	pid, err := uuid.Parse(parentID)
	if err != nil {
		return nil, err
	}
	return a.accountingSvc.GetChartAccount(ctx, pid)
}

// accountPath builds the dotted path of an account from its parent.
func accountPath(parent *accounting.ChartOfAccount, code string) string {
	base := parent.Path
	if base == "" {
		base = parent.Code.String()
	}
	if base == "" {
		return code
	}
	return base + "." + code
}

// ListJournalEntriesRequest filters the journal listing.
type ListJournalEntriesRequest struct {
	Status string `json:"status"`
	PaginationRequest
}

// ListJournalEntries returns paged journal entries.
func (a *App) ListJournalEntries(req ListJournalEntriesRequest) (PageResult, error) {
	ctx := a.Context()
	filter := accounting.JournalEntryFilter{
		CompanyID:   &demoCompanyID,
		Status:      req.Status,
		PageRequest: req.toPageRequest(),
	}
	page, err := a.accountingSvc.ListEntries(ctx, filter)
	if err != nil {
		return PageResult{}, err
	}
	items := make([]*JournalEntryDTO, 0, len(page.Items))
	for _, e := range page.Items {
		items = append(items, toJournalEntryDTO(e))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// CreateJournalEntryLineRequest is one line of a new entry.
type CreateJournalEntryLineRequest struct {
	AccountID   string `json:"accountId"`
	Description string `json:"description"`
	Debit       string `json:"debit"`
	Credit      string `json:"credit"`
}

// CreateJournalEntryRequest creates a draft journal entry.
type CreateJournalEntryRequest struct {
	FiscalPeriodID string                          `json:"fiscalPeriodId"`
	EntryDate      string                          `json:"entryDate"`
	Description    string                          `json:"description"`
	Lines          []CreateJournalEntryLineRequest `json:"lines"`
}

// CreateJournalEntry persists a draft entry. When no fiscal period is
// supplied, the open period covering the entry date is used (and
// created on first use); the journal number is auto-generated.
func (a *App) CreateJournalEntry(req CreateJournalEntryRequest) (*JournalEntryDTO, error) {
	ctx := a.Context()
	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		return nil, err
	}
	var periodID uuid.UUID
	if req.FiscalPeriodID != "" {
		periodID, err = uuid.Parse(req.FiscalPeriodID)
		if err != nil {
			return nil, err
		}
	} else {
		period, err := a.accountingSvc.EnsureOpenFiscalPeriod(ctx, demoCompanyID, entryDate)
		if err != nil {
			return nil, err
		}
		periodID = period.ID
	}
	lines := make([]accounting.EntryLineInput, 0, len(req.Lines))
	for _, l := range req.Lines {
		accountID, err := uuid.Parse(l.AccountID)
		if err != nil {
			return nil, err
		}
		debit, err := valueobjects.MoneyFromString(l.Debit)
		if err != nil {
			return nil, err
		}
		credit, err := valueobjects.MoneyFromString(l.Credit)
		if err != nil {
			return nil, err
		}
		lines = append(lines, accounting.EntryLineInput{
			AccountID:   accountID,
			Description: l.Description,
			Debit:       debit,
			Credit:      credit,
		})
	}
	in := accounting.EntryInput{
		CompanyID:      demoCompanyID,
		FiscalPeriodID: periodID,
		Number:         "",
		EntryDate:      valueobjects.Date(entryDate),
		Description:    req.Description,
		Source:         enums.JournalTypeManual,
		Lines:          lines,
	}
	entry, err := a.accountingSvc.CreateEntry(ctx, in)
	if err != nil {
		return nil, err
	}
	return toJournalEntryDTO(entry), nil
}

// PostJournalEntry posts a draft entry, making it immutable.
func (a *App) PostJournalEntry(id string) (*JournalEntryDTO, error) {
	ctx := a.Context()
	eid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	entry, err := a.accountingSvc.Post(ctx, eid, demoCompanyID)
	if err != nil {
		return nil, err
	}
	return toJournalEntryDTO(entry), nil
}
