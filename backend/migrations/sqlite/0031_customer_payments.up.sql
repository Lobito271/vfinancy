-- 0031_customer_payments.up.sql (SQLite)
-- Module: Sales / Accounts Receivable
-- Customer payments (cash received, checks, transfers, cards) allocated
-- to one or more open invoices. Unallocated overpayments become a
-- customer advance (credit against future sales).

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
