-- 0013_user_profiles.up.sql (SQLite)
-- Module 1: Authentication — User Profiles
-- Extended profile data per user. One-to-one relationship with users.
-- Stores UI preferences (theme, language, date/number format) and
-- optional avatar reference.

CREATE TABLE user_profiles (
    id             TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id        TEXT         NOT NULL,
    avatar_url     VARCHAR(500),
    theme          VARCHAR(20)  NOT NULL DEFAULT 'system'
        CHECK (theme IN ('light', 'dark', 'system')),
    language       VARCHAR(10)  NOT NULL DEFAULT 'es-PE',
    date_format    VARCHAR(30)  NOT NULL DEFAULT 'DD/MM/YYYY',
    number_format  VARCHAR(30)  NOT NULL DEFAULT 'es-PE',
    decimal_places INTEGER      NOT NULL DEFAULT 2
        CHECK (decimal_places BETWEEN 0 AND 6),
    timezone       VARCHAR(50)  NOT NULL DEFAULT 'America/Lima',
    created_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_profiles_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT ck_profiles_language_nonblank
        CHECK (length(trim(language)) > 0)
);

CREATE UNIQUE INDEX uq_profiles_user
    ON user_profiles (user_id);
