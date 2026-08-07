-- 0001_create_companies.up.sql (SQLite)
-- Companies are the multi-tenant root. Every business table references
-- one (except globally-shared lookups like currencies and permissions).
--
-- SQLite notes:
--   - UUID columns are TEXT. The application always supplies ids, so
--     the randomblob default is only a safety net.
--   - Timestamps are INTEGER milliseconds since the Unix epoch
--     (matches the driver's _time_integer_format=unix_milli DSN).
--   - There is no set_updated_at trigger; updated_at is written by the
--     application on every UPDATE.

CREATE TABLE companies (
    id                         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    code                       VARCHAR(20)  NOT NULL,
    legal_name                 VARCHAR(200) NOT NULL,
    trade_name                 VARCHAR(200),
    tax_id                     VARCHAR(30)  NOT NULL,
    address                    TEXT,
    phone                      VARCHAR(30),
    email                      VARCHAR(200),
    country_code               CHAR(2)      NOT NULL DEFAULT 'PE',
    functional_currency_code    CHAR(3)      NOT NULL DEFAULT 'PEN',
    timezone                   VARCHAR(50)  NOT NULL DEFAULT 'America/Lima',
    fiscal_year_start_month    SMALLINT     NOT NULL DEFAULT 1
        CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
    is_active                  BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at                 TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at                 TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at                 TIMESTAMP,
    created_by                 TEXT,
    updated_by                 TEXT,

    CONSTRAINT ck_companies_legal_name_nonblank
        CHECK (length(trim(legal_name)) > 0),
    CONSTRAINT ck_companies_trade_name_nonblank_or_null
        CHECK (trade_name IS NULL OR length(trim(trade_name)) > 0)
);

CREATE UNIQUE INDEX uq_companies_code
    ON companies (code);

CREATE UNIQUE INDEX uq_companies_tax_id
    ON companies (tax_id);

CREATE INDEX idx_companies_active
    ON companies (is_active)
    WHERE deleted_at IS NULL;
