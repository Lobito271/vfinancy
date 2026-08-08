-- 0004_create_roles.up.sql (SQLite)
-- A role is a named bundle of permissions scoped to a company.

CREATE TABLE roles (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id  TEXT         NOT NULL,
    code        VARCHAR(50)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_system   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at  TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at  TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at  TIMESTAMP,
    created_by  TEXT,
    updated_by  TEXT,

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
