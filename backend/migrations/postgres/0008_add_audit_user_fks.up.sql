-- 0008_add_audit_user_fks.up.sql
-- Module 1: Authentication
-- Add the deferred foreign keys on the audit columns (created_by /
-- updated_by) that reference users. These were intentionally not added
-- when each table was first created because users did not exist yet.
--
-- The strategy used here keeps the migration order linear and avoids
-- forward references. Each table already has its created_by / updated_by
-- columns in place; we now declare the FKs.
--
-- All FKs are ON DELETE SET NULL: if a user is soft- or hard-deleted,
-- we keep the historical rows but null out the reference. The audit_logs
-- table will continue to record the action under the actor's UUID
-- even after deletion (denormalized at write time).

ALTER TABLE companies
    ADD CONSTRAINT fk_companies_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    ADD CONSTRAINT fk_companies_updated_by
        FOREIGN KEY (updated_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE branches
    ADD CONSTRAINT fk_branches_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    ADD CONSTRAINT fk_branches_updated_by
        FOREIGN KEY (updated_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE roles
    ADD CONSTRAINT fk_roles_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    ADD CONSTRAINT fk_roles_updated_by
        FOREIGN KEY (updated_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE users
    ADD CONSTRAINT fk_users_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    ADD CONSTRAINT fk_users_updated_by
        FOREIGN KEY (updated_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE role_permissions
    ADD CONSTRAINT fk_role_permissions_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE user_roles
    ADD CONSTRAINT fk_user_roles_assigned_by
        FOREIGN KEY (assigned_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL;
