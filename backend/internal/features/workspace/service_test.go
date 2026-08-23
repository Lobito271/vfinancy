package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memoryRepository struct {
	profile   *LocalProfile
	companies map[uuid.UUID]*Company
}

func (r *memoryRepository) GetProfile(context.Context) (*LocalProfile, error) {
	if r.profile == nil {
		return nil, ErrProfileNotFound
	}
	copy := *r.profile
	return &copy, nil
}

func (r *memoryRepository) CreateProfile(_ context.Context, p *LocalProfile) error {
	r.profile = p
	return nil
}

func (r *memoryRepository) UpdateProfile(_ context.Context, p *LocalProfile) error {
	r.profile = p
	return nil
}

func (r *memoryRepository) ListCompanies(context.Context) ([]*Company, error) {
	result := make([]*Company, 0, len(r.companies))
	for _, c := range r.companies {
		result = append(result, c)
	}
	return result, nil
}

func (r *memoryRepository) GetCompany(_ context.Context, id uuid.UUID) (*Company, error) {
	c, ok := r.companies[id]
	if !ok {
		return nil, ErrInvalidCompany
	}
	return c, nil
}

func (r *memoryRepository) CreateCompany(_ context.Context, c *Company) error {
	r.companies[c.ID] = c
	return nil
}

func (r *memoryRepository) UpdateCompany(_ context.Context, c *Company) error {
	r.companies[c.ID] = c
	return nil
}

func TestLocalProfileUsesActiveCompanyAndOptionalPassword(t *testing.T) {
	companyID := uuid.New()
	repo := &memoryRepository{companies: map[uuid.UUID]*Company{companyID: {
		ID: companyID, Code: "ACME", LegalName: "Acme", TaxID: "123", IsActive: true,
		FiscalYearStartMonth: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}}}
	service := NewService(repo)

	profile, err := service.CreateProfile(context.Background(), "Owner", companyID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.PasswordEnabled || !service.IsUnlocked() {
		t.Fatal("profile without password should start unlocked")
	}
	if got, err := service.CurrentCompanyID(); err != nil || got != companyID {
		t.Fatalf("active company = %v, %v", got, err)
	}
}
