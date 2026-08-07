-- 0003_create_permissions.up.sql (SQLite)
-- Global catalog of permission keys. code is the primary key.

CREATE TABLE permissions (
    code        VARCHAR(100) PRIMARY KEY,
    module      VARCHAR(50)  NOT NULL,
    action      VARCHAR(50)  NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
);

CREATE INDEX idx_permissions_module
    ON permissions (module);
