-- 0009_create_login_history.up.sql (SQLite)
-- Append-only audit of every authentication attempt, successful or not.
-- No soft delete — this is a forensic log. Retention policy is set by
-- the application (typically 12–24 months; older rows can be pruned
-- by a maintenance job in batches).
--
-- SQLite notes: UUID -> TEXT, TIMESTAMPTZ -> INTEGER ms, INET -> TEXT.

CREATE TABLE login_history (
    id                   TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id              TEXT,
    username_attempted   VARCHAR(100) NOT NULL,
    success              BOOLEAN      NOT NULL,
    failure_reason       VARCHAR(100),
    ip_address           TEXT,
    user_agent           TEXT,
    occurred_at          TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

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
