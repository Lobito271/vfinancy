package workspace

import (
	"context"
	"errors"
	"testing"

	"vfinancy/backend/internal/domain/repositories"
)

type noopTxManager struct{}

func (noopTxManager) WithinTransaction(ctx context.Context, fn repositories.TxRunner) error {
	return fn(ctx)
}

func newTestService(repo Repository) *Service {
	return NewService(repo, noopTxManager{})
}

type memoryRepository struct {
	profile *LocalProfile
}

func (r *memoryRepository) GetProfile(context.Context) (*LocalProfile, error) {
	if r.profile == nil {
		return nil, ErrProfileNotFound
	}
	copied := *r.profile
	return &copied, nil
}

func (r *memoryRepository) CreateProfile(_ context.Context, p *LocalProfile) error {
	copied := *p
	r.profile = &copied
	return nil
}

func (r *memoryRepository) UpdateProfile(_ context.Context, p *LocalProfile) error {
	copied := *p
	r.profile = &copied
	return nil
}

func setupService(t *testing.T, password string) (*Service, *memoryRepository) {
	t.Helper()
	repo := &memoryRepository{}
	service := newTestService(repo)
	if _, err := service.Setup(context.Background(), "Owner", password); err != nil {
		t.Fatal(err)
	}
	return service, repo
}

func TestSetupCreatesTrimmedUnlockedProfileWithoutPassword(t *testing.T) {
	repo := &memoryRepository{}
	service := newTestService(repo)

	profile, err := service.Setup(context.Background(), "  Owner  ", "")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Name != "Owner" {
		t.Fatalf("name = %q, want trimmed Owner", profile.Name)
	}
	if profile.PasswordEnabled || !service.IsUnlocked() {
		t.Fatal("password-less profile should start unlocked")
	}
	if !service.IsConfigured() {
		t.Fatal("setup should configure the service")
	}
	if got, err := service.Profile(); err != nil || got.ID != profile.ID {
		t.Fatalf("Profile() = %v, %v", got, err)
	}
}

func TestSetupEnablesPassword(t *testing.T) {
	_, repo := setupService(t, "Correct-horse-1")

	if !repo.profile.PasswordEnabled || repo.profile.PasswordHash == "" {
		t.Fatal("expected profile to be password-enabled")
	}
	if ok, err := VerifyPassword("Correct-horse-1", repo.profile.PasswordHash); err != nil || !ok {
		t.Fatalf("stored hash should verify against the setup password: %v", err)
	}
}

func TestSetupRejectsWeakPasswordWithoutCreatingProfile(t *testing.T) {
	repo := &memoryRepository{}
	service := newTestService(repo)

	if _, err := service.Setup(context.Background(), "Owner", "short"); err == nil {
		t.Fatal("expected weak password to be rejected")
	}
	if repo.profile != nil {
		t.Fatal("weak password should not have created the profile")
	}
}

func TestSetupRejectsExistingProfile(t *testing.T) {
	repo := &memoryRepository{}
	service := newTestService(repo)

	if _, err := service.Setup(context.Background(), "Owner", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Setup(context.Background(), "Second", ""); !errors.Is(err, ErrProfileExists) {
		t.Fatalf("err = %v, want ErrProfileExists", err)
	}
}

func TestInitializeMarksPasswordProfilesLocked(t *testing.T) {
	_, repo := setupService(t, "Correct-horse-1")

	fresh := newTestService(repo)
	if _, err := fresh.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fresh.IsUnlocked() {
		t.Fatal("password-protected profile should start locked")
	}
	if !fresh.PasswordEnabled() {
		t.Fatal("profile should report password enabled")
	}
	if err := fresh.RequireUnlocked(); !errors.Is(err, ErrProfileLocked) {
		t.Fatalf("RequireUnlocked() = %v, want ErrProfileLocked", err)
	}
}

func TestLockoutAfterFiveWrongAttempts(t *testing.T) {
	service, repo := setupService(t, "Correct-horse-1")
	service.Lock()

	for i := 0; i < 5; i++ {
		if err := service.Unlock(context.Background(), "wrong"); !errors.Is(err, ErrPasswordWrong) {
			t.Fatalf("attempt %d: err = %v, want ErrPasswordWrong", i+1, err)
		}
	}
	if err := service.Unlock(context.Background(), "Correct-horse-1"); !errors.Is(err, ErrProfileLocked) {
		t.Fatalf("err = %v, want ErrProfileLocked after lockout", err)
	}
	if service.IsUnlocked() {
		t.Fatal("locked profile should not unlock")
	}
	if repo.profile.LockedUntil == nil {
		t.Fatal("lockout should persist a LockedUntil timestamp")
	}
}

func TestUnlockResetsFailedAttempts(t *testing.T) {
	service, repo := setupService(t, "Correct-horse-1")
	service.Lock()

	for i := 0; i < 2; i++ {
		if err := service.Unlock(context.Background(), "wrong"); !errors.Is(err, ErrPasswordWrong) {
			t.Fatalf("attempt %d: err = %v", i+1, err)
		}
	}
	if err := service.Unlock(context.Background(), "Correct-horse-1"); err != nil {
		t.Fatal(err)
	}
	if !service.IsUnlocked() {
		t.Fatal("correct password should unlock")
	}
	if repo.profile.FailedAttempts != 0 || repo.profile.LockedUntil != nil {
		t.Fatal("unlock should reset the lockout state")
	}
}

func TestSetPasswordVerifiesCurrent(t *testing.T) {
	service, repo := setupService(t, "Correct-horse-1")

	if err := service.SetPassword(context.Background(), "wrong", "Another-Pass-1"); !errors.Is(err, ErrPasswordWrong) {
		t.Fatalf("err = %v, want ErrPasswordWrong", err)
	}
	if err := service.SetPassword(context.Background(), "Correct-horse-1", "short"); err == nil {
		t.Fatal("expected weak new password to be rejected")
	}
	if err := service.SetPassword(context.Background(), "Correct-horse-1", "Another-Pass-1"); err != nil {
		t.Fatal(err)
	}
	if ok, err := VerifyPassword("Another-Pass-1", repo.profile.PasswordHash); err != nil || !ok {
		t.Fatalf("password should be replaced: %v", err)
	}
}

func TestSetPasswordWithoutCurrentPassword(t *testing.T) {
	service, repo := setupService(t, "")

	if err := service.SetPassword(context.Background(), "", "Another-Pass-1"); err != nil {
		t.Fatal(err)
	}
	if !repo.profile.PasswordEnabled {
		t.Fatal("expected password to be enabled")
	}
	if ok, err := VerifyPassword("Another-Pass-1", repo.profile.PasswordHash); err != nil || !ok {
		t.Fatalf("stored hash should verify: %v", err)
	}
}

func TestRecoveryTokenIssuedOnce(t *testing.T) {
	service, _ := setupService(t, "Correct-horse-1")

	if _, err := service.GenerateRecoveryToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GenerateRecoveryToken(context.Background()); !errors.Is(err, ErrRecoveryIssued) {
		t.Fatalf("err = %v, want ErrRecoveryIssued", err)
	}
}

func TestRecoveryTokenResetsPassword(t *testing.T) {
	service, repo := setupService(t, "Correct-horse-1")

	token, err := service.GenerateRecoveryToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := service.UnlockWithRecoveryToken(context.Background(), "bad-token", "Another-Pass-1"); !errors.Is(err, ErrRecoveryTokenInvalid) {
		t.Fatalf("err = %v, want ErrRecoveryTokenInvalid", err)
	}
	if err := service.UnlockWithRecoveryToken(context.Background(), token, "weak"); err == nil {
		t.Fatal("expected weak new password to be rejected")
	}
	if err := service.UnlockWithRecoveryToken(context.Background(), token, "Another-Pass-1"); err != nil {
		t.Fatal(err)
	}
	if !service.IsUnlocked() {
		t.Fatal("recovery should unlock the workspace")
	}
	if ok, err := VerifyPassword("Another-Pass-1", repo.profile.PasswordHash); err != nil || !ok {
		t.Fatalf("password should be reset: %v", err)
	}
	if repo.profile.RecoveryTokenHash != "" {
		t.Fatal("recovery token should be cleared after use")
	}
	if repo.profile.FailedAttempts != 0 || repo.profile.LockedUntil != nil {
		t.Fatal("recovery should reset the lockout state")
	}
	if err := service.UnlockWithRecoveryToken(context.Background(), token, "Third-Pass-1"); !errors.Is(err, ErrRecoveryTokenInvalid) {
		t.Fatal("recovery token should not be reusable")
	}
}

func TestUnlockWithRecoveryTokenWithoutPasswordIsNoop(t *testing.T) {
	service, repo := setupService(t, "")

	if err := service.UnlockWithRecoveryToken(context.Background(), "any", "Another-Pass-1"); err != nil {
		t.Fatal(err)
	}
	if repo.profile.PasswordEnabled {
		t.Fatal("password should stay disabled")
	}
}
