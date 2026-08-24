-- 0040_imports_orders.up.sql (PostgreSQL)
-- Module: Purchasing / Imports ERP
--
-- Mirrors backend/migrations/sqlite/0040_imports_orders.up.sql.

ALTER TABLE purchase_orders
    ADD COLUMN order_type         VARCHAR(20)  NOT NULL DEFAULT 'general'
        CHECK (order_type IN ('general', 'customer')),
    ADD COLUMN customer_id        UUID,
    ADD COLUMN arrival_date       DATE,
    ADD COLUMN cost_usd           NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (cost_usd >= 0),
    ADD COLUMN sale_price_pen     NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (sale_price_pen >= 0),
    ADD COLUMN real_cost_pen      NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (real_cost_pen >= 0),
    ADD COLUMN projected_profit_pen NUMERIC(18,2) NOT NULL DEFAULT 0,
    ADD COLUMN anticipo           NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (anticipo >= 0),
    ADD COLUMN anticipo_date      DATE,
    ADD COLUMN faulty             BOOLEAN       NOT NULL DEFAULT FALSE,
    ADD COLUMN faulty_reason      TEXT,
    ADD COLUMN refunded_amount    NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (refunded_amount >= 0);

ALTER TABLE purchase_orders
    ADD CONSTRAINT fk_purchase_orders_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX idx_purchase_orders_order_type
    ON purchase_orders (company_id, order_type, status, order_date)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_purchase_orders_customer
    ON purchase_orders (company_id, customer_id, order_date)
    WHERE deleted_at IS NULL;

ALTER TABLE products ADD COLUMN details TEXT;

CREATE TABLE customer_order_payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID           NOT NULL,
    purchase_order_id UUID           NOT NULL,
    customer_id       UUID           NOT NULL,
    number            VARCHAR(30)    NOT NULL,
    payment_date      DATE           NOT NULL,
    amount            NUMERIC(18,2)  NOT NULL DEFAULT 0 CHECK (amount > 0),
    payment_method    VARCHAR(20)    NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'bank_transfer', 'check', 'card', 'credit', 'other')),
    currency_code     VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate     NUMERIC(18,6)  NOT NULL DEFAULT 1,
    reference         VARCHAR(100),
    notes             TEXT,
    status            VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'refunded')),
    refunded_amount   NUMERIC(18,2)  NOT NULL DEFAULT 0 CHECK (refunded_amount >= 0),
    refunded_at       TIMESTAMPTZ,
    refund_reason     TEXT,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        UUID,
    updated_by        UUID,

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
