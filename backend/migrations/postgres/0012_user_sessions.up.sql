-- 0012_user_sessions.up.sql
-- Module 1: Authentication — Session Management
-- Tracks active user sessions. LOCAL ONLY (no JWT).
-- Sessions are validated by looking up the token in this table.
-- Session timeout is enforced by the application (expires_at).
-- Session lock is enforced by the application (locked_at IS NOT NULL).

CREATE TABLE user_sessions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID         NOT NULL,
    token             VARCHAR(128) NOT NULL,
    ip_address        INET,
    user_agent        TEXT,
    device            VARCHAR(100),
    is_active         BOOLEAN      NOT NULL DEFAULT TRUE,
    locked_at         TIMESTAMPTZ,
    locked_reason     VARCHAR(200),
    expires_at        TIMESTAMPTZ  NOT NULL,
    last_activity_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_sessions_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT ck_sessions_token_nonblank
        CHECK (length(trim(token)) > 0),

    CONSTRAINT ck_sessions_expires_future
        CHECK (expires_at > created_at)
);

CREATE UNIQUE INDEX uq_sessions_token
    ON user_sessions (token);

CREATE INDEX idx_sessions_user_active
    ON user_sessions (user_id, is_active)
    WHERE is_active = TRUE;

CREATE INDEX idx_sessions_expires
    ON user_sessions (expires_at)
    WHERE is_active = TRUE;
