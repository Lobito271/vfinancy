

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


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


CREATE TABLE branches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID         NOT NULL,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    address         TEXT,
    phone           VARCHAR(30),
    email           VARCHAR(200),
    is_default      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    created_by      UUID,
    updated_by      UUID,

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

CREATE TRIGGER trg_branches_set_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE local_profiles (
    id UUID PRIMARY KEY,
    name VARCHAR(200) NOT NULL CHECK (length(trim(name)) > 0),
    password_hash TEXT,
    password_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    failed_attempts INTEGER NOT NULL DEFAULT 0 CHECK (failed_attempts >= 0),
    locked_until TIMESTAMPTZ,
    active_company_id UUID NOT NULL REFERENCES companies(id),
    theme VARCHAR(20) NOT NULL DEFAULT 'system',
    language VARCHAR(10) NOT NULL DEFAULT 'es-PE',
    date_format VARCHAR(30) NOT NULL DEFAULT 'DD/MM/YYYY',
    number_format VARCHAR(30) NOT NULL DEFAULT 'es-PE',
    decimal_places INTEGER NOT NULL DEFAULT 2 CHECK (decimal_places BETWEEN 0 AND 6),
    timezone VARCHAR(50) NOT NULL DEFAULT 'America/Lima',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_local_profiles_singleton ON local_profiles ((TRUE));


CREATE TABLE audit_logs (
    id              UUID         NOT NULL DEFAULT gen_random_uuid(),
    occurred_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    company_id      UUID         NOT NULL,
    user_id         UUID,
    table_name      VARCHAR(100) NOT NULL,
    record_id       UUID,
    action          VARCHAR(30)  NOT NULL,
    old_value       JSONB,
    new_value       JSONB,
    changed_fields  TEXT[],
    ip_address      INET,
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

CREATE OR REPLACE FUNCTION audit_logs_forbid_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is append-only; % is not allowed', TG_OP;
END;
$$;

CREATE TRIGGER trg_audit_logs_no_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION audit_logs_forbid_mutation();

CREATE TRIGGER trg_audit_logs_no_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION audit_logs_forbid_mutation();



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


CREATE TABLE notifications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID         NOT NULL,
    type         VARCHAR(50)  NOT NULL,
    title        VARCHAR(200) NOT NULL,
    message      TEXT         NOT NULL,
    record_type  VARCHAR(50),
    record_id    UUID,
    dedup_key    VARCHAR(64)  NOT NULL,
    read_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

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

CREATE INDEX idx_notifications_company_unread
    ON notifications (company_id, read_at, created_at DESC);


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

INSERT INTO currencies (code, symbol, name, decimal_places, type, is_active) VALUES
    ('PEN', 'S/',  'Sol peruano',        2, 'fiat', TRUE),
    ('USD', '$',   'Dólar estadounidense', 2, 'fiat', TRUE),
    ('EUR', '€',   'Euro',               2, 'fiat', FALSE),
    ('MXN', 'MX$', 'Peso mexicano',      2, 'fiat', FALSE),
    ('COP', 'COP$', 'Peso colombiano',   2, 'fiat', FALSE),
    ('CLP', 'CLP$', 'Peso chileno',      0, 'fiat', FALSE),
    ('BRL', 'R$',  'Real brasileño',     2, 'fiat', FALSE),
    ('ARS', 'AR$', 'Peso argentino',     2, 'fiat', FALSE),
    ('BOB', 'Bs',  'Boliviano',          2, 'fiat', FALSE)
ON CONFLICT (code) DO NOTHING;


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

INSERT INTO taxes (company_id, code, name, short_name, country_code, default_rate, is_inclusive, is_percentage, category, is_active) VALUES
    (NULL, 'IGV',      'Impuesto General a las Ventas', 'IGV',     'PE', 0.1800, FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVAP',     'Impuesto de Promoción Municipal','IVAP',    'PE', 0.0200, FALSE, TRUE, 'municipal',TRUE),
    (NULL, 'RENTA',    'Impuesto a la Renta',            'Renta',   'PE', 0.2950, FALSE, TRUE, 'income',   TRUE),
    (NULL, 'EXONERADO','Exonerado del IGV',              'Exo',     'PE', 0.0000, FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'GRATUITO', 'Operación Gratuita',             'Grat',    'PE', 0.0000, FALSE, TRUE, 'sales',    TRUE),
    (NULL, 'IVA_MX',   'Impuesto al Valor Agregado',    'IVA',     'MX', 0.1600, FALSE, TRUE, 'sales',    FALSE),
    (NULL, 'IVA_CO',   'Impuesto sobre las Ventas',     'IVA',     'CO', 0.1900, FALSE, TRUE, 'sales',    FALSE)
ON CONFLICT DO NOTHING;



CREATE TABLE exchange_rates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID,
    from_currency   VARCHAR(3)     NOT NULL,
    to_currency     VARCHAR(3)     NOT NULL,
    rate_date       DATE           NOT NULL,
    rate            NUMERIC(18,6)  NOT NULL
        CHECK (rate > 0),
    source          VARCHAR(50)    NOT NULL DEFAULT 'manual'
        CHECK (source IN ('manual', 'central_bank', 'sunat', 'bloomberg', 'other', 'apis.net.pe', 'open.er-api.com', 'fallback')),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID         NOT NULL,
    user_id      UUID,
    session_id   UUID,
    event_type   VARCHAR(50)  NOT NULL
        CHECK (event_type IN (
            'CONFIG_UPDATE',
            'BACKUP_CREATE', 'EXPORT_DATA'
        )),
    target_type  VARCHAR(100),
    target_id    UUID,
    description  TEXT,
    metadata     JSONB        NOT NULL DEFAULT '{}',
    ip_address   INET,
    device       VARCHAR(100),
    occurred_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

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

CREATE OR REPLACE FUNCTION reject_audit_events_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only: UPDATE and DELETE are not allowed';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_events_no_mutation
    BEFORE UPDATE OR DELETE ON audit_events
    FOR EACH ROW
    EXECUTE FUNCTION reject_audit_events_mutation();


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


CREATE TABLE customers (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id         UUID           NOT NULL,
    default_branch_id  UUID,
    document_type      VARCHAR(10)    NOT NULL
        CHECK (document_type IN ('DNI', 'RUC', 'CE', 'PASSPORT')),
    document_number    VARCHAR(30)    NOT NULL,
    business_name      VARCHAR(200)   NOT NULL,
    trade_name         VARCHAR(200),
    tax_category       VARCHAR(20)    NOT NULL DEFAULT 'taxed'
        CHECK (tax_category IN ('taxed', 'exempt', 'unaffected', 'export')),
    credit_limit       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (credit_limit >= 0),
    current_debt       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (current_debt >= 0),
    payment_term_days  INTEGER        NOT NULL DEFAULT 0
        CHECK (payment_term_days >= 0),
    status             VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'blocked')),
    blocked_reason     TEXT,
    email              VARCHAR(200),
    phone              VARCHAR(30),
    address            TEXT,
    is_active          BOOLEAN        NOT NULL DEFAULT TRUE,

    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,
    created_by         UUID,
    updated_by         UUID,

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

CREATE TRIGGER trg_customers_set_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE suppliers (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id         UUID           NOT NULL,
    document_type      VARCHAR(10)    NOT NULL
        CHECK (document_type IN ('DNI', 'RUC', 'CE', 'PASSPORT')),
    document_number    VARCHAR(30)    NOT NULL,
    business_name      VARCHAR(200)   NOT NULL,
    trade_name         VARCHAR(200),
    tax_id             VARCHAR(30),
    is_international   BOOLEAN        NOT NULL DEFAULT FALSE,
    default_currency   VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    current_debt       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (current_debt >= 0),
    payment_term_days  INTEGER        NOT NULL DEFAULT 0
        CHECK (payment_term_days >= 0),
    status             VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),
    email              VARCHAR(200),
    phone              VARCHAR(30),
    address            TEXT,
    is_active          BOOLEAN        NOT NULL DEFAULT TRUE,

    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,
    created_by         UUID,
    updated_by         UUID,

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

CREATE TRIGGER trg_suppliers_set_updated_at
    BEFORE UPDATE ON suppliers
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE product_categories (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    parent_id    UUID,
    path         VARCHAR(100),
    depth        INTEGER        NOT NULL DEFAULT 0
        CHECK (depth >= 0),
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    created_by   UUID,
    updated_by   UUID,

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
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    created_by   UUID,
    updated_by   UUID,

    CONSTRAINT fk_product_brands_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_product_brands_code
    ON product_brands (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE units_of_measure (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    code             VARCHAR(20)    NOT NULL,
    name             VARCHAR(100)   NOT NULL,
    symbol           VARCHAR(20),
    allows_decimals  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_units_of_measure_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_units_of_measure_code
    ON units_of_measure (company_id, code);

CREATE TABLE warehouses (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID           NOT NULL,
    branch_id         UUID           NOT NULL,
    code              VARCHAR(20)    NOT NULL,
    name              VARCHAR(200)   NOT NULL,
    address           TEXT,
    manager_id        UUID,
    is_default        BOOLEAN        NOT NULL DEFAULT FALSE,
    allows_clearance  BOOLEAN        NOT NULL DEFAULT FALSE,
    is_active         BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        UUID,
    updated_by        UUID,

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
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID           NOT NULL,
    sku            VARCHAR(50)    NOT NULL,
    barcode        VARCHAR(50),
    description    TEXT           NOT NULL,
    category_id    UUID,
    brand_id       UUID,
    unit_id        UUID           NOT NULL,
    tax_id         UUID           NOT NULL,
    cost_usd       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (cost_usd >= 0),
    sale_price     NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (sale_price >= 0),
    sale_currency  VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    min_stock      NUMERIC(18,4)  NOT NULL DEFAULT 0,
    max_stock      NUMERIC(18,4)  NOT NULL DEFAULT 0,
    weight         NUMERIC(18,4)  NOT NULL DEFAULT 0,
    details        TEXT,
    is_active      BOOLEAN        NOT NULL DEFAULT TRUE,
    is_service     BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,
    created_by     UUID,
    updated_by     UUID,

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

CREATE TRIGGER trg_product_categories_set_updated_at
    BEFORE UPDATE ON product_categories
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_product_brands_set_updated_at
    BEFORE UPDATE ON product_brands
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_warehouses_set_updated_at
    BEFORE UPDATE ON warehouses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_products_set_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE fiscal_periods (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID           NOT NULL,
    name         VARCHAR(100)   NOT NULL,
    period_start DATE           NOT NULL,
    period_end   DATE           NOT NULL,
    status       VARCHAR(20)    NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closing', 'closed')),
    closed_at    TIMESTAMPTZ,
    closed_by    UUID,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by   UUID,
    updated_by   UUID,

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

CREATE TRIGGER trg_fiscal_periods_set_updated_at
    BEFORE UPDATE ON fiscal_periods
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE chart_of_accounts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    code             VARCHAR(20)    NOT NULL,
    name             VARCHAR(200)   NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('asset', 'liability', 'equity', 'income', 'expense')),
    parent_id        UUID,
    path             VARCHAR(100),
    depth            INTEGER        NOT NULL DEFAULT 1
        CHECK (depth >= 1),
    is_active        BOOLEAN        NOT NULL DEFAULT TRUE,
    allows_movement  BOOLEAN        NOT NULL DEFAULT TRUE,
    description      TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by       UUID,
    updated_by       UUID,

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

CREATE TRIGGER trg_chart_of_accounts_set_updated_at
    BEFORE UPDATE ON chart_of_accounts
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE journal_entries (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID           NOT NULL,
    fiscal_period_id  UUID           NOT NULL,
    number            VARCHAR(30)    NOT NULL,
    entry_date        DATE           NOT NULL,
    posting_date      DATE,
    description       TEXT,
    source            VARCHAR(20)    NOT NULL
        CHECK (source IN ('sale', 'purchase', 'payment', 'manual', 'adjustment', 'closing', 'opening')),
    source_id         UUID,
    status            VARCHAR(20)    NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'posted', 'reversed')),
    reverses_entry_id UUID,
    reversed_by_entry_id UUID,
    posted_at         TIMESTAMPTZ,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by        UUID,
    posted_by         UUID,

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
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id      UUID           NOT NULL,
    line_number           INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    account_id            UUID           NOT NULL,
    description           TEXT,
    debit                 NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (debit >= 0),
    credit                NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (credit >= 0),
    currency_code         VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate         NUMERIC(18,6)  NOT NULL DEFAULT 1,
    amount_in_txn_currency NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_journal_entry_lines_entry
        FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_journal_entry_lines_account
        FOREIGN KEY (account_id) REFERENCES chart_of_accounts(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_journal_entry_lines_one_side
        CHECK ((debit > 0) <> (credit > 0))
);

CREATE INDEX idx_journal_entry_lines_entry
    ON journal_entry_lines (journal_entry_id, line_number);

CREATE INDEX idx_journal_entry_lines_account
    ON journal_entry_lines (account_id);

CREATE TRIGGER trg_journal_entries_set_updated_at
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE bank_accounts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    branch_id        UUID,
    bank_name        VARCHAR(100)   NOT NULL,
    account_number   VARCHAR(50)    NOT NULL,
    account_type     VARCHAR(20)    NOT NULL DEFAULT 'checking'
        CHECK (account_type IN ('checking', 'savings')),
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    gl_account_id    UUID           NOT NULL,
    current_balance  NUMERIC(18,2)  NOT NULL DEFAULT 0,
    is_default       BOOLEAN        NOT NULL DEFAULT FALSE,
    is_active        BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    created_by       UUID,
    updated_by       UUID,

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
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID           NOT NULL,
    branch_id         UUID,
    issuer            VARCHAR(100)   NOT NULL,
    last_four         VARCHAR(4)     NOT NULL,
    card_holder       VARCHAR(200)   NOT NULL,
    expiration_month  INTEGER        NOT NULL
        CHECK (expiration_month BETWEEN 1 AND 12),
    expiration_year   INTEGER        NOT NULL
        CHECK (expiration_year BETWEEN 2000 AND 2100),
    credit_limit      NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (credit_limit >= 0),
    current_balance   NUMERIC(18,2)  NOT NULL DEFAULT 0,
    cut_off_day       INTEGER        NOT NULL DEFAULT 1
        CHECK (cut_off_day BETWEEN 1 AND 31),
    payment_due_day   INTEGER        NOT NULL DEFAULT 1
        CHECK (payment_due_day BETWEEN 1 AND 31),
    currency_code     VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    gl_account_id     UUID           NOT NULL,
    is_active         BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        UUID,
    updated_by        UUID,

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
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id  UUID           NOT NULL,
    transaction_date TIMESTAMPTZ    NOT NULL,
    value_date       TIMESTAMPTZ,
    description      TEXT,
    amount           NUMERIC(18,2)  NOT NULL,
    type             VARCHAR(20)    NOT NULL DEFAULT 'other'
        CHECK (type IN ('deposit', 'withdrawal', 'fee', 'interest', 'transfer', 'other')),
    reference        VARCHAR(100),
    balance_after    NUMERIC(18,2),
    is_reconciled    BOOLEAN        NOT NULL DEFAULT FALSE,
    reconciled_at    TIMESTAMPTZ,
    reconciled_by    UUID,
    journal_entry_id UUID,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_bank_accounts_set_updated_at
    BEFORE UPDATE ON bank_accounts
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_credit_cards_set_updated_at
    BEFORE UPDATE ON credit_cards
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_bank_transactions_set_updated_at
    BEFORE UPDATE ON bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE international_returns (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    supplier_id      UUID           NOT NULL,
    number           VARCHAR(30)    NOT NULL,
    return_date      DATE           NOT NULL,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'USD',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    subtotal         NUMERIC(18,2)  NOT NULL DEFAULT 0,
    total            NUMERIC(18,2)  NOT NULL DEFAULT 0,
    reason           TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'authorized', 'received', 'reconciled', 'cancelled')),
    authorized_at    TIMESTAMPTZ,
    authorized_by    UUID,
    received_at      TIMESTAMPTZ,
    reconciled_at    TIMESTAMPTZ,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by       UUID,
    updated_by       UUID,

    CONSTRAINT fk_international_returns_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_international_returns_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_international_returns_total
        CHECK (total >= 0)
);

CREATE UNIQUE INDEX uq_international_returns_number
    ON international_returns (company_id, number);

CREATE TABLE international_return_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    international_return_id UUID           NOT NULL,
    product_id              UUID           NOT NULL,
    quantity                NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity > 0),
    unit_cost               NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_cost >= 0),
    line_total              NUMERIC(18,2)  NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_international_returns_set_updated_at
    BEFORE UPDATE ON international_returns
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE purchase_orders (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    supplier_id      UUID           NOT NULL,
    branch_id        UUID,
    number           VARCHAR(30)    NOT NULL,
    order_date       DATE           NOT NULL,
    expected_date    DATE,
    received_date    DATE,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'received', 'paid', 'reconciled', 'cancelled')),
    subtotal         NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (subtotal >= 0),
    tax_amount       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    total            NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (total >= 0),
    paid_amount      NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (paid_amount >= 0),
    discount_amount  NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes            TEXT,
    order_type       VARCHAR(20)    NOT NULL DEFAULT 'general'
        CHECK (order_type IN ('general', 'customer')),
    customer_id      UUID,
    credit_card_id   UUID,
    supplier_order_number VARCHAR(100),
    arrival_date     DATE,
    cost_usd         NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (cost_usd >= 0),
    sale_price_pen   NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (sale_price_pen >= 0),
    real_cost_pen    NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (real_cost_pen >= 0),
    projected_profit_pen NUMERIC(18,2) NOT NULL DEFAULT 0,
    anticipo         NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (anticipo >= 0),
    anticipo_date    DATE,
    faulty           BOOLEAN        NOT NULL DEFAULT FALSE,
    faulty_reason    TEXT,
    refunded_amount  NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (refunded_amount >= 0),
    cancelled_at     TIMESTAMPTZ,
    cancelled_reason TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    created_by       UUID,
    updated_by       UUID,

    CONSTRAINT fk_purchase_orders_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_purchase_orders_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_purchase_orders_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_purchase_orders_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_purchase_orders_card
        FOREIGN KEY (credit_card_id) REFERENCES credit_cards(id)
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

CREATE INDEX idx_purchase_orders_order_type
    ON purchase_orders (company_id, order_type, status, order_date)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_purchase_orders_customer
    ON purchase_orders (company_id, customer_id, order_date)
    WHERE deleted_at IS NULL;

CREATE TABLE customer_order_payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID           NOT NULL,
    purchase_order_id UUID           NOT NULL,
    customer_id       UUID           NOT NULL,
    number            VARCHAR(30)    NOT NULL,
    payment_date      DATE           NOT NULL,
    amount            NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (amount > 0),
    payment_method    VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    currency_code     VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate     NUMERIC(18,6)  NOT NULL DEFAULT 1,
    reference         VARCHAR(100),
    notes             TEXT,
    status            VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'refunded')),
    refunded_amount   NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (refunded_amount >= 0),
    refunded_at       TIMESTAMPTZ,
    refund_reason     TEXT,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        UUID,
    updated_by        UUID,

    CONSTRAINT fk_customer_order_payments_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customer_order_payments_order
        FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customer_order_payments_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_customer_order_payments_number
    ON customer_order_payments (company_id, number)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customer_order_payments_order
    ON customer_order_payments (purchase_order_id, payment_date);

CREATE INDEX idx_customer_order_payments_customer
    ON customer_order_payments (customer_id, payment_date);

CREATE TABLE purchase_order_items (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id  UUID           NOT NULL,
    product_id         UUID           NOT NULL,
    line_number        INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    quantity_ordered   NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity_ordered > 0),
    quantity_received  NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity_received >= 0),
    unit_cost          NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_cost >= 0),
    discount_percent   NUMERIC(18,4)  NOT NULL DEFAULT 0,
    discount_amount    NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    tax_rate           NUMERIC(18,4)  NOT NULL DEFAULT 0,
    tax_amount         NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    line_total         NUMERIC(18,2)  NOT NULL DEFAULT 0,
    description        TEXT,
    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_purchase_order_items_order
        FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_purchase_order_items_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_purchase_order_items_received
        CHECK (quantity_received <= quantity_ordered)
);

CREATE INDEX idx_purchase_order_items_order
    ON purchase_order_items (purchase_order_id);

CREATE INDEX idx_purchase_order_items_product
    ON purchase_order_items (product_id);

CREATE TABLE inventory_batches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID           NOT NULL,
    product_id      UUID           NOT NULL,
    warehouse_id    UUID           NOT NULL,
    supplier_id     UUID,
    purchase_order_item_id UUID,
    lot             VARCHAR(100),
    batch_code      VARCHAR(50),
    arrival_date    TIMESTAMPTZ    NOT NULL,
    expiry_date     TIMESTAMPTZ,
    quantity        NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity >= 0),
    original_quantity NUMERIC(18,4) NOT NULL DEFAULT 0
        CHECK (original_quantity >= 0),
    unit_cost       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_cost >= 0),
    currency_code   VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate   NUMERIC(18,6)  NOT NULL DEFAULT 1,
    status          VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'depleted', 'written_off', 'voided')),
    clearance_date  TIMESTAMPTZ,
    is_clearance    BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

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
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    batch_id         UUID           NOT NULL,
    product_id       UUID           NOT NULL,
    warehouse_id     UUID           NOT NULL,
    movement_date    TIMESTAMPTZ    NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('purchase', 'purchase_receipt', 'sale', 'void_sale', 'void_purchase', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out', 'void_out')),
    reference_type   VARCHAR(30),
    reference_id     UUID,
    quantity_delta   NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity_delta <> 0),
    balance_after    NUMERIC(18,4)  NOT NULL DEFAULT 0,
    unit_cost        NUMERIC(18,2)  NOT NULL DEFAULT 0,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    notes            TEXT,
    created_by       UUID,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_inventory_batches_set_updated_at
    BEFORE UPDATE ON inventory_batches
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE sales (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID           NOT NULL,
    branch_id      UUID           NOT NULL,
    customer_id    UUID           NOT NULL,
    number         VARCHAR(30)    NOT NULL,
    sale_date      DATE           NOT NULL,
    due_date       DATE,
    status         VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'partial', 'paid', 'cancelled')),
    subtotal       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (subtotal >= 0),
    tax_amount     NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    total          NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (total >= 0),
    paid_amount    NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (paid_amount >= 0),
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    cost_total     NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (cost_total >= 0),
    profit         NUMERIC(18,2)  NOT NULL DEFAULT 0,
    currency_code  VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate  NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes          TEXT,
    cancelled_at   TIMESTAMPTZ,
    cancelled_reason TEXT,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,
    created_by     UUID,
    updated_by     UUID,

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
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id          UUID           NOT NULL,
    product_id       UUID           NOT NULL,
    inventory_batch_id UUID,
    line_number      INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    quantity         NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity > 0),
    unit_price       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_price >= 0),
    discount_percent NUMERIC(18,4)  NOT NULL DEFAULT 0,
    discount_amount  NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    tax_rate         NUMERIC(18,4)  NOT NULL DEFAULT 0,
    tax_amount       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    line_total       NUMERIC(18,2)  NOT NULL DEFAULT 0,
    cost_snapshot    NUMERIC(18,2)  NOT NULL DEFAULT 0,
    description      TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_sales_set_updated_at
    BEFORE UPDATE ON sales
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE customer_payments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    customer_id      UUID           NOT NULL,
    branch_id        UUID,
    number           VARCHAR(30)    NOT NULL,
    payment_date     DATE           NOT NULL,
    amount           NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (amount > 0),
    payment_method   VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    reference        VARCHAR(100),
    bank_account_id  UUID,
    cash_register_id UUID,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes            TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    created_by       UUID,
    updated_by       UUID,

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
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_payment_id UUID          NOT NULL,
    sale_id            UUID           NOT NULL,
    allocated_amount   NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (allocated_amount > 0),
    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    customer_id      UUID           NOT NULL,
    number           VARCHAR(30)    NOT NULL,
    advance_date     DATE           NOT NULL,
    amount           NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (amount >= 0),
    remaining        NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (remaining >= 0),
    payment_method   VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    reference        VARCHAR(100),
    bank_account_id  UUID,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes            TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    created_by       UUID,
    updated_by       UUID,

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
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_advance_id  UUID           NOT NULL,
    sale_id              UUID           NOT NULL,
    applied_amount       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (applied_amount > 0),
    created_at           TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_customer_payments_set_updated_at
    BEFORE UPDATE ON customer_payments
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_customer_advances_set_updated_at
    BEFORE UPDATE ON customer_advances
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();


CREATE TABLE supplier_payments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    supplier_id      UUID           NOT NULL,
    branch_id        UUID,
    number           VARCHAR(30)    NOT NULL,
    payment_date     DATE           NOT NULL,
    amount           NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (amount > 0),
    payment_method   VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    reference        VARCHAR(100),
    bank_account_id  UUID,
    cash_register_id UUID,
    credit_card_id   UUID,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes            TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    created_by       UUID,
    updated_by       UUID,

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
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_payment_id UUID           NOT NULL,
    purchase_order_id   UUID           NOT NULL,
    allocated_amount    NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (allocated_amount > 0),
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_purchase_orders_set_updated_at
    BEFORE UPDATE ON purchase_orders
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_supplier_payments_set_updated_at
    BEFORE UPDATE ON supplier_payments
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
