-- 0005_create_role_permissions.up.sql (SQLite)
-- Junction table role × permission. Composite primary key.

CREATE TABLE role_permissions (
    role_id          TEXT         NOT NULL,
    permission_code  VARCHAR(100) NOT NULL,
    created_at       TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by       TEXT,

    PRIMARY KEY (role_id, permission_code),

    CONSTRAINT fk_role_permissions_role
        FOREIGN KEY (role_id) REFERENCES roles(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_role_permissions_permission
        FOREIGN KEY (permission_code) REFERENCES permissions(code)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_role_permissions_permission
    ON role_permissions (permission_code);
