package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"vfinancy/backend/infrastructure/config"
	"vfinancy/backend/infrastructure/database"
	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/infrastructure/migrations"
	"vfinancy/backend/infrastructure/postgres"
)

// TestMain starts a connection to the embedded test postgres and
// applies all migrations. The test is then free to use the *sql.DB
// exposed by getDB().
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Println("setup failed:", err)
		// We cannot skip via m.Skip; instead we exit non-zero so the
		// failure is visible in the test report.
		os.Exit(1)
	}
	defer teardown()
	m.Run()
}

var (
	testDB     *database.DB
	testConfig *config.Config
)

// getDB returns the test database handle.
func getDB(t testing.TB) *database.DB {
	t.Helper()
	if testDB == nil {
		t.Fatal("test database not initialized; TestMain should have set it up")
	}
	return testDB
}

func setup() error {
	// Locate the migrations directory from the current working dir.
	cwd, _ := os.Getwd()
	migrationsDir, err := filepath.Abs(filepath.Join(cwd, "..", "..", "..", "..", "migrations"))
	if err != nil {
		return fmt.Errorf("migrations dir: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.Database = config.DatabaseConfig{
		Host:         "127.0.0.1",
		Port:         5433,
		User:         "postgres",
		Name:         "vfinancy_test",
		SSLMode:      "disable",
		MaxOpen:      5,
		MaxIdle:      2,
		MaxLifetime:  5 * time.Minute,
		MigrationDir: migrationsDir,
	}

	testConfig = cfg
	log := logger.New("info", "text", "stdout")

	// Use the admin connection to (a) ensure the test database exists
	// and (b) clear any cached state from a previous run. We do NOT
	// drop+recreate: pgx connection pooling means the runner's
	// underlying *sql.DB may still hold a connection to the dropped
	// database. Instead we TRUNCATE every business table between
	// test invocations, which is the standard test isolation pattern.
	if err := ensureFreshDB(cfg); err != nil {
		return fmt.Errorf("ensure fresh db: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("DEBUG: connecting with DSN =", cfg.Database.DSN())

	db, err := postgres.Connect(ctx, &cfg.Database, log)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	testDB = db

	// Apply migrations if the schema is empty.
	runner := migrations.NewRunner(cfg.Database.MigrationDir, db.DB, log)
	if err := runner.Up(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	// Module 2 (Master Data) tables are not in the migrations. The
	// repository tests for those modules need them. The migration
	// runner has just run, so the companies table exists and the
	// customers table can be created with a valid FK.
	ensureConn, err := sql.Open(postgres.DriverName, fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Name))
	if err != nil {
		return fmt.Errorf("ensure open: %w", err)
	}
	defer ensureConn.Close()
	if err := ensureBusinessTables(ensureConn); err != nil {
		return fmt.Errorf("ensure business tables: %w", err)
	}
	return nil
}

func teardown() {
	if testDB != nil {
		_ = testDB.Close()
	}
}

func recreateDB(cfg *config.Config, log *logger.Logger) error { return ensureFreshDB(cfg) }

// ensureFreshDB makes sure the test database exists, all business
// tables are empty, and any previous-test connections are closed. The
// migration runner is responsible for creating the schema (it runs
// against an empty DB and the migrations are idempotent).
func ensureFreshDB(cfg *config.Config) error {
	adminDSN := fmt.Sprintf("host=%s port=%d user=%s dbname=postgres sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User)
	conn, err := sql.Open(postgres.DriverName, adminDSN)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create the database if it does not yet exist.
	var exists bool
	if err := conn.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.Database.Name,
	).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err := conn.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Database.Name)); err != nil {
			return err
		}
	}
	// From here on, talk to the test database directly (not postgres).
	testDSN := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Name)
	testConn, err := sql.Open(postgres.DriverName, testDSN)
	if err != nil {
		return err
	}
	defer testConn.Close()
	// TRUNCATE every business table that has rows. We use CASCADE so
	// that FK constraints do not block us. CASCADE on TRUNCATE is
	// safe in a test environment because all data is going away.
	// If the schema does not exist yet (first run), the truncate
	// raises "relation does not exist" — we treat that as a no-op.
	// Note: customers is not in the migrations; it's created in
	// ensureBusinessTables after the migrations have run.
	_, err = testConn.Exec(`
		TRUNCATE TABLE
			audit_logs,
			login_history,
			user_roles,
			role_permissions,
			users,
			roles,
			permissions,
			branches,
			companies
		RESTART IDENTITY CASCADE`)
	if err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "does not exist") {
			return err
		}
	}
	return nil
}

// ensureBusinessTables creates the minimum subset of the Module 2 /
// Module 3 schema needed by the repository tests. It is idempotent
// (uses IF NOT EXISTS) so the test runner can call it on every
// invocation. As Phase 1.2+ lands the corresponding migrations,
// these statements can be removed.
func ensureBusinessTables(conn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS customers (
			id UUID PRIMARY KEY,
			company_id UUID NOT NULL REFERENCES companies(id),
			default_branch_id UUID REFERENCES branches(id),
			document_type VARCHAR(10) NOT NULL,
			document_number VARCHAR(30) NOT NULL,
			business_name VARCHAR(200) NOT NULL,
			trade_name VARCHAR(200),
			tax_category VARCHAR(30) NOT NULL DEFAULT 'taxed',
			credit_limit NUMERIC(18,2) NOT NULL DEFAULT 0,
			current_debt NUMERIC(18,2) NOT NULL DEFAULT 0,
			payment_term_days INTEGER NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			blocked_reason TEXT,
			email VARCHAR(200),
			phone VARCHAR(30),
			address TEXT,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			created_by UUID,
			updated_by UUID
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_company_doc
			ON customers (company_id, document_type, LOWER(document_number))
			WHERE deleted_at IS NULL`,
		// Add columns that may have been missed in earlier runs of this
		// helper (the table may pre-date a code change).
		`ALTER TABLE customers ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE`,
	}
	for _, s := range stmts {
		if _, err := conn.Exec(s); err != nil {
			return fmt.Errorf("ensure table: %w", err)
		}
	}
	return nil
}

// sortableMigrations is unused but kept as a placeholder for the
// migration runner's internal sort.
var _ = sort.Strings
