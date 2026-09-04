-- 0001: Add notifications table if not exists
-- The table was originally in 0000 but may be missing in databases created before it was added.

CREATE TABLE IF NOT EXISTS notifications (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT         NOT NULL,
    type         VARCHAR(50)  NOT NULL,
    title        VARCHAR(200) NOT NULL,
    message      TEXT         NOT NULL,
    record_type  VARCHAR(50),
    record_id    TEXT,
    dedup_key    VARCHAR(64)  NOT NULL,
    read_at      TIMESTAMP,
    created_at   TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at   TIMESTAMP,

    CONSTRAINT fk_notifications_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT ck_notifications_type_nonblank
        CHECK (length(trim(type)) > 0),

    CONSTRAINT ck_notifications_title_nonblank
        CHECK (length(trim(title)) > 0),

    CONSTRAINT ck_notifications_dedup_nonblank
        CHECK (length(trim(dedup_key)) > 0),

    CONSTRAINT uq_notifications_company_type_dedup
        UNIQUE (company_id, type, dedup_key)
);

CREATE INDEX IF NOT EXISTS idx_notifications_company_unread
    ON notifications (company_id, read_at, created_at DESC);
