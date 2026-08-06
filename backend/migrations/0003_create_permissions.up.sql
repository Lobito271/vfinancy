-- 0003_create_permissions.up.sql
-- Module 1: Authentication
-- Global catalog of permission keys. No company_id: permissions are
-- universal across all tenants (a permission like "customers.delete" is
-- the same string everywhere). The mapping of permission → role is
-- per-company and lives in role_permissions.
--
-- code is the primary key, in the form "module.action" (e.g. "customers.delete").
-- VARCHAR(100) gives room for future hierarchical keys like
-- "sales.refund.approve".
--
-- No soft delete: permissions are a controlled catalog. Removing a
-- permission is a deliberate, audited operation; an obsolete permission
-- is best kept for historical audit_log references.
--
-- No audit columns: same rationale — this is a global lookup, not a
-- per-tenant business table.

CREATE TABLE permissions (
    code        VARCHAR(100) PRIMARY KEY,
    module      VARCHAR(50)  NOT NULL,
    action      VARCHAR(50)  NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_module
    ON permissions (module);
