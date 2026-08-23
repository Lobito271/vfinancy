package workspace

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID                   uuid.UUID
	Code                 string
	LegalName            string
	TradeName            string
	TaxID                string
	Address              string
	Phone                string
	Email                string
	CountryCode          string
	FunctionalCurrency   string
	Timezone             string
	FiscalYearStartMonth int
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}

type LocalProfile struct {
	ID              uuid.UUID
	Name            string
	PasswordHash    string
	PasswordEnabled bool
	FailedAttempts  int
	LockedUntil     *time.Time
	ActiveCompanyID uuid.UUID
	Theme           string
	Language        string
	DateFormat      string
	NumberFormat    string
	DecimalPlaces   int
	Timezone        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (p *LocalProfile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidProfile
	}
	if p.ActiveCompanyID == uuid.Nil {
		return ErrCompanyRequired
	}
	return nil
}

func (c *Company) Validate() error {
	if strings.TrimSpace(c.Code) == "" || strings.TrimSpace(c.LegalName) == "" || strings.TrimSpace(c.TaxID) == "" {
		return ErrInvalidCompany
	}
	if c.FiscalYearStartMonth < 1 || c.FiscalYearStartMonth > 12 {
		return ErrInvalidCompany
	}
	return nil
}
