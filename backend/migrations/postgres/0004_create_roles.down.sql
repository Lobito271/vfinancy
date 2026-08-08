-- 0004_create_roles.down.sql

DROP TRIGGER IF EXISTS trg_roles_set_updated_at ON roles;
DROP TABLE IF EXISTS roles;
