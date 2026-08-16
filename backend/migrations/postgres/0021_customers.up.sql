-- 0021_customers.up.sql (PostgreSQL)
-- Module: Sales / Accounts Receivable
-- Normalized customer master. current_debt is maintained by the sales
-- and payment services; AR aging is derived from the sales table.

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
