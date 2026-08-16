-- 0023_products.up.sql (SQLite)
-- Module: Products / Master data
-- Catalog master: categories (hierarchical), brands, units of measure,
-- warehouses, and the products themselves. Stock is never stored on the
-- product; it is derived from inventory_movements (0029).

CREATE TABLE product_categories (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    parent_id    TEXT,
    path         VARCHAR(100),
    depth        INTEGER        NOT NULL DEFAULT 0
        CHECK (depth >= 0),
    created_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at   TIMESTAMP,
    created_by   TEXT,
    updated_by   TEXT,

    CONSTRAINT fk_product_categories_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_product_categories_parent
        FOREIGN KEY (parent_id) REFERENCES product_categories(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_product_categories_code
    ON product_categories (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE product_brands (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    created_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at   TIMESTAMP,
    created_by   TEXT,
    updated_by   TEXT,

    CONSTRAINT fk_product_brands_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_product_brands_code
    ON product_brands (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE units_of_measure (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    code             VARCHAR(20)    NOT NULL,
    name             VARCHAR(100)   NOT NULL,
    symbol           VARCHAR(20),
    allows_decimals  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_units_of_measure_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_units_of_measure_code
    ON units_of_measure (company_id, code);

CREATE TABLE warehouses (
    id                TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id        TEXT           NOT NULL,
    branch_id         TEXT           NOT NULL,
    code              VARCHAR(20)    NOT NULL,
    name              VARCHAR(200)   NOT NULL,
    address           TEXT,
    manager_id        TEXT,
    is_default        BOOLEAN        NOT NULL DEFAULT FALSE,
    allows_clearance  BOOLEAN        NOT NULL DEFAULT FALSE,
    is_active         BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at        TIMESTAMP,
    created_by        TEXT,
    updated_by        TEXT,

    CONSTRAINT fk_warehouses_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_warehouses_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_warehouses_manager
        FOREIGN KEY (manager_id) REFERENCES users(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_warehouses_code
    ON warehouses (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE products (
    id             TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id     TEXT           NOT NULL,
    sku            VARCHAR(50)    NOT NULL,
    barcode        VARCHAR(50),
    description    TEXT           NOT NULL,
    category_id    TEXT,
    brand_id       TEXT,
    unit_id        TEXT           NOT NULL,
    tax_id         TEXT           NOT NULL,
    cost           TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(cost AS REAL) >= 0),
    sale_price     TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(sale_price AS REAL) >= 0),
    sale_currency  VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    min_stock      TEXT           NOT NULL DEFAULT '0.0000',
    max_stock      TEXT           NOT NULL DEFAULT '0.0000',
    weight         TEXT           NOT NULL DEFAULT '0.0000',
    is_active      BOOLEAN        NOT NULL DEFAULT TRUE,
    is_service     BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at     TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    deleted_at     TIMESTAMP,
    created_by     TEXT,
    updated_by     TEXT,

    CONSTRAINT fk_products_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id) REFERENCES product_categories(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_products_brand
        FOREIGN KEY (brand_id) REFERENCES product_brands(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_products_unit
        FOREIGN KEY (unit_id) REFERENCES units_of_measure(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_products_tax
        FOREIGN KEY (tax_id) REFERENCES taxes(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_products_description_nonblank
        CHECK (length(trim(description)) > 0)
);

CREATE UNIQUE INDEX uq_products_sku
    ON products (company_id, sku)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_products_barcode
    ON products (company_id, barcode)
    WHERE deleted_at IS NULL AND barcode IS NOT NULL;

CREATE INDEX idx_products_category
    ON products (company_id, category_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_products_brand
    ON products (company_id, brand_id)
    WHERE deleted_at IS NULL;
