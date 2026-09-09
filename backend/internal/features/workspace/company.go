package workspace

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// LocalProfile is the singleton workspace owner record backed by the
// local_profiles table. It carries the optional password state used to
// lock and unlock the desktop app.
type LocalProfile struct {
	ID                uuid.UUID
	Name              string
	PasswordHash      string
	RecoveryTokenHash string
	PasswordEnabled   bool
	FailedAttempts    int
	LockedUntil       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewLocalProfile builds a profile with a fresh ID and UTC timestamps.
// The name is trimmed and must not be blank.
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

// Touch stamps UpdatedAt with the current UTC time.
func (p *LocalProfile) Touch() {
	p.UpdatedAt = time.Now().UTC()
}

// Validate reports whether the profile carries a usable name.
func (p *LocalProfile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidProfile
	}
	return nil
}
