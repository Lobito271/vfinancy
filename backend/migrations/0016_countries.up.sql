-- 0016_countries.up.sql
-- Module 2: Administration — Country Catalog
-- ISO 3166-1 alpha-2 country reference. Global (not company-scoped).
-- Provides locale, default currency, and document type labels.

CREATE TABLE countries (
    code                    VARCHAR(2) PRIMARY KEY,
    name                    VARCHAR(100) NOT NULL,
    locale                  VARCHAR(10)  NOT NULL DEFAULT 'es-PE',
    currency_code           VARCHAR(3)   NOT NULL DEFAULT 'PEN',
    tax_id_label            VARCHAR(50)  NOT NULL DEFAULT 'Tax ID',
    personal_id_label       VARCHAR(50)  NOT NULL DEFAULT 'ID',
    default_document_types  TEXT[]       NOT NULL DEFAULT '{}',
    is_active               BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_countries_code_uppercase
        CHECK (code = UPPER(code) AND length(code) = 2),

    CONSTRAINT ck_countries_name_nonblank
        CHECK (length(trim(name)) > 0),

    CONSTRAINT fk_countries_currency
        FOREIGN KEY (currency_code) REFERENCES currencies(code)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_countries_active
    ON countries (is_active)
    WHERE is_active = TRUE;

-- Seed default countries
INSERT INTO countries (code, name, locale, currency_code, tax_id_label, personal_id_label, default_document_types, is_active) VALUES
    ('PE', 'Perú',       'es-PE', 'PEN', 'RUC',    'DNI',     ARRAY['DNI', 'RUC', 'CE', 'PASAPORTE'], TRUE),
    ('MX', 'México',     'es-MX', 'MXN', 'RFC',    'CURP',    ARRAY['RFC', 'CURP'],                   FALSE),
    ('CO', 'Colombia',   'es-CO', 'COP', 'NIT',    'CC',      ARRAY['NIT', 'CC', 'CE'],               FALSE),
    ('CL', 'Chile',      'es-CL', 'CLP', 'RUT',    'RUN',     ARRAY['RUT', 'RUN'],                    FALSE),
    ('AR', 'Argentina',  'es-AR', 'ARS', 'CUIT',   'DNI',     ARRAY['CUIT', 'DNI'],                   FALSE),
    ('EC', 'Ecuador',    'es-EC', 'USD', 'RUC',    'CI',      ARRAY['RUC', 'CI', 'PASAPORTE'],        FALSE),
    ('BO', 'Bolivia',    'es-BO', 'BOB', 'NIT',    'CI',      ARRAY['NIT', 'CI'],                     FALSE),
    ('US', 'Estados Unidos', 'en-US', 'USD', 'EIN', 'SSN',    ARRAY['EIN', 'SSN', 'ITIN'],            FALSE),
    ('ES', 'España',     'es-ES', 'EUR', 'NIF',    'DNI',     ARRAY['NIF', 'DNI', 'NIE', 'PASAPORTE'], FALSE)
ON CONFLICT (code) DO NOTHING;
