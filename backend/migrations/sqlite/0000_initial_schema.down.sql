

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
        CHECK (status IN ('active', 'depleted', 'written_off')),
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
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out')),
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


PRAGMA defer_foreign_keys = ON;

UPDATE taxes
SET id = replace(id, '-', '')
WHERE id LIKE '%-%';

UPDATE products
SET tax_id = replace(tax_id, '-', '')
WHERE tax_id IS NOT NULL AND tax_id LIKE '%-%';

UPDATE application_settings
SET id = replace(id, '-', '')
WHERE id LIKE '%-%';




DELETE FROM warehouses
WHERE id = '00000000-0000-0000-0000-0000000000c1'
  AND company_id = '00000000-0000-0000-0000-000000000001';


DROP TABLE IF EXISTS supplier_payment_allocations;
DROP TABLE IF EXISTS supplier_payments;
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;


DROP TABLE IF EXISTS customer_advance_applications;
DROP TABLE IF EXISTS customer_advances;
DROP TABLE IF EXISTS customer_payment_allocations;
DROP TABLE IF EXISTS customer_payments;


DROP TABLE IF EXISTS sale_items;
DROP TABLE IF EXISTS sales;


DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory_batches;


DROP TABLE IF EXISTS international_return_items;
DROP TABLE IF EXISTS international_returns;


DROP TABLE IF EXISTS bank_transactions;
DROP TABLE IF EXISTS credit_cards;
DROP TABLE IF EXISTS bank_accounts;


DROP TABLE IF EXISTS journal_entry_lines;
DROP TABLE IF EXISTS journal_entries;


DROP TABLE IF EXISTS chart_of_accounts;


DROP TABLE IF EXISTS fiscal_periods;


DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS units_of_measure;
DROP TABLE IF EXISTS product_brands;
DROP TABLE IF EXISTS product_categories;


DROP TABLE IF EXISTS suppliers;


DROP TABLE IF EXISTS customers;


DROP TRIGGER IF EXISTS trg_countries_sync_delete;
DROP TRIGGER IF EXISTS trg_currencies_sync_delete;
DROP TRIGGER IF EXISTS trg_taxes_sync_delete;
DROP TRIGGER IF EXISTS trg_application_settings_sync_delete;
DROP TRIGGER IF EXISTS trg_branches_sync_delete;
DROP TRIGGER IF EXISTS trg_companies_sync_delete;

DROP TABLE IF EXISTS sync_tombstones;
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS sync_cursors;
DROP TABLE IF EXISTS sync_devices;


DROP TRIGGER IF EXISTS trg_audit_events_no_update;
DROP TRIGGER IF EXISTS trg_audit_events_no_delete;
DROP TABLE IF EXISTS audit_events;


DROP TABLE IF EXISTS exchange_rates;


DROP TABLE IF EXISTS taxes;


DROP TABLE IF EXISTS countries;


DROP TABLE IF EXISTS currencies;


DROP TABLE IF EXISTS application_settings;


DROP TRIGGER IF EXISTS trg_audit_logs_no_update;
DROP TRIGGER IF EXISTS trg_audit_logs_no_delete;
DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS local_profiles;


DROP TABLE IF EXISTS branches;


DROP TABLE IF EXISTS companies;


SELECT 1;
