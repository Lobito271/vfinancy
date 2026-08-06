package identity

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/valueobjects"
)

func mustEmail(t *testing.T) valueobjects.Email {
	t.Helper()
	e, _ := valueobjects.NewEmail("u@e.com")
	return e
}

func mustFullName(t *testing.T, s ...string) valueobjects.FullName {
	t.Helper()
	if len(s) == 0 {
		s = []string{"Test User"}
	}
	n, _ := valueobjects.NewFullName(s[0])
	return n
}

func mustShort(t *testing.T, s string) valueobjects.ShortCode {
	t.Helper()
	c, _ := valueobjects.NewShortCode(s)
	return c
}

func TestCompanyNew(t *testing.T) {
	now := time.Now()
	c, err := NewCompany(now, NewCompanyOptions{
		Code:                 mustShort(t, "vfi"),
		LegalName:            mustFullName(t, "vfinancy"),
		TradeName:            valueobjects.FullName{}, // optional
		TaxID:                "20600000001",
		CountryCode:          "PE",
		FunctionalCurrency:   valueobjects.MustCurrencyCode("PEN"),
		Timezone:             "America/Lima",
		FiscalYearStartMonth: 1,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !c.IsActive {
		t.Error("should be active by default")
	}
	if c.FiscalYearStartMonth != 1 {
		t.Errorf("fiscal month: %d", c.FiscalYearStartMonth)
	}
}

func TestCompanyInvalidFiscalMonth(t *testing.T) {
	now := time.Now()
	if _, err := NewCompany(now, NewCompanyOptions{
		Code:                 mustShort(t, "vfi"),
		LegalName:            mustFullName(t, "vfinancy"),
		TaxID:                "1",
		CountryCode:          "PE",
		FunctionalCurrency:   valueobjects.MustCurrencyCode("PEN"),
		FiscalYearStartMonth: 13,
	}); err == nil {
		t.Error("month 13 should fail")
	}
	if _, err := NewCompany(now, NewCompanyOptions{
		Code:                 mustShort(t, "vfi"),
		LegalName:            mustFullName(t, "vfinancy"),
		TaxID:                "1",
		CountryCode:          "PE",
		FunctionalCurrency:   valueobjects.MustCurrencyCode("PEN"),
		FiscalYearStartMonth: 0,
	}); err == nil {
		t.Error("month 0 should fail")
	}
}

func TestCompanyDeactivate(t *testing.T) {
	c, _ := NewCompany(time.Now(), NewCompanyOptions{
		Code:                 mustShort(t, "vfi"),
		LegalName:            mustFullName(t, "vfinancy"),
		TaxID:                "1",
		CountryCode:          "PE",
		FunctionalCurrency:   valueobjects.MustCurrencyCode("PEN"),
		FiscalYearStartMonth: 1,
	})
	c.Deactivate()
	if c.IsActive {
		t.Error("should be inactive")
	}
	c.Activate()
	if !c.IsActive {
		t.Error("should be active")
	}
}

func TestBranchNew(t *testing.T) {
	b, err := NewBranch(time.Now(), NewBranchOptions{
		CompanyID: uuid.New(),
		Code:      mustShort(t, "sede-01"),
		Name:      mustFullName(t, "Sede Central"),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !b.IsActive {
		t.Error("should be active")
	}
}

func TestBranchRequiresCompany(t *testing.T) {
	if _, err := NewBranch(time.Now(), NewBranchOptions{
		CompanyID: uuid.Nil,
		Code:      mustShort(t, "x"),
		Name:      mustFullName(t, "X"),
	}); err == nil {
		t.Error("empty company should fail")
	}
}

func TestUserLockout(t *testing.T) {
	maxAttempts := 3
	lockout := time.Minute
	now := time.Now()

	u, err := NewUser(now, NewUserOptions{
		CompanyID:    uuid.New(),
		Username:     mustShort(t, "admin"),
		Email:        mustEmail(t),
		FullName:     mustFullName(t, "Admin"),
		PasswordHash: "$argon2id$...",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !u.MustChangePassword {
		t.Error("new user must require password change")
	}

	// Two failed attempts, not locked
	u.RecordFailedLogin(maxAttempts, lockout, now)
	u.RecordFailedLogin(maxAttempts, lockout, now)
	if u.IsLocked(now) {
		t.Error("should not be locked after 2 attempts")
	}

	// Third attempt locks
	u.RecordFailedLogin(maxAttempts, lockout, now)
	if !u.IsLocked(now) {
		t.Error("should be locked after 3 attempts")
	}

	// Successful login resets
	ip := "127.0.0.1"
	later := now.Add(2 * time.Minute)
	u.RecordSuccessfulLogin(later, ip)
	if u.IsLocked(later) {
		t.Error("should not be locked after successful login")
	}
	if u.FailedLoginAttempts != 0 {
		t.Errorf("attempts: %d", u.FailedLoginAttempts)
	}
	if u.LastLoginIP != ip {
		t.Errorf("ip: %s", u.LastLoginIP)
	}
	if u.MustChangePassword {
		t.Error("login should clear must-change flag")
	}
}

func TestUserStatusTransitions(t *testing.T) {
	u, _ := NewUser(time.Now(), NewUserOptions{
		CompanyID:    uuid.New(),
		Username:     mustShort(t, "u"),
		Email:        mustEmail(t),
		FullName:     mustFullName(t, "User"),
		PasswordHash: "h",
	})
	now := time.Now()
	if got := u.Status(now); got != enums.UserStatusActive {
		t.Errorf("status: %s", got)
	}
	u.LockedUntil = ptrTime(now.Add(time.Hour))
	if got := u.Status(now); got != enums.UserStatusLocked {
		t.Errorf("status locked: %s", got)
	}
	u.Deactivate()
	if got := u.Status(now); got != enums.UserStatusInactive {
		t.Errorf("status inactive: %s", got)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestRoleGrantSystemRoleRejected(t *testing.T) {
	r, _ := NewRole(time.Now(), NewRoleOptions{
		CompanyID: uuid.New(),
		Code:      mustShort(t, "admin"),
		Name:      mustFullName(t, "Admin"),
		Type:      enums.RoleTypeSystem,
	})
	if err := r.Grant(Permission{Code: "customers.view"}); err == nil {
		t.Error("system role grant should fail")
	}
	if err := r.Rename(mustFullName(t, "X")); err == nil {
		t.Error("system role rename should fail")
	}
}

func TestRoleGrantCustomRole(t *testing.T) {
	r, _ := NewRole(time.Now(), NewRoleOptions{
		CompanyID: uuid.New(),
		Code:      mustShort(t, "manager"),
		Name:      mustFullName(t, "Manager"),
		Type:      enums.RoleTypeCustom,
	})
	if err := r.Grant(Permission{Code: "sales.view"}); err != nil {
		t.Errorf("grant: %v", err)
	}
	if !r.HasPermission("sales.view") {
		t.Error("should have permission")
	}
	r.Revoke("sales.view")
	if r.HasPermission("sales.view") {
		t.Error("should not have permission after revoke")
	}
}
