-- 0010_create_audit_logs.up.sql
-- Module 1: Authentication
-- Global, append-only audit log. Every meaningful mutation in the
-- application produces a row here. The application is responsible for
-- writing the row in the same transaction as the mutation, capturing
-- the pre/post state as JSONB.
--
-- Storage considerations:
--   - This table grows fast. A future migration will convert it to a
--     RANGE-partitioned table on (occurred_at, monthly buckets). The
--     current shape is compatible with that conversion: PK is
--     (id, occurred_at) so the partition key is part of the PK, and
--     there is no UNIQUE constraint that would span partitions.
--   - For now, plain table + good indexes is enough for the v1 volume.
--
-- field-level sensitivity: the application MUST scrub secrets
-- (password_hash, tokens) before writing. We do not enforce this in
-- the database because GIN-trigram-style checks on JSONB are
-- expensive; an audit review or CI check covers it.
--
-- No soft delete. No UPDATE allowed by the application. The trigger
-- below rejects any UPDATE or DELETE to enforce append-only at the DB
-- level (defense in depth).

CREATE TABLE audit_logs (
    id              UUID         NOT NULL DEFAULT gen_random_uuid(),
    occurred_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    company_id      UUID         NOT NULL,
    user_id         UUID,
    table_name      VARCHAR(100) NOT NULL,
    record_id       UUID,
    action          VARCHAR(30)  NOT NULL,
    old_value       JSONB,
    new_value       JSONB,
    changed_fields  TEXT[],
    ip_address      INET,
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

-- Indexes — designed for the three main query patterns.
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
CREATE OR REPLACE FUNCTION audit_logs_forbid_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is append-only; % is not allowed', TG_OP;
END;
$$;

CREATE TRIGGER trg_audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION audit_logs_forbid_mutation();

CREATE TRIGGER trg_audit_logs_no_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION audit_logs_forbid_mutation();

-- No "set_updated_at" trigger: this table does not have an updated_at
-- column, and an UPDATE is not legal anyway.
