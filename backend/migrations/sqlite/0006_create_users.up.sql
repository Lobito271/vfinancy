-- 0006_create_users.up.sql (SQLite)
-- Users authenticate against this table. The username regex check is
-- expressed with GLOB (SQLite has no POSIX regex by default).

CREATE TABLE users (
    id                       TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id               TEXT         NOT NULL,
    default_branch_id        TEXT,
    username                 VARCHAR(100) NOT NULL,
    email                    VARCHAR(200) NOT NULL,
    full_name                VARCHAR(200) NOT NULL,
    password_hash            TEXT         NOT NULL,
    must_change_password     BOOLEAN      NOT NULL DEFAULT TRUE,
    failed_login_attempts    INTEGER      NOT NULL DEFAULT 0
        CHECK (failed_login_attempts >= 0),
    locked_until             TIMESTAMP,
    last_login_at            TIMESTAMP,
    last_login_ip            TEXT,
    is_active                BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at               TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at               TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at               TIMESTAMP,
    created_by               TEXT,
    updated_by               TEXT,

    CONSTRAINT fk_users_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT fk_users_default_branch
        FOREIGN KEY (default_branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,

    CONSTRAINT ck_users_username_lowercase
        CHECK (username = LOWER(username) AND length(username) > 0 AND username GLOB '[a-z0-9._-]*'),

    CONSTRAINT ck_users_email_lowercase
        CHECK (email = LOWER(email) AND email LIKE '%@%'),

    CONSTRAINT ck_users_full_name_nonblank
        CHECK (length(trim(full_name)) > 0),

    CONSTRAINT ck_users_failed_attempts_cap
        CHECK (failed_login_attempts <= 1000),

    CONSTRAINT ck_users_lockout_window
        CHECK (locked_until IS NULL OR locked_until > created_at)
);

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
