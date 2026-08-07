-- 0019_audit_events.down.sql (SQLite)

DROP TRIGGER IF EXISTS trg_audit_events_no_update;
DROP TRIGGER IF EXISTS trg_audit_events_no_delete;
DROP TABLE IF EXISTS audit_events;
