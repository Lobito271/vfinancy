-- 0027_bank_accounts.up.sql (PostgreSQL)
-- Module: Treasury / Reconciliation
-- Company bank accounts, the bank-transaction ledger that is reconciled
-- against bank statements, and company-issued credit cards.

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
