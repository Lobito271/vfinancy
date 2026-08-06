-- 0014_application_settings.up.sql
-- Module 2: Administration — Application Settings
-- Key-value store for company-level configuration.
-- Each setting has a category for grouping and a JSONB value
-- to support simple strings, numbers, booleans, or nested objects.

CREATE TABLE application_settings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID         NOT NULL,
    key          VARCHAR(100) NOT NULL,
    value        JSONB        NOT NULL DEFAULT '{}',
    category     VARCHAR(50)  NOT NULL DEFAULT 'general',
    label        VARCHAR(200),
    description  TEXT,
    is_public    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by   UUID,

    CONSTRAINT fk_settings_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_settings_updated_by
        FOREIGN KEY (updated_by) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL,

    CONSTRAINT ck_settings_key_nonblank
        CHECK (length(trim(key)) > 0),

    CONSTRAINT ck_settings_category_nonblank
        CHECK (length(trim(category)) > 0)
);

CREATE UNIQUE INDEX uq_settings_company_key
    ON application_settings (company_id, key);

CREATE INDEX idx_settings_company_category
    ON application_settings (company_id, category);

CREATE TRIGGER trg_settings_set_updated_at
    BEFORE UPDATE ON application_settings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Seed default application settings for the demo company
INSERT INTO application_settings (company_id, key, value, category, label, description, is_public) VALUES
    ('00000000-0000-0000-0000-000000000001', 'business.name',           '"vfinancy S.A.C."',       'business',   'Razón social',       'Nombre legal de la empresa',           TRUE),
    ('00000000-0000-0000-0000-000000000001', 'business.trade_name',     '"vfinancy"',              'business',   'Nombre comercial',   'Nombre comercial de la empresa',       TRUE),
    ('00000000-0000-0000-0000-000000000001', 'business.tax_id',         '"20600000001"',            'business',   'RUC',                'Número de registro tributario',        TRUE),
    ('00000000-0000-0000-0000-000000000001', 'business.address',        '"Lima, Perú"',             'business',   'Dirección',          'Dirección fiscal de la empresa',       TRUE),
    ('00000000-0000-0000-0000-000000000001', 'business.phone',          '""',                      'business',   'Teléfono',           'Teléfono de contacto',                 TRUE),
    ('00000000-0000-0000-0000-000000000001', 'business.email',          '"admin@vfinancy.local"',   'business',   'Correo electrónico', 'Correo de contacto',                   TRUE),
    ('00000000-0000-0000-0000-000000000001', 'business.logo',           '""',                      'business',   'Logo',               'URL del logo de la empresa',           TRUE),

    ('00000000-0000-0000-0000-000000000001', 'defaults.currency',       '"PEN"',                   'defaults',   'Moneda por defecto', 'Código ISO 4217 de la moneda principal',     TRUE),
    ('00000000-0000-0000-0000-000000000001', 'defaults.tax_code',       '"IGV"',                   'defaults',   'Impuesto por defecto', 'Código del impuesto aplicado por defecto',  TRUE),
    ('00000000-0000-0000-0000-000000000001', 'defaults.expiry_alert_days', '25',                    'defaults',   'Días alerta vencimiento', 'Días antes del vencimiento para alertar', TRUE),
    ('00000000-0000-0000-0000-000000000001', 'defaults.country',        '"PE"',                    'defaults',   'País por defecto',   'Código ISO 3166 del país principal',         TRUE),

    ('00000000-0000-0000-0000-000000000001', 'format.date',             '"DD/MM/YYYY"',            'format',     'Formato de fecha',   'Formato de fecha preferido',                TRUE),
    ('00000000-0000-0000-0000-000000000001', 'format.number',           '"es-PE"',                 'format',     'Formato numérico',   'Locale para formato numérico',              TRUE),
    ('00000000-0000-0000-0000-000000000001', 'format.decimals',         '2',                       'format',     'Decimales',          'Cantidad de decimales por defecto',         TRUE),

    ('00000000-0000-0000-0000-000000000001', 'system.language',         '"es-PE"',                 'system',     'Idioma',             'Idioma de la interfaz',                       TRUE),
    ('00000000-0000-0000-0000-000000000001', 'system.theme',            '"light"',                 'system',     'Tema',               'Tema visual de la aplicación',                TRUE),
    ('00000000-0000-0000-0000-000000000001', 'system.timezone',         '"America/Lima"',          'system',     'Zona horaria',       'Zona horaria del sistema',                    TRUE),
    ('00000000-0000-0000-0000-000000000001', 'system.fiscal_year_start', '1',                      'system',     'Inicio año fiscal',  'Mes de inicio del año fiscal (1-12)',         TRUE),

    ('00000000-0000-0000-0000-000000000001', 'backup.folder',           '""',                      'backup',     'Carpeta de respaldo', 'Ruta de la carpeta de respaldos',       FALSE),
    ('00000000-0000-0000-0000-000000000001', 'backup.export_folder',    '""',                      'backup',     'Carpeta de exportación', 'Ruta de la carpeta de exportaciones', FALSE),
    ('00000000-0000-0000-0000-000000000001', 'backup.auto_frequency',   '"daily"',                 'backup',     'Frecuencia automática', 'Frecuencia de respaldo automático',    FALSE)
ON CONFLICT (company_id, key) DO NOTHING;
