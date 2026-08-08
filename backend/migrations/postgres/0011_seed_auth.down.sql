-- 0011_seed_auth.down.sql
-- Reverse 0011: remove the seeded admin user, then the roles, then the
-- branch and company. The global permissions catalog is intentionally NOT
-- removed: deleting it would orphan audit_logs.action values that are
-- CHECK-constrained. Permissions are part of the global catalog, not a
-- tenant-scoped resource.
--
-- If you really want to wipe everything, run 0003_create_permissions.down.sql
-- manually after this.

DELETE FROM user_roles
 WHERE user_id = '00000000-0000-0000-0000-0000000000aa';

DELETE FROM users
 WHERE id = '00000000-0000-0000-0000-0000000000aa';

DELETE FROM role_permissions
 WHERE role_id IN (
    '00000000-0000-0000-0000-0000000000a1',
    '00000000-0000-0000-0000-0000000000a2',
    '00000000-0000-0000-0000-0000000000a3',
    '00000000-0000-0000-0000-0000000000a4',
    '00000000-0000-0000-0000-0000000000a5',
    '00000000-0000-0000-0000-0000000000a6'
 );

DELETE FROM roles
 WHERE id IN (
    '00000000-0000-0000-0000-0000000000a1',
    '00000000-0000-0000-0000-0000000000a2',
    '00000000-0000-0000-0000-0000000000a3',
    '00000000-0000-0000-0000-0000000000a4',
    '00000000-0000-0000-0000-0000000000a5',
    '00000000-0000-0000-0000-0000000000a6'
 );

DELETE FROM branches
 WHERE id = '00000000-0000-0000-0000-000000000001';

DELETE FROM companies
 WHERE id = '00000000-0000-0000-0000-000000000001';
