-- 0040_imports_orders.up.sql (SQLite)
-- Module: Purchasing / Imports ERP
--
-- Imports ERP: purchase orders become either "general" (store stock)
-- or "customer" (specific customer order). Customer orders support
-- partial down payments (anticipo) recorded against the order, with
-- the remaining balance tracked as "por cobrar" until the goods
-- arrive. Import costs are tracked in USD and converted to PEN using
-- the order's exchange rate plus a fixed 0.07 import surcharge; the
-- projected profit is the expected PEN sale price minus that real
-- cost. A faulty-arrival workflow records the arrival date, voids the
-- order (restoring inventory) and refunds every customer down payment.

ALTER TABLE purchase_orders ADD COLUMN order_type TEXT NOT NULL DEFAULT 'general';
ALTER TABLE purchase_orders ADD COLUMN customer_id TEXT;
ALTER TABLE purchase_orders ADD COLUMN arrival_date TIMESTAMP;
ALTER TABLE purchase_orders ADD COLUMN cost_usd TEXT NOT NULL DEFAULT '0.00';
ALTER TABLE purchase_orders ADD COLUMN sale_price_pen TEXT NOT NULL DEFAULT '0.00';
ALTER TABLE purchase_orders ADD COLUMN real_cost_pen TEXT NOT NULL DEFAULT '0.00';
ALTER TABLE purchase_orders ADD COLUMN projected_profit_pen TEXT NOT NULL DEFAULT '0.00';
ALTER TABLE purchase_orders ADD COLUMN anticipo TEXT NOT NULL DEFAULT '0.00';
ALTER TABLE purchase_orders ADD COLUMN anticipo_date TIMESTAMP;
ALTER TABLE purchase_orders ADD COLUMN faulty INTEGER NOT NULL DEFAULT 0;
ALTER TABLE purchase_orders ADD COLUMN faulty_reason TEXT;
ALTER TABLE purchase_orders ADD COLUMN refunded_amount TEXT NOT NULL DEFAULT '0.00';

CREATE INDEX idx_purchase_orders_order_type
    ON purchase_orders (company_id, order_type, status, order_date)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_purchase_orders_customer
    ON purchase_orders (company_id, customer_id, order_date)
    WHERE deleted_at IS NULL;

ALTER TABLE products ADD COLUMN details TEXT;

CREATE TABLE customer_order_payments (
    id                TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id        TEXT           NOT NULL,
    purchase_order_id TEXT           NOT NULL,
    customer_id       TEXT           NOT NULL,
    number            VARCHAR(30)    NOT NULL,
    payment_date      TIMESTAMP      NOT NULL,
    amount            TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(amount AS REAL) > 0),
    payment_method    VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    currency_code     VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate     TEXT           NOT NULL DEFAULT '1.000000',
    reference         VARCHAR(100),
    notes             TEXT,
    status            VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'refunded')),
    refunded_amount   TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(refunded_amount AS REAL) >= 0),
    refunded_at       TIMESTAMP,
    refund_reason     TEXT,
    created_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at        TIMESTAMP,
    created_by        TEXT,
    updated_by        TEXT,

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
