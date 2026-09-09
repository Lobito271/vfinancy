package workspace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"vfinancy/backend/internal/domain/repositories"
)

type Service struct {
	repo     Repository
	txm      repositories.TransactionManager
	mu       sync.RWMutex
	profile  *LocalProfile
	unlocked bool
}

// NewService builds the workspace service over the local-profile
// repository and the shared transaction manager.
func NewService(repo Repository, txm repositories.TransactionManager) *Service {
	return &Service{repo: repo, txm: txm}
}

// Initialize loads the stored local profile into memory and marks the
// workspace unlocked when the profile has no password. It returns
// ErrProfileNotFound when no profile has been set up yet.
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
	s.profile = profile
	s.unlocked = !profile.PasswordEnabled
	return cloneProfile(profile), nil
}

// Setup creates the one-and-only local profile and returns it. It
// fails with ErrProfileExists when a profile already exists. A
// non-empty password is strength-checked and hashed, enabling the
// lock screen.
func (s *Service) Setup(ctx context.Context, name, password string) (*LocalProfile, error) {
	profile, err := NewLocalProfile(name)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if err := ValidatePasswordStrength(password); err != nil {
			return nil, err
		}
		hash, err := HashPassword(password, nil)
		if err != nil {
			return nil, err
		}
		profile.PasswordHash = hash
		profile.PasswordEnabled = true
	}
	if err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		_, err := s.repo.GetProfile(ctx)
		switch {
		case err == nil:
			return ErrProfileExists
		case !errors.Is(err, ErrProfileNotFound):
			return err
		}
		return s.repo.CreateProfile(ctx, profile)
	}); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.profile = profile
	s.unlocked = true
	s.mu.Unlock()
	return cloneProfile(profile), nil
}

// Profile returns a copy of the in-memory profile, or
// ErrProfileNotFound when none is loaded.
func (s *Service) Profile() (*LocalProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.profile == nil {
		return nil, ErrProfileNotFound
	}
	return cloneProfile(s.profile), nil
}

// IsConfigured reports whether a profile is loaded in memory.
func (s *Service) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile != nil
}

// PasswordEnabled reports whether the loaded profile requires a password.
func (s *Service) PasswordEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile != nil && s.profile.PasswordEnabled
}

// IsUnlocked reports whether the workspace is currently unlocked.
func (s *Service) IsUnlocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.unlocked
}

// RequireUnlocked returns ErrProfileLocked while the workspace is locked.
func (s *Service) RequireUnlocked() error {
	if !s.IsUnlocked() {
		return ErrProfileLocked
	}
	return nil
}

// Lock locks the workspace.
func (s *Service) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unlocked = false
}

// Unlock verifies the password and unlocks the workspace. Five
// consecutive wrong attempts lock the profile for fifteen minutes.
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

// SetPassword verifies the current password and replaces it. When the
// profile has no password yet the current one is not required.
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
	s.profile.Touch()
	if err := s.repo.UpdateProfile(ctx, s.profile); err != nil {
		return err
	}
	s.unlocked = true
	return nil
}

// RemovePassword verifies the current password and disables the lock
// screen, clearing any pending recovery token and lockout state.
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
	s.profile.RecoveryTokenHash = ""
	s.profile.FailedAttempts = 0
	s.profile.LockedUntil = nil
	s.profile.Touch()
	if err := s.repo.UpdateProfile(ctx, s.profile); err != nil {
		return err
	}
	s.unlocked = true
	return nil
}

// GenerateRecoveryToken creates a one-time recovery secret, stores only
// its hash, and returns the plaintext value. It can only be issued
// once per profile; subsequent calls return ErrRecoveryIssued. The
// token must be shown to the user at activation time and never
// recovered again from the stored hash.
func (s *Service) GenerateRecoveryToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profile == nil {
		return "", ErrProfileNotFound
	}
	if s.profile.RecoveryTokenHash != "" {
		return "", ErrRecoveryIssued
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	hash, err := HashPassword(token, nil)
	if err != nil {
		return "", err
	}
	s.profile.RecoveryTokenHash = hash
	s.profile.Touch()
	if err := s.repo.UpdateProfile(ctx, s.profile); err != nil {
		return "", err
	}
	return token, nil
}

// UnlockWithRecoveryToken resets a forgotten password using the
// one-time recovery token, clears the lockout state, and unlocks the
// workspace. It works even while the profile is locked.
func (s *Service) UnlockWithRecoveryToken(ctx context.Context, token, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.profile == nil {
		return ErrProfileNotFound
	}
	if !s.profile.PasswordEnabled {
		return nil
	}
	match, err := VerifyPassword(token, s.profile.RecoveryTokenHash)
	if err != nil || !match {
		return ErrRecoveryTokenInvalid
	}
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}
	hash, err := HashPassword(newPassword, nil)
	if err != nil {
		return err
	}
	s.profile.PasswordHash = hash
	s.profile.RecoveryTokenHash = ""
	s.profile.FailedAttempts = 0
	s.profile.LockedUntil = nil
	s.profile.Touch()
	if err := s.repo.UpdateProfile(ctx, s.profile); err != nil {
		return err
	}
	s.unlocked = true
	return nil
}

func cloneProfile(p *LocalProfile) *LocalProfile {
	copied := *p
	return &copied
}
