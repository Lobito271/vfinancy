-- 0015_currencies.up.sql
-- Module 2: Administration — Currency Catalog
-- ISO 4217 currency catalog. Global (not company-scoped).
-- Supports future currencies without schema changes.

CREATE TABLE currencies (
    code           VARCHAR(3) PRIMARY KEY,
    symbol         VARCHAR(10)  NOT NULL,
    name           VARCHAR(100) NOT NULL,
    decimal_places INTEGER      NOT NULL DEFAULT 2
        CHECK (decimal_places BETWEEN 0 AND 6),
    type           VARCHAR(20)  NOT NULL DEFAULT 'fiat'
        CHECK (type IN ('fiat', 'crypto', 'commodity')),
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_currencies_code_uppercase
        CHECK (code = UPPER(code) AND length(code) = 3),

    CONSTRAINT ck_currencies_name_nonblank
        CHECK (length(trim(name)) > 0)
);

CREATE INDEX idx_currencies_active
    ON currencies (is_active)
    WHERE is_active = TRUE;

CREATE TRIGGER trg_currencies_set_updated_at
    BEFORE UPDATE ON currencies
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Seed default currencies
INSERT INTO currencies (code, symbol, name, decimal_places, type, is_active) VALUES
    ('PEN', 'S/',  'Sol peruano',        2, 'fiat', TRUE),
    ('USD', '$',   'Dólar estadounidense', 2, 'fiat', TRUE),
    ('EUR', '€',   'Euro',               2, 'fiat', FALSE),
    ('MXN', 'MX$', 'Peso mexicano',      2, 'fiat', FALSE),
    ('COP', 'COP$', 'Peso colombiano',   2, 'fiat', FALSE),
    ('CLP', 'CLP$', 'Peso chileno',      0, 'fiat', FALSE),
    ('BRL', 'R$',  'Real brasileño',     2, 'fiat', FALSE)
ON CONFLICT (code) DO NOTHING;
