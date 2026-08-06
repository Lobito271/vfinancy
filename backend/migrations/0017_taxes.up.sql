-- 0017_taxes.up.sql
-- Module 2: Administration — Tax Catalog
-- Company-scoped tax definitions. Each company can customize its own
-- tax rates. Global templates (is_global = TRUE) are available to all
-- companies as starting points.

CREATE TABLE taxes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID,
    code           VARCHAR(20)  NOT NULL,
    name           VARCHAR(100) NOT NULL,
    short_name     VARCHAR(30)  NOT NULL,
    country_code   VARCHAR(2)   NOT NULL,
    default_rate   NUMERIC(6,4) NOT NULL DEFAULT 0.0000
        CHECK (default_rate >= 0 AND default_rate <= 1),
    is_inclusive   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_percentage  BOOLEAN      NOT NULL DEFAULT TRUE,
    category       VARCHAR(20)  NOT NULL DEFAULT 'sales'
        CHECK (category IN ('sales', 'purchase', 'income', 'municipal', 'other')),
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,

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

CREATE TRIGGER trg_taxes_set_updated_at
    BEFORE UPDATE ON taxes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Seed default tax templates (company_id = NULL means global template)
INSERT INTO taxes (company_id, code, name, short_name, country_code, default_rate, is_inclusive, is_percentage, category, is_active) VALUES
    (NULL, 'IGV',      'Impuesto General a las Ventas', 'IGV',     'PE', 0.1800, FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVAP',     'Impuesto de Promoción Municipal','IVAP',    'PE', 0.0200, FALSE, TRUE, 'municipal',TRUE),
    (NULL, 'RENTA',    'Impuesto a la Renta',            'Renta',   'PE', 0.2950, FALSE, TRUE, 'income',   TRUE),
    (NULL, 'EXONERADO','Exonerado del IGV',              'Exo',     'PE', 0.0000, FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'GRATUITO', 'Operación Gratuita',             'Grat',    'PE', 0.0000, FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVA_MX',   'Impuesto al Valor Agregado',    'IVA',     'MX', 0.1600, FALSE, TRUE, 'sales',    FALSE),
    (NULL, 'IVA_CO',   'Impuesto sobre las Ventas',     'IVA',     'CO', 0.1900, FALSE, TRUE, 'sales',    FALSE)
ON CONFLICT DO NOTHING;

-- Seed company-specific taxes for the demo company
INSERT INTO taxes (company_id, code, name, short_name, country_code, default_rate, is_inclusive, is_percentage, category, is_active) VALUES
    ('00000000-0000-0000-0000-000000000001', 'IGV',      'Impuesto General a las Ventas', 'IGV',     'PE', 0.1800, FALSE, TRUE, 'sales',    TRUE),
    ('00000000-0000-0000-0000-000000000001', 'IVAP',     'Impuesto de Promoción Municipal','IVAP',    'PE', 0.0200, FALSE, TRUE, 'municipal',TRUE),
    ('00000000-0000-0000-0000-000000000001', 'RENTA',    'Impuesto a la Renta',            'Renta',   'PE', 0.2950, FALSE, TRUE, 'income',   TRUE),
    ('00000000-0000-0000-0000-000000000001', 'EXONERADO','Exonerado del IGV',              'Exo',     'PE', 0.0000, FALSE, TRUE, 'sales',    TRUE)
ON CONFLICT DO NOTHING;
