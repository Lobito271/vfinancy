-- 0007_create_user_roles.up.sql
-- Module 1: Authentication
-- Junction user × role, with an OPTIONAL branch scope.
--
-- Why a surrogate id and partial unique indexes? The natural key
-- would be (user_id, role_id, branch_id) but PostgreSQL PRIMARY KEY
-- columns are implicitly NOT NULL, which would forbid the legitimate
-- case of a company-wide role grant (no branch). Two partial unique
-- indexes express the constraint precisely:
--   * a user can have a given role at most once per branch
--   * a user can have a given role at most once with branch_id = NULL
--
-- This gives us a clean model:
--   ('seller', branch A) + ('seller', branch B) + ('seller', NULL)  ✓
-- while preventing:
--   ('seller', branch A) + ('seller', branch A)                    ✗
--   ('seller', NULL) + ('seller', NULL)                            ✗
--
-- expires_at: optional time-bound grants (e.g. "accountant for audit week").
--
-- FKs: CASCADE on user/role/branch — grants are pure linkage.

CREATE TABLE user_roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL,
    role_id     UUID         NOT NULL,
    branch_id   UUID,
    assigned_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    assigned_by UUID,
    expires_at  TIMESTAMPTZ,

    CONSTRAINT fk_user_roles_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_user_roles_role
        FOREIGN KEY (role_id) REFERENCES roles(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_user_roles_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT ck_user_roles_expires_after_assigned
        CHECK (expires_at IS NULL OR expires_at > assigned_at)
);

CREATE UNIQUE INDEX uq_user_roles_per_branch
    ON user_roles (user_id, role_id, branch_id)
    WHERE branch_id IS NOT NULL;

CREATE UNIQUE INDEX uq_user_roles_company_wide
    ON user_roles (user_id, role_id)
    WHERE branch_id IS NULL;

CREATE INDEX idx_user_roles_user_active
    ON user_roles (user_id)
    WHERE expires_at IS NULL;

CREATE INDEX idx_user_roles_role
    ON user_roles (role_id);
