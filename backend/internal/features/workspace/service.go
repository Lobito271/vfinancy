package workspace

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo     Repository
	mu       sync.RWMutex
	profile  *LocalProfile
	unlocked bool
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Initialize(ctx context.Context) (*LocalProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profile != nil {
		return cloneProfile(s.profile), nil
	}
	profile, err := s.repo.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	company, err := s.repo.GetCompany(ctx, profile.ActiveCompanyID)
	if err != nil {
		return nil, err
	}
	if !company.IsActive || company.DeletedAt != nil {
		return nil, ErrCompanyInactive
	}
	s.profile = profile
	s.unlocked = !profile.PasswordEnabled
	return cloneProfile(profile), nil
}

func (s *Service) CreateProfile(ctx context.Context, name string, companyID uuid.UUID) (*LocalProfile, error) {
	company, err := s.repo.GetCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if !company.IsActive || company.DeletedAt != nil {
		return nil, ErrCompanyInactive
	}
	profile := &LocalProfile{
		ID:              uuid.New(),
		Name:            strings.TrimSpace(name),
		ActiveCompanyID: companyID,
		Theme:           "system",
		Language:        "es-PE",
		DateFormat:      "DD/MM/YYYY",
		NumberFormat:    "es-PE",
		DecimalPlaces:   2,
		Timezone:        "America/Lima",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.profile = profile
	s.unlocked = true
	s.mu.Unlock()
	return cloneProfile(profile), nil
}

func (s *Service) Profile() (*LocalProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.profile == nil {
		return nil, ErrProfileNotFound
	}
	return cloneProfile(s.profile), nil
}

func (s *Service) ListCompanies(ctx context.Context) ([]*Company, error) {
	return s.repo.ListCompanies(ctx)
}

func (s *Service) GetCompany(ctx context.Context, id uuid.UUID) (*Company, error) {
	return s.repo.GetCompany(ctx, id)
}

func (s *Service) CreateCompany(ctx context.Context, company *Company) error {
	if company.ID == uuid.Nil {
		company.ID = uuid.New()
	}
	if company.CreatedAt.IsZero() {
		company.CreatedAt = time.Now().UTC()
	}
	if company.UpdatedAt.IsZero() {
		company.UpdatedAt = company.CreatedAt
	}
	if company.Timezone == "" {
		company.Timezone = "America/Lima"
	}
	if company.CountryCode == "" {
		company.CountryCode = "PE"
	}
	if company.FunctionalCurrency == "" {
		company.FunctionalCurrency = "PEN"
	}
	if company.FiscalYearStartMonth == 0 {
		company.FiscalYearStartMonth = 1
	}
	company.IsActive = true
	if err := company.Validate(); err != nil {
		return err
	}
	return s.repo.CreateCompany(ctx, company)
}

func (s *Service) UpdateCompany(ctx context.Context, company *Company) error {
	if err := company.Validate(); err != nil {
		return err
	}
	company.UpdatedAt = time.Now().UTC()
	return s.repo.UpdateCompany(ctx, company)
}

func (s *Service) CurrentCompanyID() (uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.profile == nil || s.profile.ActiveCompanyID == uuid.Nil {
		return uuid.Nil, ErrCompanyRequired
	}
	return s.profile.ActiveCompanyID, nil
}

func (s *Service) IsUnlocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.unlocked
}

func (s *Service) RequireUnlocked() error {
	if !s.IsUnlocked() {
		return ErrProfileLocked
	}
	return nil
}

func (s *Service) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unlocked = false
}

func (s *Service) Unlock(ctx context.Context, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profile == nil {
		return ErrProfileNotFound
	}
	if !s.profile.PasswordEnabled {
		s.unlocked = true
		return nil
	}
	now := time.Now().UTC()
	if s.profile.LockedUntil != nil && s.profile.LockedUntil.After(now) {
		return ErrProfileLocked
	}
	match, err := VerifyPassword(password, s.profile.PasswordHash)
	if err != nil || !match {
		s.profile.FailedAttempts++
		if s.profile.FailedAttempts >= 5 {
			locked := now.Add(15 * time.Minute)
			s.profile.LockedUntil = &locked
			s.profile.FailedAttempts = 0
		}
		_ = s.repo.UpdateProfile(ctx, s.profile)
		return ErrPasswordWrong
	}
	s.profile.FailedAttempts = 0
	s.profile.LockedUntil = nil
	s.unlocked = true
	return s.repo.UpdateProfile(ctx, s.profile)
}

func (s *Service) SetPassword(ctx context.Context, current, next string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profile == nil {
		return ErrProfileNotFound
	}
	if s.profile.PasswordEnabled {
		match, err := VerifyPassword(current, s.profile.PasswordHash)
		if err != nil || !match {
			return ErrPasswordWrong
		}
	}
	if err := ValidatePasswordStrength(next); err != nil {
		return err
	}
	hash, err := HashPassword(next, nil)
	if err != nil {
		return err
	}
	s.profile.PasswordHash = hash
	s.profile.PasswordEnabled = true
	s.profile.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateProfile(ctx, s.profile); err != nil {
		return err
	}
	s.unlocked = true
	return nil
}

func (s *Service) RemovePassword(ctx context.Context, current string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profile == nil {
		return ErrProfileNotFound
	}
	if s.profile.PasswordEnabled {
		match, err := VerifyPassword(current, s.profile.PasswordHash)
		if err != nil || !match {
			return ErrPasswordWrong
		}
	}
	s.profile.PasswordHash = ""
	s.profile.PasswordEnabled = false
	s.profile.FailedAttempts = 0
	s.profile.LockedUntil = nil
	s.profile.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateProfile(ctx, s.profile); err != nil {
		return err
	}
	s.unlocked = true
	return nil
}

func (s *Service) SetActiveCompany(ctx context.Context, id uuid.UUID) error {
	company, err := s.repo.GetCompany(ctx, id)
	if err != nil {
		return err
	}
	if !company.IsActive || company.DeletedAt != nil {
		return ErrCompanyInactive
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profile == nil {
		return ErrProfileNotFound
	}
	s.profile.ActiveCompanyID = id
	s.profile.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateProfile(ctx, s.profile); err != nil {
		return err
	}
	return nil
}

func cloneProfile(p *LocalProfile) *LocalProfile {
	copy := *p
	return &copy
}
