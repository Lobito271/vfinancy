-- 0010_create_audit_logs.up.sql (SQLite)
-- Global, append-only audit log. Every meaningful mutation in the
-- application produces a row here. The application is responsible for
-- writing the row in the same transaction as the mutation, capturing
-- the pre/post state as TEXT (JSON-encoded).
--
-- SQLite notes:
--   - UUID -> TEXT, TIMESTAMPTZ -> INTEGER ms, JSONB -> TEXT (stored as
--     a JSON string literal, as written by the application), TEXT[] ->
--     TEXT (SQL array literal like '{a,b}'), INET -> TEXT.
--   - No UPDATE/DELETE allowed by the application. The triggers below
--     reject any UPDATE or DELETE to enforce append-only at the DB
--     level (defense in depth).

CREATE TABLE audit_logs (
    id              TEXT         NOT NULL DEFAULT (lower(hex(randomblob(16)))),
    occurred_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    company_id      TEXT         NOT NULL,
    user_id         TEXT,
    table_name      VARCHAR(100) NOT NULL,
    record_id       TEXT,
    action          VARCHAR(30)  NOT NULL,
    old_value       TEXT,
    new_value       TEXT,
    changed_fields  TEXT,
    ip_address      TEXT,
    user_agent      TEXT,
    device          VARCHAR(200),

    PRIMARY KEY (id, occurred_at),

    CONSTRAINT fk_audit_logs_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT fk_audit_logs_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,

    CONSTRAINT ck_audit_logs_action
        CHECK (action IN (
            'INSERT', 'UPDATE', 'DELETE', 'HARD_DELETE',
            'LOGIN', 'LOGOUT', 'LOGIN_FAILED',
            'APPROVE', 'REJECT', 'CANCEL',
            'CONCILIATE', 'CLOSE_PERIOD', 'REOPEN_PERIOD',
            'EXPORT', 'PRINT', 'SEND'
        ))
);

CREATE INDEX idx_audit_logs_company_time
    ON audit_logs (company_id, occurred_at DESC);

CREATE INDEX idx_audit_logs_record
    ON audit_logs (table_name, record_id, occurred_at DESC)
    WHERE record_id IS NOT NULL;

CREATE INDEX idx_audit_logs_user_time
    ON audit_logs (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX idx_audit_logs_action_time
    ON audit_logs (action, occurred_at DESC);

CREATE INDEX idx_audit_logs_company_action_time
    ON audit_logs (company_id, action, occurred_at DESC);

-- Defense in depth: append-only enforcement.
CREATE TRIGGER trg_audit_logs_no_update
    BEFORE UPDATE ON audit_logs
BEGIN
    SELECT RAISE(ABORT, 'audit_logs is append-only; UPDATE is not allowed');
END;

CREATE TRIGGER trg_audit_logs_no_delete
    BEFORE DELETE ON audit_logs
BEGIN
    SELECT RAISE(ABORT, 'audit_logs is append-only; DELETE is not allowed');
END;

-- No "set_updated_at" trigger: this table does not have an updated_at
-- column, and an UPDATE is not legal anyway.
