-- 0020_sync_infra.up.sql (SQLite)
-- Part 2: device registry + replication plumbing for cloud sync.
--
-- Model: row-level, last-writer-wins (LWW) replication driven by a
-- per-table update watermark.
--   - sync_devices    : the device registry. Exactly one row has
--                       is_local = TRUE (this desktop install).
--   - sync_cursors    : per (device, table) watermark, ms since epoch,
--                       of the last change processed.
--   - sync_conflicts  : audit of every LWW conflict resolved; the
--                       losing side and both timestamps are kept.
--   - sync_tombstones : hard-delete markers so a deletion propagates
--                       to every device after the row itself is gone.
--
-- Row changes travel as "all rows whose time column > cursor" in both
-- directions (push and pull use the same generic query). Hard deletes
-- are captured by the AFTER DELETE triggers below; there are no
-- INSERT/UPDATE triggers, so applying a pulled row never re-queues
-- anything at the trigger level (a tombstone re-echo is idempotent).
--
-- Timestamps are INTEGER milliseconds since the Unix epoch (matches the
-- driver's _time_integer_format=unix_milli DSN option).
--
-- The cloud mirror (migrations/postgres/0020) declares the same tables
-- but WITHOUT the triggers: the server applies pushes directly and
-- records tombstones explicitly.

CREATE TABLE sync_devices (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT NOT NULL,
    name         VARCHAR(100) NOT NULL,
    platform     VARCHAR(30)  NOT NULL DEFAULT 'desktop',
    token        TEXT NOT NULL,
    is_local     BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    last_seen_at TIMESTAMP,
    created_at   TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT ck_sync_devices_name_nonblank
        CHECK (length(trim(name)) > 0)
);

CREATE UNIQUE INDEX uq_sync_devices_local
    ON sync_devices (is_local)
    WHERE is_local = TRUE;

CREATE TABLE sync_cursors (
    device_id       TEXT          NOT NULL,
    table_name      VARCHAR(100)  NOT NULL,
    last_updated_at TIMESTAMP     NOT NULL DEFAULT 0,

    PRIMARY KEY (device_id, table_name),

    CONSTRAINT fk_sync_cursors_device
        FOREIGN KEY (device_id) REFERENCES sync_devices(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE sync_conflicts (
    id                TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    device_id         TEXT,
    table_name        VARCHAR(100) NOT NULL,
    record_id         TEXT NOT NULL,
    operation         VARCHAR(20)  NOT NULL DEFAULT 'UPDATE'
        CHECK (operation IN ('UPDATE', 'DELETE')),
    local_updated_at  TIMESTAMP,
    remote_updated_at TIMESTAMP,
    resolution        VARCHAR(20)  NOT NULL DEFAULT 'LOCAL_WON'
        CHECK (resolution IN ('LOCAL_WON', 'REMOTE_WON')),
    message           TEXT,
    created_at        TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_sync_conflicts_device
        FOREIGN KEY (device_id) REFERENCES sync_devices(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_sync_conflicts_record
    ON sync_conflicts (table_name, record_id);

CREATE TABLE sync_tombstones (
    table_name VARCHAR(100) NOT NULL,
    record_id  TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    PRIMARY KEY (table_name, record_id)
);

CREATE INDEX idx_sync_tombstones_time
    ON sync_tombstones (table_name, updated_at);

-- =========================================================================
-- Hard-delete capture
-- SQLite record_id for a single-column PK is the id value itself; for
-- composite PKs (role_permissions) it is a JSON array literal produced
-- by json_array(), matching what the replication code builds.
-- =========================================================================

CREATE TRIGGER trg_companies_sync_delete AFTER DELETE ON companies BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('companies', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_branches_sync_delete AFTER DELETE ON branches BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('branches', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_roles_sync_delete AFTER DELETE ON roles BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('roles', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_users_sync_delete AFTER DELETE ON users BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('users', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_user_roles_sync_delete AFTER DELETE ON user_roles BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('user_roles', OLD.id, OLD.assigned_at);
END;

CREATE TRIGGER trg_user_profiles_sync_delete AFTER DELETE ON user_profiles BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('user_profiles', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_user_sessions_sync_delete AFTER DELETE ON user_sessions BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('user_sessions', OLD.id, OLD.last_activity_at);
END;

CREATE TRIGGER trg_application_settings_sync_delete AFTER DELETE ON application_settings BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('application_settings', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_taxes_sync_delete AFTER DELETE ON taxes BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('taxes', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_currencies_sync_delete AFTER DELETE ON currencies BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('currencies', OLD.code, OLD.updated_at);
END;

CREATE TRIGGER trg_countries_sync_delete AFTER DELETE ON countries BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('countries', OLD.code, OLD.created_at);
END;

CREATE TRIGGER trg_permissions_sync_delete AFTER DELETE ON permissions BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('permissions', OLD.code, OLD.created_at);
END;

CREATE TRIGGER trg_role_permissions_sync_delete AFTER DELETE ON role_permissions BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('role_permissions', json_array(OLD.role_id, OLD.permission_code), OLD.created_at);
END;
