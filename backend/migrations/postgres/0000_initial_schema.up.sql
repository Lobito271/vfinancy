-- vfinancy cloud mirror schema (PostgreSQL). Contains exactly the replicated
-- set: every business table except device-local data (local_profiles, sync bookkeeping).
-- Applied by `cli migrate --postgres` or by the sync worker on first connect.

CREATE TABLE application_settings (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key        VARCHAR(100) NOT NULL,
    value      TEXT         NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT ck_settings_key_nonblank CHECK (length(trim(key)) > 0)
);

CREATE UNIQUE INDEX uq_settings_key ON application_settings (key);

CREATE TABLE exchange_rates (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency VARCHAR(3)  NOT NULL,
    to_currency   VARCHAR(3)  NOT NULL,
    rate_date     DATE        NOT NULL,
    rate          TEXT        NOT NULL CHECK (CAST(rate AS numeric) > 0),
    source        VARCHAR(50) NOT NULL DEFAULT 'manual'
        CHECK (source IN ('manual', 'sunat', 'other', 'apis.net.pe', 'open.er-api.com', 'fallback')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT ck_exchange_rates_different_currencies CHECK (from_currency <> to_currency)
);

CREATE UNIQUE INDEX uq_exchange_rates_pair_date ON exchange_rates (from_currency, to_currency, rate_date);

CREATE TABLE customers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type   VARCHAR(3),
    document_number VARCHAR(30),
    business_name   VARCHAR(200) NOT NULL,
    email           VARCHAR(200),
    phone           VARCHAR(30),
    address         TEXT,
    current_debt    TEXT         NOT NULL DEFAULT '0.00' CHECK (CAST(current_debt AS numeric) >= 0),
    status          VARCHAR(20)  NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    created_by      TEXT,
    updated_by      TEXT,

    CONSTRAINT ck_customers_business_name_nonblank CHECK (length(trim(business_name)) > 0),
    CONSTRAINT ck_customers_doc_pair
        CHECK ((document_type IS NULL AND document_number IS NULL)
            OR (document_type IS NOT NULL AND document_number IS NOT NULL)),
    CONSTRAINT ck_customers_doc_type CHECK (document_type IS NULL OR document_type IN ('DNI', 'RUC')),
    CONSTRAINT ck_customers_dni
        CHECK (document_type <> 'DNI' OR document_number ~ '^[0-9]{8}$'),
    CONSTRAINT ck_customers_ruc
        CHECK (document_type <> 'RUC' OR document_number ~ '^(10|20)[0-9]{9}$'),
    CONSTRAINT ck_customers_status CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX uq_customers_document
    ON customers (document_type, document_number)
    WHERE deleted_at IS NULL AND document_number IS NOT NULL;

CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku         VARCHAR(50)  NOT NULL,
    description TEXT         NOT NULL,
    unit_code   VARCHAR(20)  NOT NULL DEFAULT 'Unidad',
    cost_usd    TEXT         NOT NULL DEFAULT '0.00' CHECK (CAST(cost_usd AS numeric) >= 0),
    sale_price  TEXT         NOT NULL DEFAULT '0.00' CHECK (CAST(sale_price AS numeric) >= 0),
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,

    CONSTRAINT ck_products_description_nonblank CHECK (length(trim(description)) > 0)
);

CREATE UNIQUE INDEX uq_products_sku ON products (sku) WHERE deleted_at IS NULL;

CREATE TABLE credit_cards (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer           VARCHAR(100) NOT NULL,
    last_four        VARCHAR(4)   NOT NULL CHECK (last_four ~ '^[0-9]{4}$'),
    credit_limit     TEXT         NOT NULL DEFAULT '0.00' CHECK (CAST(credit_limit AS numeric) >= 0),
    current_balance  TEXT         NOT NULL DEFAULT '0.00' CHECK (CAST(current_balance AS numeric) >= 0),
    cut_off_day      INTEGER      NOT NULL CHECK (cut_off_day BETWEEN 1 AND 31),
    payment_due_day  INTEGER      NOT NULL CHECK (payment_due_day BETWEEN 1 AND 31),
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ,
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT ck_credit_cards_issuer_nonblank CHECK (length(trim(issuer)) > 0)
);

CREATE INDEX idx_credit_cards_active ON credit_cards (is_active) WHERE deleted_at IS NULL;

CREATE TABLE purchase_orders (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number         VARCHAR(30) NOT NULL,
    order_date     TIMESTAMPTZ NOT NULL,
    expected_date  TIMESTAMPTZ,
    received_date  TIMESTAMPTZ,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'received', 'cancelled')),
    currency_code  VARCHAR(3)  NOT NULL DEFAULT 'USD',
    exchange_rate  TEXT        NOT NULL DEFAULT '1.000000' CHECK (CAST(exchange_rate AS numeric) > 0),
    notes          TEXT,
    order_type     VARCHAR(20) NOT NULL DEFAULT 'general' CHECK (order_type IN ('general', 'customer')),
    customer_id    UUID,
    credit_card_id UUID,
    arrival_date   TIMESTAMPTZ,
    cost_usd       TEXT        NOT NULL DEFAULT '0.00' CHECK (CAST(cost_usd AS numeric) >= 0),
    sale_price_pen TEXT        NOT NULL DEFAULT '0.00' CHECK (CAST(sale_price_pen AS numeric) >= 0),
    real_cost_pen  TEXT        NOT NULL DEFAULT '0.00' CHECK (CAST(real_cost_pen AS numeric) >= 0),
    projected_profit_pen TEXT NOT NULL DEFAULT '0.00',
    refund_amount  TEXT        NOT NULL DEFAULT '0.00' CHECK (CAST(refund_amount AS numeric) >= 0),
    faulty         BOOLEAN     NOT NULL DEFAULT FALSE,
    faulty_reason  TEXT,
    cancelled_at   TIMESTAMPTZ,
    cancelled_reason TEXT,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    created_by     TEXT,
    updated_by     TEXT,

    CONSTRAINT fk_purchase_orders_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_purchase_orders_card
        FOREIGN KEY (credit_card_id) REFERENCES credit_cards(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_purchase_orders_dates CHECK (expected_date IS NULL OR expected_date >= order_date)
);

CREATE UNIQUE INDEX uq_purchase_orders_number ON purchase_orders (number) WHERE deleted_at IS NULL;
CREATE INDEX idx_purchase_orders_status ON purchase_orders (status, order_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_purchase_orders_type ON purchase_orders (order_type, status, order_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_purchase_orders_card ON purchase_orders (credit_card_id, order_date) WHERE deleted_at IS NULL;

CREATE TABLE purchase_order_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID       NOT NULL,
    product_id        UUID,
    line_number       INTEGER    NOT NULL DEFAULT 1 CHECK (line_number >= 1),
    description       TEXT       NOT NULL,
    unit_code         VARCHAR(20) NOT NULL DEFAULT 'Unidad',
    quantity_ordered  TEXT       NOT NULL CHECK (CAST(quantity_ordered AS numeric) > 0),
    quantity_received TEXT       NOT NULL DEFAULT '0.0000' CHECK (CAST(quantity_received AS numeric) >= 0),
    unit_cost_usd     TEXT       NOT NULL CHECK (CAST(unit_cost_usd AS numeric) >= 0),
    line_total_usd    TEXT       NOT NULL DEFAULT '0.00',
    sale_price_pen    TEXT       NOT NULL DEFAULT '0.00' CHECK (CAST(sale_price_pen AS numeric) >= 0),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_purchase_order_items_order
        FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_purchase_order_items_product
        FOREIGN KEY (product_id) REFERENCES products(id) ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT ck_purchase_order_items_description_nonblank CHECK (length(trim(description)) > 0)
);

CREATE INDEX idx_purchase_order_items_order ON purchase_order_items (purchase_order_id);
CREATE INDEX idx_purchase_order_items_product ON purchase_order_items (product_id);

CREATE TABLE inventory_batches (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id             UUID       NOT NULL,
    purchase_order_item_id UUID,
    arrival_date           TIMESTAMPTZ NOT NULL,
    quantity               TEXT       NOT NULL DEFAULT '0.0000' CHECK (CAST(quantity AS numeric) >= 0),
    original_quantity      TEXT       NOT NULL DEFAULT '0.0000' CHECK (CAST(original_quantity AS numeric) >= 0),
    unit_cost              TEXT       NOT NULL DEFAULT '0.00' CHECK (CAST(unit_cost AS numeric) >= 0),
    exchange_rate          TEXT       NOT NULL DEFAULT '1.000000' CHECK (CAST(exchange_rate AS numeric) > 0),
    status                 VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'depleted', 'voided')),
    is_clearance           BOOLEAN    NOT NULL DEFAULT FALSE,

    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by             TEXT,
    updated_by             TEXT,

    CONSTRAINT fk_inventory_batches_product
        FOREIGN KEY (product_id) REFERENCES products(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_batches_poi
        FOREIGN KEY (purchase_order_item_id) REFERENCES purchase_order_items(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_inventory_batches_product ON inventory_batches (product_id, status);
CREATE INDEX idx_inventory_batches_clearance ON inventory_batches (is_clearance, status) WHERE is_clearance = TRUE AND status = 'active';

CREATE TABLE inventory_movements (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id       UUID      NOT NULL,
    product_id     UUID      NOT NULL,
    movement_date  TIMESTAMPTZ NOT NULL,
    type           VARCHAR(20) NOT NULL
        CHECK (type IN ('purchase_receipt', 'sale', 'void_sale', 'void_purchase', 'adjustment_in', 'adjustment_out')),
    reference_type VARCHAR(30),
    reference_id   TEXT,
    quantity_delta TEXT      NOT NULL CHECK (CAST(quantity_delta AS numeric) <> 0),
    balance_after  TEXT      NOT NULL DEFAULT '0.0000',
    unit_cost      TEXT      NOT NULL DEFAULT '0.00',
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_inventory_movements_batch
        FOREIGN KEY (batch_id) REFERENCES inventory_batches(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_product
        FOREIGN KEY (product_id) REFERENCES products(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_inventory_movements_batch ON inventory_movements (batch_id, movement_date);
CREATE INDEX idx_inventory_movements_product ON inventory_movements (product_id, movement_date);
CREATE INDEX idx_inventory_movements_reference ON inventory_movements (reference_type, reference_id);

CREATE TABLE sales (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id      UUID        NOT NULL,
    number           VARCHAR(30) NOT NULL,
    sale_date        TIMESTAMPTZ NOT NULL,
    due_date         TIMESTAMPTZ,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'partial', 'paid', 'cancelled')),
    sale_type        VARCHAR(20) NOT NULL DEFAULT 'stock' CHECK (sale_type IN ('stock', 'client_order')),
    total            TEXT        NOT NULL DEFAULT '0.00' CHECK (CAST(total AS numeric) >= 0),
    paid_amount      TEXT        NOT NULL DEFAULT '0.00' CHECK (CAST(paid_amount AS numeric) >= 0),
    cost_total       TEXT        NOT NULL DEFAULT '0.00' CHECK (CAST(cost_total AS numeric) >= 0),
    profit           TEXT        NOT NULL DEFAULT '0.00',
    notes            TEXT,
    cancelled_at     TIMESTAMPTZ,
    cancelled_reason TEXT,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ,
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_sales_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_sales_dates CHECK (due_date IS NULL OR due_date >= sale_date)
);

CREATE UNIQUE INDEX uq_sales_number ON sales (number) WHERE deleted_at IS NULL;
CREATE INDEX idx_sales_customer ON sales (customer_id, sale_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_sales_status ON sales (status, sale_date) WHERE deleted_at IS NULL;

CREATE TABLE sale_items (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id            UUID      NOT NULL,
    product_id         UUID      NOT NULL,
    inventory_batch_id UUID,
    line_number        INTEGER   NOT NULL DEFAULT 1 CHECK (line_number >= 1),
    quantity           TEXT      NOT NULL CHECK (CAST(quantity AS numeric) > 0),
    unit_price         TEXT      NOT NULL CHECK (CAST(unit_price AS numeric) >= 0),
    line_total         TEXT      NOT NULL DEFAULT '0.00',
    cost_snapshot      TEXT      NOT NULL DEFAULT '0.00',
    description        TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_sale_items_sale
        FOREIGN KEY (sale_id) REFERENCES sales(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_sale_items_product
        FOREIGN KEY (product_id) REFERENCES products(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_sale_items_batch
        FOREIGN KEY (inventory_batch_id) REFERENCES inventory_batches(id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_sale_items_sale ON sale_items (sale_id);
CREATE INDEX idx_sale_items_product ON sale_items (product_id);

CREATE TABLE customer_payments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id    UUID        NOT NULL,
    number         VARCHAR(30) NOT NULL,
    payment_date   TIMESTAMPTZ NOT NULL,
    amount         TEXT        NOT NULL CHECK (CAST(amount AS numeric) > 0),
    payment_method VARCHAR(20) NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'transfer', 'other')),
    reference      VARCHAR(100),
    notes          TEXT,
    status         VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'refunded')),

    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    created_by     TEXT,
    updated_by     TEXT,

    CONSTRAINT fk_customer_payments_customer
        FOREIGN KEY (customer_id) REFERENCES customers(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_customer_payments_number ON customer_payments (number) WHERE deleted_at IS NULL;
CREATE INDEX idx_customer_payments_customer ON customer_payments (customer_id, payment_date) WHERE deleted_at IS NULL;

CREATE TABLE customer_payment_allocations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_payment_id UUID      NOT NULL,
    sale_id             UUID      NOT NULL,
    allocated_amount    TEXT      NOT NULL CHECK (CAST(allocated_amount AS numeric) > 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_customer_payment_alloc_payment
        FOREIGN KEY (customer_payment_id) REFERENCES customer_payments(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_customer_payment_alloc_sale
        FOREIGN KEY (sale_id) REFERENCES sales(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_customer_payment_alloc_payment ON customer_payment_allocations (customer_payment_id);
CREATE INDEX idx_customer_payment_alloc_sale ON customer_payment_allocations (sale_id);

CREATE TABLE import_lots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(30) NOT NULL,
    description TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT
);

CREATE UNIQUE INDEX uq_import_lots_code ON import_lots (code) WHERE deleted_at IS NULL;

CREATE TABLE import_lot_purchase_orders (
    import_lot_id     UUID      NOT NULL,
    purchase_order_id UUID      NOT NULL,
    added_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (import_lot_id, purchase_order_id),
    CONSTRAINT fk_import_lot_members_lot
        FOREIGN KEY (import_lot_id) REFERENCES import_lots(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_import_lot_members_purchase
        FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_import_lot_members_purchase ON import_lot_purchase_orders (purchase_order_id);
