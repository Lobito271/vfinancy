-- 0001_create_companies.up.sql
-- Module 1: Authentication
-- Companies are the multi-tenant root. Every business table references one
-- (except globally-shared lookups like currencies and permissions).
--
-- Defaults:
--   - functional_currency_code defaults to 'PEN' (Peru, our launch market).
--   - fiscal_year_start_month defaults to 1 (January). Companies in other
--     jurisdictions can override.
--   - is_active defaults to TRUE; soft delete via deleted_at.
--
-- Audit columns are kept in sync by the set_updated_at() trigger created in 0000.
--
-- Indexes:
--   - PK on id
--   - UNIQUE on code (no soft-delete partial: codes are short stable
--     identifiers, never reused)
--   - UNIQUE on tax_id (RUC / EIN / NIT) per company
--   - Partial index on is_active WHERE deleted_at IS NULL for fast
--     "active companies" lookups.

CREATE TABLE companies (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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

    CONSTRAINT ck_companies_legal_name_nonblank
        CHECK (length(trim(legal_name)) > 0),
    CONSTRAINT ck_companies_trade_name_nonblank_or_null
        CHECK (trade_name IS NULL OR length(trim(trade_name)) > 0),

    created_at                 TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at                 TIMESTAMPTZ,
    created_by                 UUID,
    updated_by                 UUID
);

CREATE UNIQUE INDEX uq_companies_code
    ON companies (code);

CREATE UNIQUE INDEX uq_companies_tax_id
    ON companies (tax_id);

CREATE INDEX idx_companies_active
    ON companies (is_active)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_companies_set_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
