-- 0007_create_user_roles.up.sql (SQLite)
-- Junction user × role, with an optional branch scope. Partial unique
-- indexes express the (user, role, branch) cardinality precisely.

CREATE TABLE user_roles (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id     TEXT         NOT NULL,
    role_id     TEXT         NOT NULL,
    branch_id   TEXT,
    assigned_at TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    assigned_by TEXT,
    expires_at  TIMESTAMP,

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
