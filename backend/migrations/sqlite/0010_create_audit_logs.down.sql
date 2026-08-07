-- 0010_create_audit_logs.down.sql (SQLite)

DROP TRIGGER IF EXISTS trg_audit_logs_no_update;
DROP TRIGGER IF EXISTS trg_audit_logs_no_delete;
DROP TABLE IF EXISTS audit_logs;
