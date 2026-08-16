-- 0023_products.up.sql (PostgreSQL)
-- Module: Products / Master data
-- Catalog master: categories (hierarchical), brands, units of measure,
-- warehouses, and the products themselves. Stock is never stored on the
-- product; it is derived from inventory_movements (0029).

CREATE TABLE product_categories (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    parent_id    UUID,
    path         VARCHAR(100),
    depth        INTEGER        NOT NULL DEFAULT 0
        CHECK (depth >= 0),
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    created_by   UUID,
    updated_by   UUID,

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
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID           NOT NULL,
    code         VARCHAR(20)    NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    created_by   UUID,
    updated_by   UUID,

    CONSTRAINT fk_product_brands_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_product_brands_code
    ON product_brands (company_id, code)
    WHERE deleted_at IS NULL;

CREATE TABLE units_of_measure (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    code             VARCHAR(20)    NOT NULL,
    name             VARCHAR(100)   NOT NULL,
    symbol           VARCHAR(20),
    allows_decimals  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_units_of_measure_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_units_of_measure_code
    ON units_of_measure (company_id, code);

CREATE TABLE warehouses (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID           NOT NULL,
    branch_id         UUID           NOT NULL,
    code              VARCHAR(20)    NOT NULL,
    name              VARCHAR(200)   NOT NULL,
    address           TEXT,
    manager_id        UUID,
    is_default        BOOLEAN        NOT NULL DEFAULT FALSE,
    allows_clearance  BOOLEAN        NOT NULL DEFAULT FALSE,
    is_active         BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        UUID,
    updated_by        UUID,

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
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID           NOT NULL,
    sku            VARCHAR(50)    NOT NULL,
    barcode        VARCHAR(50),
    description    TEXT           NOT NULL,
    category_id    UUID,
    brand_id       UUID,
    unit_id        UUID           NOT NULL,
    tax_id         UUID           NOT NULL,
    cost           NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (cost >= 0),
    sale_price     NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (sale_price >= 0),
    sale_currency  VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    min_stock      NUMERIC(18,4)  NOT NULL DEFAULT 0,
    max_stock      NUMERIC(18,4)  NOT NULL DEFAULT 0,
    weight         NUMERIC(18,4)  NOT NULL DEFAULT 0,
    is_active      BOOLEAN        NOT NULL DEFAULT TRUE,
    is_service     BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,
    created_by     UUID,
    updated_by     UUID,

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

CREATE TRIGGER trg_product_categories_set_updated_at
    BEFORE UPDATE ON product_categories
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_product_brands_set_updated_at
    BEFORE UPDATE ON product_brands
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_warehouses_set_updated_at
    BEFORE UPDATE ON warehouses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_products_set_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
