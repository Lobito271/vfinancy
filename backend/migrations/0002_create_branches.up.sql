-- 0002_create_branches.up.sql
-- Module 1: Authentication
-- Branches are physical locations belonging to a company. They provide
-- scope for user roles and for sales/purchases/treasury later.
--
-- FK to companies: ON DELETE RESTRICT — you cannot delete a company that
-- still has branches; archive them first (is_active = FALSE) and let the
-- downstream tables deal with historical rows.
--
-- Only one branch per company can be the default. Enforced with a partial
-- unique index on (company_id) WHERE is_default = TRUE — partial because
-- a NULL is_default (false) row is allowed to coexist.
--
-- Soft delete: branch codes (like "001", "002") are NOT unique across time
-- once soft-deleted, hence the UNIQUE index is partial on deleted_at IS NULL.

CREATE TABLE branches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID         NOT NULL,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    address         TEXT,
    phone           VARCHAR(30),
    email           VARCHAR(200),
    is_default      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      UUID,
    updated_by      UUID,

    CONSTRAINT fk_branches_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_branches_company_code
    ON branches (company_id, code)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_branches_company_default
    ON branches (company_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_branches_company_active
    ON branches (company_id, is_active)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_branches_set_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
