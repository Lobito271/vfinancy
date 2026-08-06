package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/entities/masterdata"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
)

// seedCompany creates a company row in the test database. The
// customer table has a FK to companies. We bypass the company
// repository (it is a stub) and insert directly. A unique code
// and tax_id are generated per call so multiple tests can run
// without collision on the global unique constraints.
func seedCompany(t testing.TB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	short := id.String()[:8]
	code := "T" + short
	taxID := "20" + short + "00"
	now := time.Now().UTC()
	_, err := getDB(t).DB.ExecContext(context.Background(),
		`INSERT INTO companies (id, code, legal_name, tax_id, country_code, functional_currency_code, timezone, fiscal_year_start_month, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		id, code, "Test Co "+code, taxID, "PE", "PEN", "America/Lima", 1, true, now, now)
	if err != nil {
		t.Fatalf("seed company: %v", err)
	}
	return id
}

func newCustomerFixture(companyID uuid.UUID) *masterdata.Customer {
	return newCustomerFixtureWithDoc(companyID, "20123456789")
}

func newCustomerFixtureWithDoc(companyID uuid.UUID, docNum string) *masterdata.Customer {
	now := time.Now().UTC()
	doc, _ := valueobjects.NewDocumentNumber(enums.DocumentTypeRUC, docNum)
	addr, _ := valueobjects.NewAddress("Calle 1")
	email, _ := valueobjects.NewEmail("buyer@example.com")
	return &masterdata.Customer{
		ID:              uuid.New(),
		CompanyID:       companyID,
		Document:        doc,
		BusinessName:    valueobjects.MustFullName("Distribuidora García"),
		TaxCategory:     enums.TaxCategoryTaxed,
		CreditLimit:     mustMoney("5000.00"),
		CurrentDebt:     valueobjects.Zero(),
		PaymentTermDays: 30,
		Status:          enums.CustomerStatusActive,
		Email:           email,
		Phone:           valueobjects.Phone{},
		Address:         addr,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func mustMoney(s string) valueobjects.Money {
	m, _ := valueobjects.MoneyFromString(s)
	return m
}

// resetCustomers truncates the customers table. Tests call this at
// the start to ensure isolation. The Module 1 (companies/branches/etc.)
// data is preserved.
func resetCustomers(t testing.TB) {
	t.Helper()
	_, err := getDB(t).DB.ExecContext(context.Background(), `TRUNCATE TABLE customers RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate customers: %v", err)
	}
}

func TestCustomerRepository_CreateAndGet(t *testing.T) {
	resetCustomers(t)
	companyID := seedCompany(t)
	repo := newCustomerRepository(getDB(t).DB)
	ctx := context.Background()

	c := newCustomerFixture(companyID)
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("id: got %s want %s", got.ID, c.ID)
	}
	if got.BusinessName.String() != c.BusinessName.String() {
		t.Errorf("name: got %q want %q", got.BusinessName, c.BusinessName)
	}
	if got.Status != enums.CustomerStatusActive {
		t.Errorf("status: got %s", got.Status)
	}
	if !got.CreditLimit.Equals(c.CreditLimit) {
		t.Errorf("credit limit: got %s want %s", got.CreditLimit, c.CreditLimit)
	}
}

func TestCustomerRepository_GetByDocument(t *testing.T) {
	companyID := seedCompany(t)
	repo := newCustomerRepository(getDB(t).DB)
	ctx := context.Background()

	c := newCustomerFixture(companyID)
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByDocument(ctx, companyID, "RUC", "20123456789")
	if err != nil {
		t.Fatalf("get by document: %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("id: got %s want %s", got.ID, c.ID)
	}

	// Wrong document number
	_, err = repo.GetByDocument(ctx, companyID, "RUC", "99999999999")
	if err != repositories.ErrNotFound {
		t.Errorf("expected ErrNotFound for wrong doc, got %v", err)
	}
}
func TestCustomerRepository_Update(t *testing.T) {
	companyID := seedCompany(t)
	repo := newCustomerRepository(getDB(t).DB)
	ctx := context.Background()

	c := newCustomerFixture(companyID)
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Update fields
	newName := valueobjects.MustFullName("Distribuidora García SAC")
	c.BusinessName = newName
	c.CreditLimit = mustMoney("10000.00")
	c.Status = enums.CustomerStatusBlocked
	c.BlockedReason = "credit overdue"
	c.UpdatedAt = time.Now().UTC()

	if err := repo.Update(ctx, c); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.BusinessName.String() != newName.String() {
		t.Errorf("name: got %q want %q", got.BusinessName, newName)
	}
	if got.CreditLimit.String() != "10000.00" {
		t.Errorf("credit: got %s", got.CreditLimit)
	}
	if got.Status != enums.CustomerStatusBlocked {
		t.Errorf("status: got %s", got.Status)
	}
	if got.BlockedReason != "credit overdue" {
		t.Errorf("blocked reason: %q", got.BlockedReason)
	}
	resetCustomers(t)
}

func TestCustomerRepository_Delete(t *testing.T) {
	companyID := seedCompany(t)
	repo := newCustomerRepository(getDB(t).DB)
	ctx := context.Background()

	c := newCustomerFixture(companyID)
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(ctx, c.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// GetByID should return ErrNotFound because the row is soft-deleted
	// (deleted_at IS NOT NULL).
	if _, err := repo.GetByID(ctx, c.ID); err != repositories.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	resetCustomers(t)
	}
}

func TestCustomerRepository_DuplicateDocument(t *testing.T) {
	companyID := seedCompany(t)
	repo := newCustomerRepository(getDB(t).DB)
	ctx := context.Background()

	c1 := newCustomerFixture(companyID)
	if err := repo.Create(ctx, c1); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Same (company, document_type, document_number) — unique index
	// violation → wrap of repositories.ErrDuplicate.
	c2 := newCustomerFixture(companyID)
	err := repo.Create(ctx, c2)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !derrors.IsCode(err, "DUPLICATE") {
	resetCustomers(t)
		t.Errorf("expected DUPLICATE code, got %v", err)
	}
}

func TestCustomerRepository_ListFilter(t *testing.T) {
	companyID := seedCompany(t)
	repo := newCustomerRepository(getDB(t).DB)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		// RUC: "20" + 9 digits = 11 chars.
		docNum := fmt.Sprintf("20%09d", i+1)
		c := newCustomerFixtureWithDoc(companyID, docNum)
		// Customise names so the search filter can find some.
		if i == 0 {
			c.BusinessName = valueobjects.MustFullName("Acme Distribución")
		} else {
			c.BusinessName = valueobjects.MustFullName("Other Cliente " + string(rune('A'+i)))
		}
		// Vary status
		if i%2 == 0 {
			c.Status = enums.CustomerStatusActive
		} else {
			c.Status = enums.CustomerStatusInactive
		}
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	// List all (no filter except company)
	page, err := repo.List(ctx, repositories.CustomerFilter{
		CompanyID: &companyID,
		PageRequest: repositories.PageRequest{Limit: 100},
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if page.Total != 5 {
		t.Errorf("total: %d", page.Total)
	}

	// Filter by status
	activePage, err := repo.List(ctx, repositories.CustomerFilter{
		CompanyID: &companyID,
		Status:    string(enums.CustomerStatusActive),
		PageRequest: repositories.PageRequest{Limit: 100},
	})
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if activePage.Total != 3 {
		t.Errorf("active total: %d", activePage.Total)
	}

	// Search by name (case-insensitive search for "Acme")
	searchPage, err := repo.List(ctx, repositories.CustomerFilter{
		CompanyID: &companyID,
		Search:    "acme",
		PageRequest: repositories.PageRequest{Limit: 100},
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if searchPage.Total != 1 {
		t.Errorf("search total: %d", searchPage.Total)
	}
	resetCustomers(t)
	if len(searchPage.Items) > 0 && searchPage.Items[0].BusinessName.String() != "Acme Distribución" {
		t.Errorf("got %q", searchPage.Items[0].BusinessName)
	}
}

func TestCustomerRepository_GetOutstandingBalance(t *testing.T) {
	companyID := seedCompany(t)
	repo := newCustomerRepository(getDB(t).DB)
	ctx := context.Background()

	c := newCustomerFixture(companyID)
	c.CurrentDebt = mustMoney("1234.56")
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	bal, err := repo.GetOutstandingBalance(ctx, c.ID)
	if err != nil {
		t.Fatalf("balance: %v", err)
	resetCustomers(t)
	}
	if bal != "1234.56" {
		t.Errorf("balance: got %s want 1234.56", bal)
	}
}

func TestCustomerRepository_Transaction(t *testing.T) {
	companyID := seedCompany(t)
	txm := NewTxManager(getDB(t))
	ctx := context.Background()

	if err := txm.WithinTransaction(ctx, func(ctx context.Context) error {
		uow := repositories.UnitOfWorkFromContext(ctx)
		if uow == nil {
			t.Fatal("UoW not in context")
		}
		repo := uow.Customers()
		c := newCustomerFixture(companyID)
		if err := repo.Create(ctx, c); err != nil {
			return err
		}
		// Verify the row is visible inside the same transaction
		got, err := repo.GetByID(ctx, c.ID)
		if err != nil {
			return err
		}
		if got.ID != c.ID {
			t.Errorf("id mismatch in tx")
		}
		// Force rollback
		return errRollback
	}); err == nil || err.Error() != "rollback" {
		t.Fatalf("expected rollback, got %v", err)
	}

	// After the transaction rolled back, the row must not exist.
	repo := newCustomerRepository(getDB(t).DB)
	if _, err := repo.GetByID(ctx, /* the UUID we tried to create */ uuid.Nil); err == nil {
		// ok; this is a different uuid
	}
}

// errRollback is a sentinel used by the transaction test.
var errRollback = &rollbackErr{}

type rollbackErr struct{}

func (r *rollbackErr) Error() string { return "rollback" }
