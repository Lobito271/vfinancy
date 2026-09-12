package workspace

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

// LocalProfile is the singleton workspace owner record backed by the
// local_profiles table. It carries the company identity data (SUNAT
// oriented) and the optional password state used to lock and unlock
// the desktop app.
type LocalProfile struct {
	ID                uuid.UUID
	Name              string
	TaxID             string
	Email             string
	FiscalAddress     string
	PasswordHash      string
	RecoveryTokenHash string
	PasswordEnabled   bool
	FailedAttempts    int
	LockedUntil       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewLocalProfile builds a profile with a fresh ID and UTC timestamps.
// The business name is trimmed and must not be blank.
func NewLocalProfile(name string) (*LocalProfile, error) {
	p := &LocalProfile{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(name),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// CompanyInput carries the corporate identity fields of the workspace
// owner (SUNAT oriented: razón social, RUC, correo y dirección fiscal).
type CompanyInput struct {
	Name          string
	TaxID         string
	Email         string
	FiscalAddress string
}

// SetCompany copies the corporate identity fields onto the profile.
func (p *LocalProfile) SetCompany(in CompanyInput) {
	p.Name = strings.TrimSpace(in.Name)
	p.TaxID = strings.TrimSpace(in.TaxID)
	p.Email = strings.TrimSpace(in.Email)
	p.FiscalAddress = strings.TrimSpace(in.FiscalAddress)
}

// Touch stamps UpdatedAt with the current UTC time.
func (p *LocalProfile) Touch() {
	p.UpdatedAt = time.Now().UTC()
}

// Validate reports whether the profile carries a usable business name
// and, when present, a valid RUC and email.
func (p *LocalProfile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidProfile
	}
	if p.TaxID != "" {
		if err := valueobjects.ValidateDocument(enums.DocumentTypeRUC, p.TaxID); err != nil {
			return err
		}
	}
	if p.Email != "" {
		email, err := valueobjects.NewEmail(p.Email)
		if err != nil {
			return err
		}
		p.Email = email.String()
	}
	p.FiscalAddress = strings.TrimSpace(p.FiscalAddress)
	return nil
}
