-- 0025_chart_of_accounts.up.sql (PostgreSQL)
-- Module: Accounting
-- Company chart of accounts. Account type drives the normal balance
-- (asset/expense = debit; liability/equity/income = credit) and which
-- financial statement an account belongs to (P&L vs Balance Sheet).

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
