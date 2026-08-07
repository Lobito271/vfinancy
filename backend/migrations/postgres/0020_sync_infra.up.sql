-- 0020_sync_infra.up.sql (PostgreSQL)
-- Part 2: device registry + replication plumbing for cloud sync.
--
-- Same model as the SQLite variant (see migrations/sqlite/0020): row-
-- level, last-writer-wins replication driven by a per-table watermark.
-- The server applies pushed rows/tombstones directly, so unlike the
-- client there are NO AFTER DELETE triggers here — every mutation that
-- reaches the server carries its tombstone explicitly.

CREATE TABLE sync_devices (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID        NOT NULL,
    name         VARCHAR(100) NOT NULL,
    platform     VARCHAR(30)  NOT NULL DEFAULT 'desktop',
    token        TEXT        NOT NULL,
    is_local     BOOLEAN     NOT NULL DEFAULT FALSE,
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
    last_seen_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_sync_devices_name_nonblank
        CHECK (length(trim(name)) > 0)
);

CREATE UNIQUE INDEX uq_sync_devices_local
    ON sync_devices (is_local)
    WHERE is_local = TRUE;

CREATE TABLE sync_cursors (
    device_id       UUID          NOT NULL,
    table_name      VARCHAR(100)  NOT NULL,
    last_updated_at TIMESTAMPTZ   NOT NULL DEFAULT 'epoch',

    PRIMARY KEY (device_id, table_name),

    CONSTRAINT fk_sync_cursors_device
        FOREIGN KEY (device_id) REFERENCES sync_devices(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE sync_conflicts (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id        UUID,
    table_name       VARCHAR(100) NOT NULL,
    record_id        TEXT         NOT NULL,
    operation        VARCHAR(20)  NOT NULL DEFAULT 'UPDATE'
        CHECK (operation IN ('UPDATE', 'DELETE')),
    local_updated_at TIMESTAMPTZ,
    remote_updated_at TIMESTAMPTZ,
    resolution       VARCHAR(20)  NOT NULL DEFAULT 'LOCAL_WON'
        CHECK (resolution IN ('LOCAL_WON', 'REMOTE_WON')),
    message          TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_sync_conflicts_device
        FOREIGN KEY (device_id) REFERENCES sync_devices(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_sync_conflicts_record
    ON sync_conflicts (table_name, record_id);

CREATE TABLE sync_tombstones (
    table_name VARCHAR(100) NOT NULL,
    record_id  TEXT         NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL,

    PRIMARY KEY (table_name, record_id)
);

CREATE INDEX idx_sync_tombstones_time
    ON sync_tombstones (table_name, updated_at);
