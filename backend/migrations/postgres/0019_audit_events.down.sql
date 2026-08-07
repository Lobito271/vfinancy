-- 0019_audit_events.down.sql

DROP TRIGGER IF EXISTS trg_audit_events_no_mutation ON audit_events;
DROP FUNCTION IF EXISTS reject_audit_events_mutation();
DROP TABLE IF EXISTS audit_events;
