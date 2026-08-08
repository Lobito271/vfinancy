-- 0008_add_audit_user_fks.down.sql

ALTER TABLE companies
    DROP CONSTRAINT IF EXISTS fk_companies_created_by,
    DROP CONSTRAINT IF EXISTS fk_companies_updated_by;

ALTER TABLE branches
    DROP CONSTRAINT IF EXISTS fk_branches_created_by,
    DROP CONSTRAINT IF EXISTS fk_branches_updated_by;

ALTER TABLE roles
    DROP CONSTRAINT IF EXISTS fk_roles_created_by,
    DROP CONSTRAINT IF EXISTS fk_roles_updated_by;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS fk_users_created_by,
    DROP CONSTRAINT IF EXISTS fk_users_updated_by;

ALTER TABLE role_permissions
    DROP CONSTRAINT IF EXISTS fk_role_permissions_created_by;

ALTER TABLE user_roles
    DROP CONSTRAINT IF EXISTS fk_user_roles_assigned_by;
