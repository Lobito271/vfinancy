-- 0030_sales.up.sql (SQLite)
-- Module: Sales / Accounts Receivable
-- Sales orders and lines. cost_snapshot captures the batch unit cost at
-- sale time for margin reporting. A sale touching inventory and AR must
-- run inside a single DB transaction (inventory → sale → journal → AR).

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
