-- 0029_inventory_batches.up.sql (PostgreSQL)
-- Module: Inventory
-- Inventory is tracked in batches (arrival-date based lotting; clearance
-- after arrival_date + 25 days) with an append-only movement ledger.
-- Stock is never stored on the product row.

CREATE TABLE inventory_batches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID           NOT NULL,
    product_id      UUID           NOT NULL,
    warehouse_id    UUID           NOT NULL,
    supplier_id     UUID,
    purchase_order_item_id UUID,
    lot             VARCHAR(100),
    batch_code      VARCHAR(50),
    arrival_date    TIMESTAMPTZ    NOT NULL,
    expiry_date     TIMESTAMPTZ,
    quantity        NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity >= 0),
    original_quantity NUMERIC(18,4) NOT NULL DEFAULT 0
        CHECK (original_quantity >= 0),
    unit_cost       NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_cost >= 0),
    currency_code   VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate   NUMERIC(18,6)  NOT NULL DEFAULT 1,
    status          VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'depleted', 'written_off')),
    clearance_date  TIMESTAMPTZ,
    is_clearance    BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

    CONSTRAINT fk_inventory_batches_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_inventory_batches_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_batches_warehouse
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_batches_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_inventory_batches_poi
        FOREIGN KEY (purchase_order_item_id) REFERENCES purchase_order_items(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_inventory_batches_product
    ON inventory_batches (product_id, warehouse_id, status);

CREATE INDEX idx_inventory_batches_clearance
    ON inventory_batches (company_id, is_clearance, status)
    WHERE is_clearance = TRUE AND status = 'active';

CREATE TABLE inventory_movements (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    batch_id         UUID           NOT NULL,
    product_id       UUID           NOT NULL,
    warehouse_id     UUID           NOT NULL,
    movement_date    TIMESTAMPTZ    NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out')),
    reference_type   VARCHAR(30),
    reference_id     UUID,
    quantity_delta   NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity_delta <> 0),
    balance_after    NUMERIC(18,4)  NOT NULL DEFAULT 0,
    unit_cost        NUMERIC(18,2)  NOT NULL DEFAULT 0,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    notes            TEXT,
    created_by       UUID,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_inventory_movements_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_inventory_movements_batch
        FOREIGN KEY (batch_id) REFERENCES inventory_batches(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_inventory_movements_warehouse
        FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_inventory_movements_batch
    ON inventory_movements (batch_id, movement_date);

CREATE INDEX idx_inventory_movements_product
    ON inventory_movements (product_id, movement_date);

CREATE INDEX idx_inventory_movements_reference
    ON inventory_movements (reference_type, reference_id);

CREATE TRIGGER trg_inventory_batches_set_updated_at
    BEFORE UPDATE ON inventory_batches
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
