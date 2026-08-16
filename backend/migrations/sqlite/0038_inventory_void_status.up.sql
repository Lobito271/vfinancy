-- 0038_inventory_void_status.up.sql (SQLite)
-- Module: Inventory
-- Adds the "voided" batch status (mistake correction: quantity zeroed,
-- row kept for audit) and the "void_out" movement type that records it.
--
-- SQLite cannot ALTER a CHECK constraint, so both tables are rebuilt.
-- Foreign keys must be off during the rebuild; the migration runner
-- wraps each file in a transaction (where PRAGMA foreign_keys is a
-- no-op), so we end that transaction first and reopen it at the end.

COMMIT;
PRAGMA foreign_keys = OFF;
BEGIN;

CREATE TABLE inventory_batches_new (
    id              TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id      TEXT           NOT NULL,
    product_id      TEXT           NOT NULL,
    warehouse_id    TEXT           NOT NULL,
    supplier_id     TEXT,
    purchase_order_item_id TEXT,
    lot             VARCHAR(100),
    batch_code      VARCHAR(50),
    arrival_date    TIMESTAMP      NOT NULL,
    expiry_date     TIMESTAMP,
    quantity        TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity AS REAL) >= 0),
    original_quantity TEXT         NOT NULL DEFAULT '0.0000'
        CHECK (CAST(original_quantity AS REAL) >= 0),
    unit_cost       TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(unit_cost AS REAL) >= 0),
    currency_code   VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate   TEXT           NOT NULL DEFAULT '1.000000',
    status          VARCHAR(20)    NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'depleted', 'written_off', 'voided')),
    clearance_date  TIMESTAMP,
    is_clearance    BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by      TEXT,
    updated_by      TEXT,

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

INSERT INTO inventory_batches_new SELECT * FROM inventory_batches;
DROP TABLE inventory_batches;
ALTER TABLE inventory_batches_new RENAME TO inventory_batches;

CREATE INDEX idx_inventory_batches_product
    ON inventory_batches (product_id, warehouse_id, status);

CREATE INDEX idx_inventory_batches_clearance
    ON inventory_batches (company_id, is_clearance, status)
    WHERE is_clearance = TRUE AND status = 'active';

CREATE TABLE inventory_movements_new (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    batch_id         TEXT           NOT NULL,
    product_id       TEXT           NOT NULL,
    warehouse_id     TEXT           NOT NULL,
    movement_date    TIMESTAMP      NOT NULL,
    type             VARCHAR(20)    NOT NULL
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out', 'void_out')),
    reference_type   VARCHAR(30),
    reference_id     TEXT,
    quantity_delta   TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity_delta AS REAL) <> 0),
    balance_after    TEXT           NOT NULL DEFAULT '0.0000',
    unit_cost        TEXT           NOT NULL DEFAULT '0.00',
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    notes            TEXT,
    created_by       TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

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

INSERT INTO inventory_movements_new SELECT * FROM inventory_movements;
DROP TABLE inventory_movements;
ALTER TABLE inventory_movements_new RENAME TO inventory_movements;

CREATE INDEX idx_inventory_movements_batch
    ON inventory_movements (batch_id, movement_date);

CREATE INDEX idx_inventory_movements_product
    ON inventory_movements (product_id, movement_date);

CREATE INDEX idx_inventory_movements_reference
    ON inventory_movements (reference_type, reference_id);

COMMIT;
PRAGMA foreign_keys = ON;
BEGIN;
