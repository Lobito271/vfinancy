-- 0008_add_audit_user_fks.up.sql (SQLite)
-- PostgreSQL added the deferred FKs on the audit columns (created_by /
-- updated_by -> users) with ALTER TABLE ADD CONSTRAINT. SQLite cannot
-- add constraints after table creation, so these FKs are declared
-- inline in the respective CREATE TABLE statements instead. This
-- migration is a deliberate no-op that keeps the version history
-- identical across dialects.

SELECT 1;
