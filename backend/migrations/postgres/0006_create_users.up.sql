-- 0006_create_users.up.sql
-- Module 1: Authentication
-- Users authenticate against this table. Passwords are stored as
-- Argon2id hashes produced by the application; the format is
-- $argon2id$v=19$m=...,t=...,p=...$<salt>$<hash> — a single TEXT column
-- is enough, the algorithm + parameters are self-describing.
--
-- Lockout: failed_login_attempts counts consecutive failures.
-- locked_until is set to a future timestamp when the threshold is hit.
-- Reset to 0 on a successful login. We do NOT hard-delete on lockout
-- — the field clears after the lockout window.
--
-- FK to companies: ON DELETE RESTRICT — you cannot delete a company
-- while users reference it. Archive the company first.
-- FK to branches (default_branch_id): ON DELETE SET NULL — a user
-- losing their default branch is fine; they keep existing permissions.
--
-- Soft delete: usernames and emails are unique per company only among
-- active (non-deleted) records. A retired user's email can be reused
-- for a future hire.

CREATE TABLE users (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id               UUID         NOT NULL,
    default_branch_id        UUID,
    username                 VARCHAR(100) NOT NULL,
    email                    VARCHAR(200) NOT NULL,
    full_name                VARCHAR(200) NOT NULL,
    password_hash            TEXT         NOT NULL,
    must_change_password     BOOLEAN      NOT NULL DEFAULT TRUE,
    failed_login_attempts    INTEGER      NOT NULL DEFAULT 0
        CHECK (failed_login_attempts >= 0),
    locked_until             TIMESTAMPTZ,
    last_login_at            TIMESTAMPTZ,
    last_login_ip           INET,
    is_active                BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ,
    created_by               UUID,
    updated_by               UUID,

    CONSTRAINT fk_users_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT fk_users_default_branch
        FOREIGN KEY (default_branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,

    CONSTRAINT ck_users_username_lowercase
        CHECK (username = LOWER(username) AND username ~ '^[a-z0-9._-]+$'),

    CONSTRAINT ck_users_email_lowercase
        CHECK (email = LOWER(email) AND email LIKE '%@%'),

    CONSTRAINT ck_users_full_name_nonblank
        CHECK (length(trim(full_name)) > 0),

    CONSTRAINT ck_users_failed_attempts_cap
        CHECK (failed_login_attempts <= 1000),

    CONSTRAINT ck_users_lockout_window
        CHECK (locked_until IS NULL OR locked_until > created_at)
);

-- Uniqueness on (company_id, username) / (company_id, email) is per the
-- raw stored value, which the CHECK constraint above guarantees to be
-- already-lowercase. No LOWER() call needed in the index predicate.
CREATE UNIQUE INDEX uq_users_company_username
    ON users (company_id, username)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_users_company_email
    ON users (company_id, email)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_users_company_active
    ON users (company_id, is_active)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_users_locked
    ON users (locked_until)
    WHERE locked_until IS NOT NULL;

CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
