-- 0019_audit_events.up.sql
-- Module 2: Administration — Application Audit Events
-- Higher-level application audit events (LOGIN, LOGOUT, PASSWORD_CHANGE,
-- CONFIG_UPDATE, etc.) complementing the low-level audit_logs table
-- (migration 0010) which captures DB-level INSERT/UPDATE/DELETE changes.
--
-- This table is append-only: the application inserts records but never
-- updates or deletes them. A trigger enforces this invariant.

CREATE TABLE audit_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID         NOT NULL,
    user_id      UUID,
    session_id   UUID,
    event_type   VARCHAR(50)  NOT NULL
        CHECK (event_type IN (
            'LOGIN', 'LOGOUT', 'LOGIN_FAILED', 'PASSWORD_CHANGE',
            'SESSION_LOCK', 'SESSION_UNLOCK', 'SESSION_EXPIRED',
            'CONFIG_UPDATE', 'USER_CREATE', 'USER_UPDATE', 'USER_DEACTIVATE',
            'ROLE_CREATE', 'ROLE_UPDATE', 'ROLE_DELETE',
            'BACKUP_CREATE', 'EXPORT_DATA'
        )),
    target_type  VARCHAR(100),
    target_id    UUID,
    description  TEXT,
    metadata     JSONB        NOT NULL DEFAULT '{}',
    ip_address   INET,
    device       VARCHAR(100),
    occurred_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_audit_events_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_audit_events_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,

    CONSTRAINT fk_audit_events_session
        FOREIGN KEY (session_id) REFERENCES user_sessions(id)
        ON UPDATE CASCADE ON DELETE SET NULL,

    CONSTRAINT ck_audit_events_target_type_nonblank
        CHECK (target_type IS NULL OR length(trim(target_type)) > 0)
);

CREATE INDEX idx_audit_events_company_time
    ON audit_events (company_id, occurred_at DESC);

CREATE INDEX idx_audit_events_user
    ON audit_events (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX idx_audit_events_type
    ON audit_events (event_type, occurred_at DESC);

CREATE INDEX idx_audit_events_target
    ON audit_events (target_type, target_id)
    WHERE target_type IS NOT NULL AND target_id IS NOT NULL;

-- Append-only enforcement: reject UPDATE and DELETE
CREATE OR REPLACE FUNCTION reject_audit_events_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only: UPDATE and DELETE are not allowed';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_events_no_mutation
    BEFORE UPDATE OR DELETE ON audit_events
    FOR EACH ROW
    EXECUTE FUNCTION reject_audit_events_mutation();
