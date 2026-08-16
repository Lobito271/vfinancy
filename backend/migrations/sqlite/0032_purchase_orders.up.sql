-- 0032_purchase_orders.up.sql (SQLite)
-- Module: Purchasing / Accounts Payable
-- Purchase orders to local suppliers with received quantities and the
-- supplier payment (AP) ledger. PurchaseOrderItem links the arriving
-- inventory batch back to the purchase for lotting (0029).

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
