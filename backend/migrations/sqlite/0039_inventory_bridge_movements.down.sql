-- 0039_inventory_bridge_movements.down.sql (SQLite)
-- Reverts to the pre-bridge movement type CHECK by rebuilding the
-- movements table without purchase_receipt / void_sale / void_purchase.

COMMIT;
PRAGMA foreign_keys = OFF;
BEGIN;

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
