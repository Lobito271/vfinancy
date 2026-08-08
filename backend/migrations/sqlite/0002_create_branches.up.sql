-- 0002_create_branches.up.sql (SQLite)
-- Branches are physical locations belonging to a company. Partial
-- unique indexes work in SQLite exactly as in PostgreSQL.

CREATE TABLE branches (
    id              TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id      TEXT         NOT NULL,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    address         TEXT,
    phone           VARCHAR(30),
    email           VARCHAR(200),
    is_default      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at      TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at      TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at      TIMESTAMP,
    created_by      TEXT,
    updated_by      TEXT,

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
