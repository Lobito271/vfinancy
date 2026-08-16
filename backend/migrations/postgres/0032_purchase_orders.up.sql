-- 0032_purchase_orders.up.sql (PostgreSQL)
-- Module: Purchasing / Accounts Payable
-- Purchase orders to local suppliers with received quantities and the
-- supplier payment (AP) ledger. PurchaseOrderItem links the arriving
-- inventory batch back to the purchase for lotting (0029).

CREATE TABLE purchase_orders (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    supplier_id      UUID           NOT NULL,
    branch_id        UUID,
    number           VARCHAR(30)    NOT NULL,
    order_date       DATE           NOT NULL,
    expected_date    DATE,
    received_date    DATE,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'received', 'paid', 'reconciled', 'cancelled')),
    subtotal         NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (subtotal >= 0),
    tax_amount       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    total            NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (total >= 0),
    paid_amount      NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (paid_amount >= 0),
    discount_amount  NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes            TEXT,
    cancelled_at     TIMESTAMPTZ,
    cancelled_reason TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    created_by       UUID,
    updated_by       UUID,

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
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id  UUID           NOT NULL,
    product_id         UUID           NOT NULL,
    line_number        INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    quantity_ordered   NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity_ordered > 0),
    quantity_received  NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity_received >= 0),
    unit_cost          NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_cost >= 0),
    discount_percent   NUMERIC(18,4)  NOT NULL DEFAULT 0,
    discount_amount    NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (discount_amount >= 0),
    tax_rate           NUMERIC(18,4)  NOT NULL DEFAULT 0,
    tax_amount         NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (tax_amount >= 0),
    line_total         NUMERIC(18,2)  NOT NULL DEFAULT 0,
    description        TEXT,
    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_purchase_order_items_order
        FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_purchase_order_items_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_purchase_order_items_received
        CHECK (quantity_received <= quantity_ordered)
);

CREATE INDEX idx_purchase_order_items_order
    ON purchase_order_items (purchase_order_id);

CREATE INDEX idx_purchase_order_items_product
    ON purchase_order_items (product_id);

CREATE TABLE supplier_payments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    supplier_id      UUID           NOT NULL,
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
    credit_card_id   UUID,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    notes            TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    created_by       UUID,
    updated_by       UUID,

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
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_payment_id UUID           NOT NULL,
    purchase_order_id   UUID           NOT NULL,
    allocated_amount    NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (allocated_amount > 0),
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_purchase_orders_set_updated_at
    BEFORE UPDATE ON purchase_orders
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_supplier_payments_set_updated_at
    BEFORE UPDATE ON supplier_payments
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
