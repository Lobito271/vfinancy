-- 0020_sync_infra.down.sql (SQLite)

DROP TRIGGER IF EXISTS trg_role_permissions_sync_delete;
DROP TRIGGER IF EXISTS trg_permissions_sync_delete;
DROP TRIGGER IF EXISTS trg_countries_sync_delete;
DROP TRIGGER IF EXISTS trg_currencies_sync_delete;
DROP TRIGGER IF EXISTS trg_taxes_sync_delete;
DROP TRIGGER IF EXISTS trg_application_settings_sync_delete;
DROP TRIGGER IF EXISTS trg_user_sessions_sync_delete;
DROP TRIGGER IF EXISTS trg_user_profiles_sync_delete;
DROP TRIGGER IF EXISTS trg_user_roles_sync_delete;
DROP TRIGGER IF EXISTS trg_users_sync_delete;
DROP TRIGGER IF EXISTS trg_roles_sync_delete;
DROP TRIGGER IF EXISTS trg_branches_sync_delete;
DROP TRIGGER IF EXISTS trg_companies_sync_delete;

DROP TABLE IF EXISTS sync_tombstones;
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS sync_cursors;
DROP TABLE IF EXISTS sync_devices;
