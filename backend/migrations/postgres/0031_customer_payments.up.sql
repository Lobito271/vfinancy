-- 0031_customer_payments.up.sql (PostgreSQL)
-- Module: Sales / Accounts Receivable
-- Customer payments (cash received, checks, transfers, cards) allocated
-- to one or more open invoices. Unallocated overpayments become a
-- customer advance (credit against future sales).

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
