-- 0027_bank_accounts.up.sql (SQLite)
-- Module: Treasury / Reconciliation
-- Company bank accounts, the bank-transaction ledger that is reconciled
-- against bank statements, and company-issued credit cards.

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
