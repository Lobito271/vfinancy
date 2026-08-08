-- 0005_create_role_permissions.up.sql
-- Module 1: Authentication
-- Junction table role × permission. Composite primary key enforces
-- that a role cannot have the same permission twice.
--
-- ON DELETE CASCADE both ways: removing a role removes its permission
-- grants; removing a permission from the global catalog removes all
-- grants. This is safe because the junction carries no business data
-- beyond the linkage itself.
--
-- Audit columns limited to created_at + created_by: this is a pure
-- junction, no business state to track. No soft delete — when a role
-- or permission is removed, the rows are physically deleted.

CREATE TABLE role_permissions (
    role_id          UUID         NOT NULL,
    permission_code  VARCHAR(100) NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by       UUID,

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
