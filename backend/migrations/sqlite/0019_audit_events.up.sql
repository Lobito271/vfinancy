-- 0019_audit_events.up.sql (SQLite)
-- Module 2: Administration — Application Audit Events
-- Higher-level application audit events (LOGIN, LOGOUT, PASSWORD_CHANGE,
-- CONFIG_UPDATE, etc.) complementing the low-level audit_logs table
-- (migration 0010) which captures DB-level INSERT/UPDATE/DELETE changes.
--
-- This table is append-only: the application inserts records but never
-- updates or deletes them. A trigger enforces this invariant.
--
-- SQLite notes: UUID -> TEXT, JSONB -> TEXT (JSON string literal),
-- INET -> TEXT.

CREATE TABLE audit_events (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT         NOT NULL,
    user_id      TEXT,
    session_id   TEXT,
    event_type   VARCHAR(50)  NOT NULL
        CHECK (event_type IN (
            'LOGIN', 'LOGOUT', 'LOGIN_FAILED', 'PASSWORD_CHANGE',
            'SESSION_LOCK', 'SESSION_UNLOCK', 'SESSION_EXPIRED',
            'CONFIG_UPDATE', 'USER_CREATE', 'USER_UPDATE', 'USER_DEACTIVATE',
            'ROLE_CREATE', 'ROLE_UPDATE', 'ROLE_DELETE',
            'BACKUP_CREATE', 'EXPORT_DATA'
        )),
    target_type  VARCHAR(100),
    target_id    TEXT,
    description  TEXT,
    metadata     TEXT         NOT NULL DEFAULT '{}',
    ip_address   TEXT,
    device       VARCHAR(100),
    occurred_at  TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

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

-- Append-only enforcement: reject UPDATE and DELETE.
-- SQLite does not support "UPDATE OR DELETE" in a single trigger
-- event, so one trigger per event is used.
CREATE TRIGGER trg_audit_events_no_update
    BEFORE UPDATE ON audit_events
BEGIN
    SELECT RAISE(ABORT, 'audit_events is append-only: UPDATE is not allowed');
END;

CREATE TRIGGER trg_audit_events_no_delete
    BEFORE DELETE ON audit_events
BEGIN
    SELECT RAISE(ABORT, 'audit_events is append-only: DELETE is not allowed');
END;
