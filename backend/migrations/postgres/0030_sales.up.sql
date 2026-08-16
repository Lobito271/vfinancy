-- 0030_sales.up.sql (PostgreSQL)
-- Module: Sales / Accounts Receivable
-- Sales orders and lines. cost_snapshot captures the batch unit cost at
-- sale time for margin reporting. A sale touching inventory and AR must
-- run inside a single DB transaction (inventory → sale → journal → AR).

CREATE TABLE sales (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID           NOT NULL,
    branch_id      UUID           NOT NULL,
    customer_id    UUID           NOT NULL,
    number         VARCHAR(30)    NOT NULL,
    sale_date      DATE           NOT NULL,
    due_date       DATE,
    status         VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'partial', 'paid', 'cancelled')),
    subtotal       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (subtotal >= 0),
    tax_amount     NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    total          NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (total >= 0),
    paid_amount    NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (paid_amount >= 0),
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    cost_total     NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (cost_total >= 0),
    profit         NUMERIC(18,2)  NOT NULL DEFAULT 0,
    currency_code  VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate  NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes          TEXT,
    cancelled_at   TIMESTAMPTZ,
    cancelled_reason TEXT,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,
    created_by     UUID,
    updated_by     UUID,

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
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id          UUID           NOT NULL,
    product_id       UUID           NOT NULL,
    inventory_batch_id UUID,
    line_number      INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    quantity         NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity > 0),
    unit_price       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_price >= 0),
    discount_percent NUMERIC(18,4)  NOT NULL DEFAULT 0,
    discount_amount  NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    tax_rate         NUMERIC(18,4)  NOT NULL DEFAULT 0,
    tax_amount       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    line_total       NUMERIC(18,2)  NOT NULL DEFAULT 0,
    cost_snapshot    NUMERIC(18,2)  NOT NULL DEFAULT 0,
    description      TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_sales_set_updated_at
    BEFORE UPDATE ON sales
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
