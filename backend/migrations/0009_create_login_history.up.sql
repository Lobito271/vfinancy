-- 0009_create_login_history.up.sql
-- Module 1: Authentication
-- Append-only audit of every authentication attempt, successful or not.
-- No soft delete — this is a forensic log. Retention policy is set by
-- the application (typically 12–24 months; older rows can be pruned
-- by a maintenance job in batches).
--
-- user_id is nullable: failed attempts may target a username that does
-- not exist, in which case user_id is NULL but username_attempted is
-- captured. ON DELETE SET NULL preserves the row if the user is removed.
--
-- Indexes:
--   - (user_id, occurred_at DESC) — recent activity for a user
--   - (ip_address, occurred_at DESC) — abuse detection
--   - (occurred_at DESC) — global timeline
--   - (success, occurred_at DESC) partial on failures — incident review
--
-- We do NOT use TIMESTAMP without timezone. occurred_at is TIMESTAMPTZ;
-- the application writes UTC.

CREATE TABLE login_history (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID,
    username_attempted   VARCHAR(100) NOT NULL,
    success              BOOLEAN      NOT NULL,
    failure_reason       VARCHAR(100),
    ip_address           INET,
    user_agent           TEXT,
    occurred_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_login_history_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,

    CONSTRAINT ck_login_history_failure_consistency
        CHECK (
            (success = TRUE AND failure_reason IS NULL)
            OR (success = FALSE AND failure_reason IS NOT NULL)
        )
);

CREATE INDEX idx_login_history_user_time
    ON login_history (user_id, occurred_at DESC);

CREATE INDEX idx_login_history_ip_time
    ON login_history (ip_address, occurred_at DESC)
    WHERE ip_address IS NOT NULL;

CREATE INDEX idx_login_history_time
    ON login_history (occurred_at DESC);

CREATE INDEX idx_login_history_failures
    ON login_history (occurred_at DESC)
    WHERE success = FALSE;
