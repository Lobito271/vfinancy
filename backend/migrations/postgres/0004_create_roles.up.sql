-- 0004_create_roles.up.sql
-- Module 1: Authentication
-- A role is a named bundle of permissions scoped to a company.
-- Each company ships with a default set of system roles (admin, manager,
-- accountant, seller, warehouse, viewer) seeded in a later migration.
--
-- is_system: TRUE for the seeded roles. System roles cannot be deleted
-- or renamed; the application layer should refuse these operations.
--
-- FK: company_id ON DELETE RESTRICT. A company with roles cannot be deleted.
--
-- Soft delete with partial unique index on (company_id, code)
-- WHERE deleted_at IS NULL so codes like "admin" can be re-used after
-- a role is logically deleted.

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID         NOT NULL,
    code        VARCHAR(50)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_system   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  UUID,
    updated_by  UUID,

    CONSTRAINT ck_roles_name_nonblank
        CHECK (length(trim(name)) > 0),

    CONSTRAINT fk_roles_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_roles_company_code
    ON roles (company_id, code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_roles_company_active
    ON roles (company_id, is_active)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_roles_set_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
