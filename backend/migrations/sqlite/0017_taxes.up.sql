-- 0017_taxes.up.sql (SQLite)
-- Module 2: Administration — Tax Catalog
-- Company-scoped tax definitions. Each company can customize its own
-- tax rates. Global templates (is_global = TRUE) are available to all
-- companies as starting points.
--
-- SQLite notes: NUMERIC(6,4) -> TEXT holding the exact decimal string
-- (the repository binds and scans it as text and parses it with
-- PercentageFromString). The rate range is enforced with a CAST check.

CREATE TABLE taxes (
    id             TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id     TEXT,
    code           VARCHAR(20)  NOT NULL,
    name           VARCHAR(100) NOT NULL,
    short_name     VARCHAR(30)  NOT NULL,
    country_code   VARCHAR(2)   NOT NULL,
    default_rate   TEXT         NOT NULL DEFAULT '0.0000'
        CHECK (CAST(default_rate AS REAL) >= 0 AND CAST(default_rate AS REAL) <= 1),
    is_inclusive   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_percentage  BOOLEAN      NOT NULL DEFAULT TRUE,
    category       VARCHAR(20)  NOT NULL DEFAULT 'sales'
        CHECK (category IN ('sales', 'purchase', 'income', 'municipal', 'other')),
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at     TIMESTAMP,

    CONSTRAINT fk_taxes_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_taxes_country
        FOREIGN KEY (country_code) REFERENCES countries(code)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT ck_taxes_code_nonblank
        CHECK (length(trim(code)) > 0),

    CONSTRAINT ck_taxes_name_nonblank
        CHECK (length(trim(name)) > 0)
);

CREATE UNIQUE INDEX uq_taxes_company_code
    ON taxes (COALESCE(company_id, '00000000-0000-0000-0000-000000000000'), code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_taxes_company_active
    ON taxes (company_id, is_active)
    WHERE deleted_at IS NULL AND is_active = TRUE;

CREATE INDEX idx_taxes_country
    ON taxes (country_code)
    WHERE deleted_at IS NULL;

-- Seed default tax templates (company_id = NULL means global template).
-- The unique index uq_taxes_company_code treats NULL company_id as the
-- zero UUID, so INSERT OR IGNORE makes this idempotent.
INSERT OR IGNORE INTO taxes (company_id, code, name, short_name, country_code, default_rate, is_inclusive, is_percentage, category, is_active) VALUES
    (NULL, 'IGV',      'Impuesto General a las Ventas', 'IGV',     'PE', '0.1800', FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVAP',     'Impuesto de Promoción Municipal','IVAP',    'PE', '0.0200', FALSE, TRUE, 'municipal',TRUE),
    (NULL, 'RENTA',    'Impuesto a la Renta',            'Renta',   'PE', '0.2950', FALSE, TRUE, 'income',   TRUE),
    (NULL, 'EXONERADO','Exonerado del IGV',              'Exo',     'PE', '0.0000', FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'GRATUITO', 'Operación Gratuita',             'Grat',    'PE', '0.0000', FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVA_MX',   'Impuesto al Valor Agregado',    'IVA',     'MX', '0.1600', FALSE, TRUE, 'sales',    FALSE),
    (NULL, 'IVA_CO',   'Impuesto sobre las Ventas',     'IVA',     'CO', '0.1900', FALSE, TRUE, 'sales',    FALSE);

-- Seed company-specific taxes for the demo company
INSERT OR IGNORE INTO taxes (company_id, code, name, short_name, country_code, default_rate, is_inclusive, is_percentage, category, is_active) VALUES
    ('00000000-0000-0000-0000-000000000001', 'IGV',      'Impuesto General a las Ventas', 'IGV',     'PE', '0.1800', FALSE, TRUE, 'sales',    TRUE),
    ('00000000-0000-0000-0000-000000000001', 'IVAP',     'Impuesto de Promoción Municipal','IVAP',    'PE', '0.0200', FALSE, TRUE, 'municipal',TRUE),
    ('00000000-0000-0000-0000-000000000001', 'RENTA',    'Impuesto a la Renta',            'Renta',   'PE', '0.2950', FALSE, TRUE, 'income',   TRUE),
    ('00000000-0000-0000-0000-000000000001', 'EXONERADO','Exonerado del IGV',              'Exo',     'PE', '0.0000', FALSE, TRUE, 'sales',    TRUE);
