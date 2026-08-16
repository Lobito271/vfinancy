-- 0022_suppliers.up.sql (PostgreSQL)
-- Module: Purchasing / Accounts Payable
-- Supplier master. International suppliers carry their own default
-- currency (e.g. USD) and drive the international-returns flow.

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
