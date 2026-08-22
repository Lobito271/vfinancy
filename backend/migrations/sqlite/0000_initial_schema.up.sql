

SELECT 1;


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

INSERT INTO companies (id, code, legal_name, trade_name, tax_id)
VALUES ('00000000-0000-0000-0000-000000000001', 'SETUP', 'setup placeholder', 'setup', '00000000000');

CREATE UNIQUE INDEX uq_companies_code
    ON companies (code);

CREATE UNIQUE INDEX uq_companies_tax_id
    ON companies (tax_id);

CREATE INDEX idx_companies_active
    ON companies (is_active)
    WHERE deleted_at IS NULL;


CREATE TABLE branches (
    id              TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id      TEXT         NOT NULL,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    address         TEXT,
    phone           VARCHAR(30),
    email           VARCHAR(200),
    is_default      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at      TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at      TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at      TIMESTAMP,
    created_by      TEXT,
    updated_by      TEXT,

    CONSTRAINT fk_branches_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_branches_company_code
    ON branches (company_id, code)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_branches_company_default
    ON branches (company_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_branches_company_active
    ON branches (company_id, is_active)
    WHERE deleted_at IS NULL;

CREATE TABLE local_profiles (
    id TEXT PRIMARY KEY,
    name VARCHAR(200) NOT NULL CHECK (length(trim(name)) > 0),
    password_hash TEXT,
    password_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    failed_attempts INTEGER NOT NULL DEFAULT 0 CHECK (failed_attempts >= 0),
    locked_until TIMESTAMP,
    active_company_id TEXT NOT NULL REFERENCES companies(id),
    theme VARCHAR(20) NOT NULL DEFAULT 'system',
    language VARCHAR(10) NOT NULL DEFAULT 'es-PE',
    date_format VARCHAR(30) NOT NULL DEFAULT 'DD/MM/YYYY',
    number_format VARCHAR(30) NOT NULL DEFAULT 'es-PE',
    decimal_places INTEGER NOT NULL DEFAULT 2 CHECK (decimal_places BETWEEN 0 AND 6),
    timezone VARCHAR(50) NOT NULL DEFAULT 'America/Lima',
    created_at TIMESTAMP NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at TIMESTAMP NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
);

CREATE UNIQUE INDEX uq_local_profiles_singleton ON local_profiles ((1));



CREATE TABLE audit_logs (
    id              TEXT         NOT NULL DEFAULT (lower(hex(randomblob(16)))),
    occurred_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    company_id      TEXT         NOT NULL,
    user_id         TEXT,
    table_name      VARCHAR(100) NOT NULL,
    record_id       TEXT,
    action          VARCHAR(30)  NOT NULL,
    old_value       TEXT,
    new_value       TEXT,
    changed_fields  TEXT,
    ip_address      TEXT,
    user_agent      TEXT,
    device          VARCHAR(200),

    PRIMARY KEY (id, occurred_at),

    CONSTRAINT fk_audit_logs_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT ck_audit_logs_action
        CHECK (action IN (
            'INSERT', 'UPDATE', 'DELETE', 'HARD_DELETE',
            'APPROVE', 'REJECT', 'CANCEL',
            'CONCILIATE', 'CLOSE_PERIOD', 'REOPEN_PERIOD',
            'EXPORT', 'PRINT', 'SEND'
        ))
);

CREATE INDEX idx_audit_logs_company_time
    ON audit_logs (company_id, occurred_at DESC);

CREATE INDEX idx_audit_logs_record
    ON audit_logs (table_name, record_id, occurred_at DESC)
    WHERE record_id IS NOT NULL;

CREATE INDEX idx_audit_logs_user_time
    ON audit_logs (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX idx_audit_logs_action_time
    ON audit_logs (action, occurred_at DESC);

CREATE INDEX idx_audit_logs_company_action_time
    ON audit_logs (company_id, action, occurred_at DESC);

CREATE TRIGGER trg_audit_logs_no_update
    BEFORE UPDATE ON audit_logs
BEGIN
    SELECT RAISE(ABORT, 'audit_logs is append-only; UPDATE is not allowed');
END;

CREATE TRIGGER trg_audit_logs_no_delete
    BEFORE DELETE ON audit_logs
BEGIN
    SELECT RAISE(ABORT, 'audit_logs is append-only; DELETE is not allowed');
END;



CREATE TABLE application_settings (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT         NOT NULL,
    key          VARCHAR(100) NOT NULL,
    value        TEXT         NOT NULL DEFAULT '{}',
    category     VARCHAR(50)  NOT NULL DEFAULT 'general',
    label        VARCHAR(200),
    description  TEXT,
    is_public    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by   TEXT,

    CONSTRAINT fk_settings_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT ck_settings_key_nonblank
        CHECK (length(trim(key)) > 0),

    CONSTRAINT ck_settings_category_nonblank
        CHECK (length(trim(category)) > 0)
);

CREATE UNIQUE INDEX uq_settings_company_key
    ON application_settings (company_id, key);

CREATE INDEX idx_settings_company_category
    ON application_settings (company_id, category);

INSERT OR IGNORE INTO application_settings (company_id, key, value, category, label, description, is_public) VALUES
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
    ('00000000-0000-0000-0000-000000000001', 'backup.auto_frequency',   '"daily"',                 'backup',     'Frecuencia automática', 'Frecuencia de respaldo automático',    FALSE);


CREATE TABLE currencies (
    code           VARCHAR(3) PRIMARY KEY,
    symbol         VARCHAR(10)  NOT NULL,
    name           VARCHAR(100) NOT NULL,
    decimal_places INTEGER      NOT NULL DEFAULT 2
        CHECK (decimal_places BETWEEN 0 AND 6),
    type           VARCHAR(20)  NOT NULL DEFAULT 'fiat'
        CHECK (type IN ('fiat', 'crypto', 'commodity')),
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at     TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT ck_currencies_code_uppercase
        CHECK (code = UPPER(code) AND length(code) = 3),

    CONSTRAINT ck_currencies_name_nonblank
        CHECK (length(trim(name)) > 0)
);

CREATE INDEX idx_currencies_active
    ON currencies (is_active)
    WHERE is_active = TRUE;

INSERT OR IGNORE INTO currencies (code, symbol, name, decimal_places, type, is_active) VALUES
    ('PEN', 'S/',  'Sol peruano',        2, 'fiat', TRUE),
    ('USD', '$',   'Dólar estadounidense', 2, 'fiat', TRUE),
    ('EUR', '€',   'Euro',               2, 'fiat', FALSE),
    ('MXN', 'MX$', 'Peso mexicano',      2, 'fiat', FALSE),
    ('COP', 'COP$', 'Peso colombiano',   2, 'fiat', FALSE),
    ('CLP', 'CLP$', 'Peso chileno',      0, 'fiat', FALSE),
    ('BRL', 'R$',  'Real brasileño',     2, 'fiat', FALSE),
    ('ARS', 'AR$', 'Peso argentino',     2, 'fiat', FALSE),
    ('BOB', 'Bs',  'Boliviano',          2, 'fiat', FALSE);


CREATE TABLE countries (
    code                    VARCHAR(2) PRIMARY KEY,
    name                    VARCHAR(100) NOT NULL,
    locale                  VARCHAR(10)  NOT NULL DEFAULT 'es-PE',
    currency_code           VARCHAR(3)   NOT NULL DEFAULT 'PEN',
    tax_id_label            VARCHAR(50)  NOT NULL DEFAULT 'Tax ID',
    personal_id_label       VARCHAR(50)  NOT NULL DEFAULT 'ID',
    default_document_types  TEXT         NOT NULL DEFAULT '{}',
    is_active               BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

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

INSERT OR IGNORE INTO countries (code, name, locale, currency_code, tax_id_label, personal_id_label, default_document_types, is_active) VALUES
    ('PE', 'Perú',       'es-PE', 'PEN', 'RUC',    'DNI',     '{DNI,RUC,CE,PASAPORTE}', TRUE),
    ('MX', 'México',     'es-MX', 'MXN', 'RFC',    'CURP',    '{RFC,CURP}',             FALSE),
    ('CO', 'Colombia',   'es-CO', 'COP', 'NIT',    'CC',      '{NIT,CC,CE}',            FALSE),
    ('CL', 'Chile',      'es-CL', 'CLP', 'RUT',    'RUN',     '{RUT,RUN}',              FALSE),
    ('AR', 'Argentina',  'es-AR', 'ARS', 'CUIT',   'DNI',     '{CUIT,DNI}',             FALSE),
    ('EC', 'Ecuador',    'es-EC', 'USD', 'RUC',    'CI',      '{RUC,CI,PASAPORTE}',     FALSE),
    ('BO', 'Bolivia',    'es-BO', 'BOB', 'NIT',    'CI',      '{NIT,CI}',               FALSE),
    ('US', 'Estados Unidos', 'en-US', 'USD', 'EIN', 'SSN',    '{EIN,SSN,ITIN}',         FALSE),
    ('ES', 'España',     'es-ES', 'EUR', 'NIF',    'DNI',     '{NIF,DNI,NIE,PASAPORTE}', FALSE);


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

INSERT OR IGNORE INTO taxes (company_id, code, name, short_name, country_code, default_rate, is_inclusive, is_percentage, category, is_active) VALUES
    (NULL, 'IGV',      'Impuesto General a las Ventas', 'IGV',     'PE', '0.1800', FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVAP',     'Impuesto de Promoción Municipal','IVAP',    'PE', '0.0200', FALSE, TRUE, 'municipal',TRUE),
    (NULL, 'RENTA',    'Impuesto a la Renta',            'Renta',   'PE', '0.2950', FALSE, TRUE, 'income',   TRUE),
    (NULL, 'EXONERADO','Exonerado del IGV',              'Exo',     'PE', '0.0000', FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'GRATUITO', 'Operación Gratuita',             'Grat',    'PE', '0.0000', FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVA_MX',   'Impuesto al Valor Agregado',    'IVA',     'MX', '0.1600', FALSE, TRUE, 'sales',    FALSE),
    (NULL, 'IVA_CO',   'Impuesto sobre las Ventas',     'IVA',     'CO', '0.1900', FALSE, TRUE, 'sales',    FALSE);



CREATE TABLE exchange_rates (
    id              TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id      TEXT,
    from_currency   VARCHAR(3)     NOT NULL,
    to_currency     VARCHAR(3)     NOT NULL,
    rate_date       DATE           NOT NULL,
    rate            TEXT           NOT NULL
        CHECK (CAST(rate AS REAL) > 0),
    source          VARCHAR(50)    NOT NULL DEFAULT 'manual'
        CHECK (source IN ('manual', 'central_bank', 'sunat', 'bloomberg', 'other')),
    created_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_exchange_rates_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_exchange_rates_from
        FOREIGN KEY (from_currency) REFERENCES currencies(code)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT fk_exchange_rates_to
        FOREIGN KEY (to_currency) REFERENCES currencies(code)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT ck_exchange_rates_different_currencies
        CHECK (from_currency <> to_currency)
);

CREATE UNIQUE INDEX uq_exchange_rates_pair_date
    ON exchange_rates (from_currency, to_currency, rate_date);

CREATE INDEX idx_exchange_rates_lookup
    ON exchange_rates (from_currency, to_currency, rate_date DESC);


CREATE TABLE audit_events (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT         NOT NULL,
    user_id      TEXT,
    session_id   TEXT,
    event_type   VARCHAR(50)  NOT NULL
        CHECK (event_type IN (
            'CONFIG_UPDATE',
            'BACKUP_CREATE', 'EXPORT_DATA'
        )),
    target_type  VARCHAR(100),
    target_id    TEXT,
    description  TEXT,
    metadata     TEXT         NOT NULL DEFAULT '{}',
    ip_address   TEXT,
    device       VARCHAR(100),
    occurred_at  TIMESTAMP    NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_audit_events_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT ck_audit_events_target_type_nonblank
        CHECK (target_type IS NULL OR length(trim(target_type)) > 0)
);

CREATE INDEX idx_audit_events_company_time
    ON audit_events (company_id, occurred_at DESC);

CREATE INDEX idx_audit_events_user
    ON audit_events (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX idx_audit_events_type
    ON audit_events (event_type, occurred_at DESC);

CREATE INDEX idx_audit_events_target
    ON audit_events (target_type, target_id)
    WHERE target_type IS NOT NULL AND target_id IS NOT NULL;

CREATE TRIGGER trg_audit_events_no_update
    BEFORE UPDATE ON audit_events
BEGIN
    SELECT RAISE(ABORT, 'audit_events is append-only: UPDATE is not allowed');
END;

CREATE TRIGGER trg_audit_events_no_delete
    BEFORE DELETE ON audit_events
BEGIN
    SELECT RAISE(ABORT, 'audit_events is append-only: DELETE is not allowed');
END;


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


CREATE TRIGGER trg_companies_sync_delete AFTER DELETE ON companies BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('companies', OLD.id, OLD.updated_at);
END;

CREATE TRIGGER trg_branches_sync_delete AFTER DELETE ON branches BEGIN
    INSERT INTO sync_tombstones (table_name, record_id, updated_at)
    VALUES ('branches', OLD.id, OLD.updated_at);
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


CREATE TABLE customers (
    id                 TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id         TEXT           NOT NULL,
    default_branch_id  TEXT,
    document_type      VARCHAR(10)    NOT NULL
        CHECK (document_type IN ('DNI', 'RUC', 'CE', 'PASSPORT')),
    document_number    VARCHAR(30)    NOT NULL,
    business_name      VARCHAR(200)   NOT NULL,
    trade_name         VARCHAR(200),
    tax_category       VARCHAR(20)    NOT NULL DEFAULT 'taxed'
        CHECK (tax_category IN ('taxed', 'exempt', 'unaffected', 'export')),
    credit_limit       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(credit_limit AS REAL) >= 0),
    current_debt       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(current_debt AS REAL) >= 0),
    payment_term_days  INTEGER        NOT NULL DEFAULT 0
        CHECK (payment_term_days >= 0),
    status             VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'blocked')),
    blocked_reason     TEXT,
    email              VARCHAR(200),
    phone              VARCHAR(30),
    address            TEXT,
    is_active          BOOLEAN        NOT NULL DEFAULT TRUE,

    created_at         TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at         TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at         TIMESTAMP,
    created_by         TEXT,
    updated_by         TEXT,

    CONSTRAINT fk_customers_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customers_default_branch
        FOREIGN KEY (default_branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT ck_customers_business_name_nonblank
        CHECK (length(trim(business_name)) > 0)
);

CREATE UNIQUE INDEX uq_customers_document
    ON customers (company_id, document_type, document_number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customers_status
    ON customers (company_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customers_branch
    ON customers (default_branch_id)
    WHERE deleted_at IS NULL;


CREATE TABLE suppliers (
    id                 TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id         TEXT           NOT NULL,
    document_type      VARCHAR(10)    NOT NULL
        CHECK (document_type IN ('DNI', 'RUC', 'CE', 'PASSPORT')),
    document_number    VARCHAR(30)    NOT NULL,
    business_name      VARCHAR(200)   NOT NULL,
    trade_name         VARCHAR(200),
    tax_id             VARCHAR(30),
    is_international   BOOLEAN        NOT NULL DEFAULT FALSE,
    default_currency   VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    current_debt       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(current_debt AS REAL) >= 0),
    payment_term_days  INTEGER        NOT NULL DEFAULT 0
        CHECK (payment_term_days >= 0),
    status             VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),
    email              VARCHAR(200),
    phone              VARCHAR(30),
    address            TEXT,
    is_active          BOOLEAN        NOT NULL DEFAULT TRUE,

    created_at         TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at         TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at         TIMESTAMP,
    created_by         TEXT,
    updated_by         TEXT,

    CONSTRAINT fk_suppliers_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_suppliers_default_currency
        FOREIGN KEY (default_currency) REFERENCES currencies(code)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_suppliers_business_name_nonblank
        CHECK (length(trim(business_name)) > 0)
);

CREATE UNIQUE INDEX uq_suppliers_document
    ON suppliers (company_id, document_type, document_number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_suppliers_status
    ON suppliers (company_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_suppliers_international
    ON suppliers (company_id, is_international)
    WHERE deleted_at IS NULL;


CREATE TABLE product_categories (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    parent_id    TEXT,
    path         VARCHAR(100),
    depth        INTEGER        NOT NULL DEFAULT 0
        CHECK (depth >= 0),
    created_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at   TIMESTAMP,
    created_by   TEXT,
    updated_by   TEXT,

    CONSTRAINT fk_product_categories_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_product_categories_parent
        FOREIGN KEY (parent_id) REFERENCES product_categories(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_product_categories_code
    ON product_categories (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE product_brands (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    created_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at   TIMESTAMP,
    created_by   TEXT,
    updated_by   TEXT,

    CONSTRAINT fk_product_brands_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_product_brands_code
    ON product_brands (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE units_of_measure (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    code             VARCHAR(20)    NOT NULL,
    name             VARCHAR(100)   NOT NULL,
    symbol           VARCHAR(20),
    allows_decimals  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_units_of_measure_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_units_of_measure_code
    ON units_of_measure (company_id, code);

CREATE TABLE warehouses (
    id                TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id        TEXT           NOT NULL,
    branch_id         TEXT           NOT NULL,
    code              VARCHAR(20)    NOT NULL,
    name              VARCHAR(200)   NOT NULL,
    address           TEXT,
    manager_id        TEXT,
    is_default        BOOLEAN        NOT NULL DEFAULT FALSE,
    allows_clearance  BOOLEAN        NOT NULL DEFAULT FALSE,
    is_active         BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at        TIMESTAMP,
    created_by        TEXT,
    updated_by        TEXT,

    CONSTRAINT fk_warehouses_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_warehouses_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_warehouses_code
    ON warehouses (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE products (
    id             TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id     TEXT           NOT NULL,
    sku            VARCHAR(50)    NOT NULL,
    barcode        VARCHAR(50),
    description    TEXT           NOT NULL,
    category_id    TEXT,
    brand_id       TEXT,
    unit_id        TEXT           NOT NULL,
    tax_id         TEXT           NOT NULL,
    cost           TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(cost AS REAL) >= 0),
    sale_price     TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(sale_price AS REAL) >= 0),
    sale_currency  VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    min_stock      TEXT           NOT NULL DEFAULT '0.0000',
    max_stock      TEXT           NOT NULL DEFAULT '0.0000',
    weight         TEXT           NOT NULL DEFAULT '0.0000',
    is_active      BOOLEAN        NOT NULL DEFAULT TRUE,
    is_service     BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at     TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at     TIMESTAMP,
    created_by     TEXT,
    updated_by     TEXT,

    CONSTRAINT fk_products_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id) REFERENCES product_categories(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_products_brand
        FOREIGN KEY (brand_id) REFERENCES product_brands(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_products_unit
        FOREIGN KEY (unit_id) REFERENCES units_of_measure(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_products_tax
        FOREIGN KEY (tax_id) REFERENCES taxes(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_products_description_nonblank
        CHECK (length(trim(description)) > 0)
);

CREATE UNIQUE INDEX uq_products_sku
    ON products (company_id, sku)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_products_barcode
    ON products (company_id, barcode)
    WHERE deleted_at IS NULL AND barcode IS NOT NULL;

CREATE INDEX idx_products_category
    ON products (company_id, category_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_products_brand
    ON products (company_id, brand_id)
    WHERE deleted_at IS NULL;


CREATE TABLE fiscal_periods (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT           NOT NULL,
    name         VARCHAR(100)   NOT NULL,
    period_start TIMESTAMP      NOT NULL,
    period_end   TIMESTAMP      NOT NULL,
    status       VARCHAR(20)    NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closing', 'closed')),
    closed_at    TIMESTAMP,
    closed_by    TEXT,
    created_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by   TEXT,
    updated_by   TEXT,

    CONSTRAINT fk_fiscal_periods_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT ck_fiscal_periods_range
        CHECK (period_end >= period_start)
);

CREATE UNIQUE INDEX uq_fiscal_periods_name
    ON fiscal_periods (company_id, name);

CREATE UNIQUE INDEX uq_fiscal_periods_range
    ON fiscal_periods (company_id, period_start, period_end);


CREATE TABLE chart_of_accounts (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    code             VARCHAR(20)    NOT NULL,
    name             VARCHAR(200)   NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('asset', 'liability', 'equity', 'income', 'expense')),
    parent_id        TEXT,
    path             VARCHAR(100),
    depth            INTEGER        NOT NULL DEFAULT 1
        CHECK (depth >= 1),
    is_active        BOOLEAN        NOT NULL DEFAULT TRUE,
    allows_movement  BOOLEAN        NOT NULL DEFAULT TRUE,
    description      TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_chart_of_accounts_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_chart_of_accounts_parent
        FOREIGN KEY (parent_id) REFERENCES chart_of_accounts(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT ck_chart_of_accounts_name_nonblank
        CHECK (length(trim(name)) > 0)
);

CREATE UNIQUE INDEX uq_chart_of_accounts_code
    ON chart_of_accounts (company_id, code);

CREATE INDEX idx_chart_of_accounts_type
    ON chart_of_accounts (company_id, type)
    WHERE is_active = TRUE;

CREATE INDEX idx_chart_of_accounts_parent
    ON chart_of_accounts (parent_id);


CREATE TABLE journal_entries (
    id                TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id        TEXT           NOT NULL,
    fiscal_period_id  TEXT           NOT NULL,
    number            VARCHAR(30)    NOT NULL,
    entry_date        TIMESTAMP      NOT NULL,
    posting_date      TIMESTAMP,
    description       TEXT,
    source            VARCHAR(20)    NOT NULL
        CHECK (source IN ('sale', 'purchase', 'payment', 'manual', 'adjustment', 'closing', 'opening')),
    source_id         TEXT,
    status            VARCHAR(20)    NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'posted', 'reversed')),
    reverses_entry_id TEXT,
    reversed_by_entry_id TEXT,
    posted_at         TIMESTAMP,
    created_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by        TEXT,
    posted_by         TEXT,

    CONSTRAINT fk_journal_entries_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_journal_entries_period
        FOREIGN KEY (fiscal_period_id) REFERENCES fiscal_periods(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_journal_entries_reverses
        FOREIGN KEY (reverses_entry_id) REFERENCES journal_entries(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_journal_entries_number
    ON journal_entries (company_id, number);

CREATE INDEX idx_journal_entries_period
    ON journal_entries (company_id, fiscal_period_id, entry_date);

CREATE INDEX idx_journal_entries_source
    ON journal_entries (source, source_id);

CREATE TABLE journal_entry_lines (
    id                    TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    journal_entry_id      TEXT           NOT NULL,
    line_number           INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    account_id            TEXT           NOT NULL,
    description           TEXT,
    debit                 TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(debit AS REAL) >= 0),
    credit                TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(credit AS REAL) >= 0),
    currency_code         VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate         TEXT           NOT NULL DEFAULT '1.000000',
    amount_in_txn_currency TEXT          NOT NULL DEFAULT '0.00',
    created_at            TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_journal_entry_lines_entry
        FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_journal_entry_lines_account
        FOREIGN KEY (account_id) REFERENCES chart_of_accounts(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_journal_entry_lines_one_side
        CHECK ((CAST(debit AS REAL) > 0) <> (CAST(credit AS REAL) > 0))
);

CREATE INDEX idx_journal_entry_lines_entry
    ON journal_entry_lines (journal_entry_id, line_number);

CREATE INDEX idx_journal_entry_lines_account
    ON journal_entry_lines (account_id);


CREATE TABLE bank_accounts (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    branch_id        TEXT,
    bank_name        VARCHAR(100)   NOT NULL,
    account_number   VARCHAR(50)    NOT NULL,
    account_type     VARCHAR(20)    NOT NULL DEFAULT 'checking'
        CHECK (account_type IN ('checking', 'savings')),
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    gl_account_id    TEXT           NOT NULL,
    current_balance  TEXT           NOT NULL DEFAULT '0.00',
    is_default       BOOLEAN        NOT NULL DEFAULT FALSE,
    is_active        BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at       TIMESTAMP,
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_bank_accounts_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_bank_accounts_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_bank_accounts_gl
        FOREIGN KEY (gl_account_id) REFERENCES chart_of_accounts(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_bank_accounts_number
    ON bank_accounts (company_id, account_number)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_bank_accounts_default
    ON bank_accounts (company_id, is_default)
    WHERE is_default = TRUE AND deleted_at IS NULL;

CREATE TABLE credit_cards (
    id                TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id        TEXT           NOT NULL,
    branch_id         TEXT,
    issuer            VARCHAR(100)   NOT NULL,
    last_four         VARCHAR(4)     NOT NULL,
    card_holder       VARCHAR(200)   NOT NULL,
    expiration_month  INTEGER        NOT NULL
        CHECK (expiration_month BETWEEN 1 AND 12),
    expiration_year   INTEGER        NOT NULL
        CHECK (expiration_year BETWEEN 2000 AND 2100),
    credit_limit      TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(credit_limit AS REAL) >= 0),
    current_balance   TEXT           NOT NULL DEFAULT '0.00',
    cut_off_day       INTEGER        NOT NULL DEFAULT 1
        CHECK (cut_off_day BETWEEN 1 AND 31),
    payment_due_day   INTEGER        NOT NULL DEFAULT 1
        CHECK (payment_due_day BETWEEN 1 AND 31),
    currency_code     VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    gl_account_id     TEXT           NOT NULL,
    is_active         BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at        TIMESTAMP,
    created_by        TEXT,
    updated_by        TEXT,

    CONSTRAINT fk_credit_cards_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_credit_cards_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_credit_cards_gl
        FOREIGN KEY (gl_account_id) REFERENCES chart_of_accounts(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_credit_cards_company
    ON credit_cards (company_id)
    WHERE deleted_at IS NULL;

CREATE TABLE bank_transactions (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    bank_account_id  TEXT           NOT NULL,
    transaction_date TIMESTAMP      NOT NULL,
    value_date       TIMESTAMP,
    description      TEXT,
    amount           TEXT           NOT NULL,
    type             VARCHAR(20)    NOT NULL DEFAULT 'other'
        CHECK (type IN ('deposit', 'withdrawal', 'fee', 'interest', 'transfer', 'other')),
    reference        VARCHAR(100),
    balance_after    TEXT,
    is_reconciled    BOOLEAN        NOT NULL DEFAULT FALSE,
    reconciled_at    TIMESTAMP,
    reconciled_by    TEXT,
    journal_entry_id TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_bank_transactions_account
        FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_bank_transactions_journal
        FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_bank_transactions_account_date
    ON bank_transactions (bank_account_id, transaction_date);

CREATE INDEX idx_bank_transactions_unreconciled
    ON bank_transactions (bank_account_id, is_reconciled)
    WHERE is_reconciled = FALSE;


CREATE TABLE international_returns (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    supplier_id      TEXT           NOT NULL,
    number           VARCHAR(30)    NOT NULL,
    return_date      TIMESTAMP      NOT NULL,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'USD',
    exchange_rate    TEXT           NOT NULL DEFAULT '1.000000',
    subtotal         TEXT           NOT NULL DEFAULT '0.00',
    total            TEXT           NOT NULL DEFAULT '0.00',
    reason           TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'authorized', 'received', 'reconciled', 'cancelled')),
    authorized_at    TIMESTAMP,
    authorized_by    TEXT,
    received_at      TIMESTAMP,
    reconciled_at    TIMESTAMP,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_international_returns_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_international_returns_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_international_returns_total
        CHECK (CAST(total AS REAL) >= 0)
);

CREATE UNIQUE INDEX uq_international_returns_number
    ON international_returns (company_id, number);

CREATE TABLE international_return_items (
    id                      TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    international_return_id TEXT           NOT NULL,
    product_id              TEXT           NOT NULL,
    quantity                TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity AS REAL) > 0),
    unit_cost               TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(unit_cost AS REAL) >= 0),
    line_total               TEXT          NOT NULL DEFAULT '0.00',
    created_at              TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_intl_return_items_return
        FOREIGN KEY (international_return_id) REFERENCES international_returns(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_intl_return_items_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_intl_return_items_return
    ON international_return_items (international_return_id);

CREATE INDEX idx_intl_return_items_product
    ON international_return_items (product_id);


CREATE TABLE inventory_batches (
    id              TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id      TEXT           NOT NULL,
    product_id      TEXT           NOT NULL,
    warehouse_id    TEXT           NOT NULL,
    supplier_id     TEXT,
    purchase_order_item_id TEXT,
    lot             VARCHAR(100),
    batch_code      VARCHAR(50),
    arrival_date    TIMESTAMP      NOT NULL,
    expiry_date     TIMESTAMP,
    quantity        TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity AS REAL) >= 0),
    original_quantity TEXT         NOT NULL DEFAULT '0.0000'
        CHECK (CAST(original_quantity AS REAL) >= 0),
    unit_cost       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(unit_cost AS REAL) >= 0),
    currency_code   VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate   TEXT           NOT NULL DEFAULT '1.000000',
    status          VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'depleted', 'written_off')),
    clearance_date  TIMESTAMP,
    is_clearance    BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by      TEXT,
    updated_by      TEXT,

    CONSTRAINT fk_inventory_batches_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_inventory_batches_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_batches_warehouse
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_batches_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_inventory_batches_poi
        FOREIGN KEY (purchase_order_item_id) REFERENCES purchase_order_items(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_inventory_batches_product
    ON inventory_batches (product_id, warehouse_id, status);

CREATE INDEX idx_inventory_batches_clearance
    ON inventory_batches (company_id, is_clearance, status)
    WHERE is_clearance = TRUE AND status = 'active';

CREATE TABLE inventory_movements (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    batch_id         TEXT           NOT NULL,
    product_id       TEXT           NOT NULL,
    warehouse_id     TEXT           NOT NULL,
    movement_date    TIMESTAMP      NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out')),
    reference_type   VARCHAR(30),
    reference_id     TEXT,
    quantity_delta   TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity_delta AS REAL) <> 0),
    balance_after    TEXT           NOT NULL DEFAULT '0.0000',
    unit_cost        TEXT           NOT NULL DEFAULT '0.00',
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    notes            TEXT,
    created_by       TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_inventory_movements_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_inventory_movements_batch
        FOREIGN KEY (batch_id) REFERENCES inventory_batches(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_warehouse
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_inventory_movements_batch
    ON inventory_movements (batch_id, movement_date);

CREATE INDEX idx_inventory_movements_product
    ON inventory_movements (product_id, movement_date);

CREATE INDEX idx_inventory_movements_reference
    ON inventory_movements (reference_type, reference_id);


CREATE TABLE sales (
    id             TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id     TEXT           NOT NULL,
    branch_id      TEXT           NOT NULL,
    customer_id    TEXT           NOT NULL,
    number         VARCHAR(30)    NOT NULL,
    sale_date      TIMESTAMP      NOT NULL,
    due_date       TIMESTAMP,
    status         VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'partial', 'paid', 'cancelled')),
    subtotal       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(subtotal AS REAL) >= 0),
    tax_amount     TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(tax_amount AS REAL) >= 0),
    total          TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(total AS REAL) >= 0),
    paid_amount    TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(paid_amount AS REAL) >= 0),
    discount_amount TEXT          NOT NULL DEFAULT '0.00'
        CHECK (CAST(discount_amount AS REAL) >= 0),
    cost_total     TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(cost_total AS REAL) >= 0),
    profit         TEXT           NOT NULL DEFAULT '0.00',
    currency_code  VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate  TEXT           NOT NULL DEFAULT '1.000000',
    notes          TEXT,
    cancelled_at   TIMESTAMP,
    cancelled_reason TEXT,
    created_at     TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at     TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at     TIMESTAMP,
    created_by     TEXT,
    updated_by     TEXT,

    CONSTRAINT fk_sales_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_sales_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_sales_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_sales_dates
        CHECK (due_date IS NULL OR due_date >= sale_date)
);

CREATE UNIQUE INDEX uq_sales_number
    ON sales (company_id, number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_sales_customer
    ON sales (customer_id, sale_date);

CREATE INDEX idx_sales_status
    ON sales (company_id, status, sale_date);

CREATE TABLE sale_items (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    sale_id          TEXT           NOT NULL,
    product_id       TEXT           NOT NULL,
    inventory_batch_id TEXT,
    line_number      INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    quantity         TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity AS REAL) > 0),
    unit_price       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(unit_price AS REAL) >= 0),
    discount_percent TEXT           NOT NULL DEFAULT '0.0000',
    discount_amount  TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(discount_amount AS REAL) >= 0),
    tax_rate         TEXT           NOT NULL DEFAULT '0.00',
    tax_amount       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(tax_amount AS REAL) >= 0),
    line_total       TEXT           NOT NULL DEFAULT '0.00',
    cost_snapshot    TEXT           NOT NULL DEFAULT '0.00',
    description      TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_sale_items_sale
        FOREIGN KEY (sale_id) REFERENCES sales(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_sale_items_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_sale_items_batch
        FOREIGN KEY (inventory_batch_id) REFERENCES inventory_batches(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_sale_items_sale
    ON sale_items (sale_id);

CREATE INDEX idx_sale_items_product
    ON sale_items (product_id);


CREATE TABLE customer_payments (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    customer_id      TEXT           NOT NULL,
    branch_id        TEXT,
    number           VARCHAR(30)    NOT NULL,
    payment_date     TIMESTAMP      NOT NULL,
    amount           TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(amount AS REAL) > 0),
    payment_method   VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    reference        VARCHAR(100),
    bank_account_id  TEXT,
    cash_register_id TEXT,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    TEXT           NOT NULL DEFAULT '1.000000',
    notes            TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at       TIMESTAMP,
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_customer_payments_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customer_payments_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_customer_payments_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_customer_payments_bank
        FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_customer_payments_number
    ON customer_payments (company_id, number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customer_payments_customer
    ON customer_payments (customer_id, payment_date);

CREATE TABLE customer_payment_allocations (
    id                 TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    customer_payment_id TEXT          NOT NULL,
    sale_id            TEXT           NOT NULL,
    allocated_amount   TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(allocated_amount AS REAL) > 0),
    created_at         TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_customer_payment_alloc_payment
        FOREIGN KEY (customer_payment_id) REFERENCES customer_payments(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customer_payment_alloc_sale
        FOREIGN KEY (sale_id) REFERENCES sales(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_customer_payment_alloc_payment
    ON customer_payment_allocations (customer_payment_id);

CREATE INDEX idx_customer_payment_alloc_sale
    ON customer_payment_allocations (sale_id);

CREATE TABLE customer_advances (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    customer_id      TEXT           NOT NULL,
    number           VARCHAR(30)    NOT NULL,
    advance_date     TIMESTAMP      NOT NULL,
    amount           TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(amount AS REAL) >= 0),
    remaining        TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(remaining AS REAL) >= 0),
    payment_method   VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    reference        VARCHAR(100),
    bank_account_id  TEXT,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    TEXT           NOT NULL DEFAULT '1.000000',
    notes            TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at       TIMESTAMP,
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_customer_advances_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customer_advances_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_customer_advances_bank
        FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_customer_advances_number
    ON customer_advances (company_id, number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customer_advances_customer
    ON customer_advances (customer_id);

CREATE TABLE customer_advance_applications (
    id                   TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    customer_advance_id  TEXT           NOT NULL,
    sale_id              TEXT           NOT NULL,
    applied_amount       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(applied_amount AS REAL) > 0),
    created_at           TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_customer_adv_app_advance
        FOREIGN KEY (customer_advance_id) REFERENCES customer_advances(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customer_adv_app_sale
        FOREIGN KEY (sale_id) REFERENCES sales(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_customer_adv_app_advance
    ON customer_advance_applications (customer_advance_id);

CREATE INDEX idx_customer_adv_app_sale
    ON customer_advance_applications (sale_id);


CREATE TABLE purchase_orders (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    supplier_id      TEXT           NOT NULL,
    branch_id        TEXT,
    number           VARCHAR(30)    NOT NULL,
    order_date       TIMESTAMP      NOT NULL,
    expected_date    TIMESTAMP,
    received_date    TIMESTAMP,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'received', 'paid', 'reconciled', 'cancelled')),
    subtotal         TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(subtotal AS REAL) >= 0),
    tax_amount       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(tax_amount AS REAL) >= 0),
    total            TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(total AS REAL) >= 0),
    paid_amount      TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(paid_amount AS REAL) >= 0),
    discount_amount  TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(discount_amount AS REAL) >= 0),
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    TEXT           NOT NULL DEFAULT '1.000000',
    notes            TEXT,
    cancelled_at     TIMESTAMP,
    cancelled_reason TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at       TIMESTAMP,
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_purchase_orders_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_purchase_orders_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_purchase_orders_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT ck_purchase_orders_dates
        CHECK (expected_date IS NULL OR expected_date >= order_date)
);

CREATE UNIQUE INDEX uq_purchase_orders_number
    ON purchase_orders (company_id, number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_purchase_orders_supplier
    ON purchase_orders (supplier_id, order_date);

CREATE INDEX idx_purchase_orders_status
    ON purchase_orders (company_id, status, order_date);

CREATE TABLE purchase_order_items (
    id                 TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    purchase_order_id  TEXT           NOT NULL,
    product_id         TEXT           NOT NULL,
    line_number        INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    quantity_ordered   TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity_ordered AS REAL) > 0),
    quantity_received  TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity_received AS REAL) >= 0),
    unit_cost          TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(unit_cost AS REAL) >= 0),
    discount_percent   TEXT           NOT NULL DEFAULT '0.0000',
    discount_amount    TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(discount_amount AS REAL) >= 0),
    tax_rate           TEXT           NOT NULL DEFAULT '0.00',
    tax_amount         TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(tax_amount AS REAL) >= 0),
    line_total         TEXT           NOT NULL DEFAULT '0.00',
    description        TEXT,
    created_at         TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_purchase_order_items_order
        FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_purchase_order_items_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_purchase_order_items_received
        CHECK (CAST(quantity_received AS REAL) <= CAST(quantity_ordered AS REAL))
);

CREATE INDEX idx_purchase_order_items_order
    ON purchase_order_items (purchase_order_id);

CREATE INDEX idx_purchase_order_items_product
    ON purchase_order_items (product_id);

CREATE TABLE supplier_payments (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    supplier_id      TEXT           NOT NULL,
    branch_id        TEXT,
    number           VARCHAR(30)    NOT NULL,
    payment_date     TIMESTAMP      NOT NULL,
    amount           TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(amount AS REAL) > 0),
    payment_method   VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    reference        VARCHAR(100),
    bank_account_id  TEXT,
    cash_register_id TEXT,
    credit_card_id   TEXT,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    TEXT           NOT NULL DEFAULT '1.000000',
    notes            TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at       TIMESTAMP,
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_supplier_payments_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_supplier_payments_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_supplier_payments_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_supplier_payments_bank
        FOREIGN KEY (bank_account_id) REFERENCES bank_accounts(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_supplier_payments_card
        FOREIGN KEY (credit_card_id) REFERENCES credit_cards(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_supplier_payments_number
    ON supplier_payments (company_id, number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_supplier_payments_supplier
    ON supplier_payments (supplier_id, payment_date);

CREATE TABLE supplier_payment_allocations (
    id                  TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    supplier_payment_id TEXT           NOT NULL,
    purchase_order_id   TEXT           NOT NULL,
    allocated_amount    TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(allocated_amount AS REAL) > 0),
    created_at          TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_supplier_payment_alloc_payment
        FOREIGN KEY (supplier_payment_id) REFERENCES supplier_payments(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_supplier_payment_alloc_po
        FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_supplier_payment_alloc_payment
    ON supplier_payment_allocations (supplier_payment_id);

CREATE INDEX idx_supplier_payment_alloc_po
    ON supplier_payment_allocations (purchase_order_id);




PRAGMA defer_foreign_keys = ON;

UPDATE taxes
SET id = lower(substr(id, 1, 8) || '-' || substr(id, 9, 4) || '-' || substr(id, 13, 4) || '-' || substr(id, 17, 4) || '-' || substr(id, 21, 12))
WHERE id NOT LIKE '%-%';

UPDATE products
SET tax_id = lower(substr(tax_id, 1, 8) || '-' || substr(tax_id, 9, 4) || '-' || substr(tax_id, 13, 4) || '-' || substr(tax_id, 17, 4) || '-' || substr(tax_id, 21, 12))
WHERE tax_id IS NOT NULL AND tax_id NOT LIKE '%-%';

UPDATE application_settings
SET id = lower(substr(id, 1, 8) || '-' || substr(id, 9, 4) || '-' || substr(id, 13, 4) || '-' || substr(id, 17, 4) || '-' || substr(id, 21, 12))
WHERE id NOT LIKE '%-%';


COMMIT;
PRAGMA foreign_keys = OFF;
BEGIN;

CREATE TABLE inventory_batches_new (
    id              TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id      TEXT           NOT NULL,
    product_id      TEXT           NOT NULL,
    warehouse_id    TEXT           NOT NULL,
    supplier_id     TEXT,
    purchase_order_item_id TEXT,
    lot             VARCHAR(100),
    batch_code      VARCHAR(50),
    arrival_date    TIMESTAMP      NOT NULL,
    expiry_date     TIMESTAMP,
    quantity        TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity AS REAL) >= 0),
    original_quantity TEXT         NOT NULL DEFAULT '0.0000'
        CHECK (CAST(original_quantity AS REAL) >= 0),
    unit_cost       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(unit_cost AS REAL) >= 0),
    currency_code   VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate   TEXT           NOT NULL DEFAULT '1.000000',
    status          VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'depleted', 'written_off', 'voided')),
    clearance_date  TIMESTAMP,
    is_clearance    BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by      TEXT,
    updated_by      TEXT,

    CONSTRAINT fk_inventory_batches_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_inventory_batches_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_batches_warehouse
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_batches_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_inventory_batches_poi
        FOREIGN KEY (purchase_order_item_id) REFERENCES purchase_order_items(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

INSERT INTO inventory_batches_new SELECT * FROM inventory_batches;
DROP TABLE inventory_batches;
ALTER TABLE inventory_batches_new RENAME TO inventory_batches;

CREATE INDEX idx_inventory_batches_product
    ON inventory_batches (product_id, warehouse_id, status);

CREATE INDEX idx_inventory_batches_clearance
    ON inventory_batches (company_id, is_clearance, status)
    WHERE is_clearance = TRUE AND status = 'active';

CREATE TABLE inventory_movements_new (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    batch_id         TEXT           NOT NULL,
    product_id       TEXT           NOT NULL,
    warehouse_id     TEXT           NOT NULL,
    movement_date    TIMESTAMP      NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out', 'void_out')),
    reference_type   VARCHAR(30),
    reference_id     TEXT,
    quantity_delta   TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity_delta AS REAL) <> 0),
    balance_after    TEXT           NOT NULL DEFAULT '0.0000',
    unit_cost        TEXT           NOT NULL DEFAULT '0.00',
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    notes            TEXT,
    created_by       TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_inventory_movements_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_inventory_movements_batch
        FOREIGN KEY (batch_id) REFERENCES inventory_batches(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_warehouse
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

INSERT INTO inventory_movements_new SELECT * FROM inventory_movements;
DROP TABLE inventory_movements;
ALTER TABLE inventory_movements_new RENAME TO inventory_movements;

CREATE INDEX idx_inventory_movements_batch
    ON inventory_movements (batch_id, movement_date);

CREATE INDEX idx_inventory_movements_product
    ON inventory_movements (product_id, movement_date);

CREATE INDEX idx_inventory_movements_reference
    ON inventory_movements (reference_type, reference_id);

COMMIT;
DELETE FROM companies WHERE id = '00000000-0000-0000-0000-000000000001';
PRAGMA foreign_keys = ON;
BEGIN;


COMMIT;
PRAGMA foreign_keys = OFF;
BEGIN;

CREATE TABLE inventory_movements_new (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    batch_id         TEXT           NOT NULL,
    product_id       TEXT           NOT NULL,
    warehouse_id     TEXT           NOT NULL,
    movement_date    TIMESTAMP      NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('purchase', 'purchase_receipt', 'sale', 'void_sale', 'void_purchase', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out', 'void_out')),
    reference_type   VARCHAR(30),
    reference_id     TEXT,
    quantity_delta   TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity_delta AS REAL) <> 0),
    balance_after    TEXT           NOT NULL DEFAULT '0.0000',
    unit_cost        TEXT           NOT NULL DEFAULT '0.00',
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    notes            TEXT,
    created_by       TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_inventory_movements_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_inventory_movements_batch
        FOREIGN KEY (batch_id) REFERENCES inventory_batches(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_warehouse
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

INSERT INTO inventory_movements_new SELECT * FROM inventory_movements;
DROP TABLE inventory_movements;
ALTER TABLE inventory_movements_new RENAME TO inventory_movements;

CREATE INDEX idx_inventory_movements_batch
    ON inventory_movements (batch_id, movement_date);

CREATE INDEX idx_inventory_movements_product
    ON inventory_movements (product_id, movement_date);

CREATE INDEX idx_inventory_movements_reference
    ON inventory_movements (reference_type, reference_id);

COMMIT;
PRAGMA foreign_keys = ON;
BEGIN;
