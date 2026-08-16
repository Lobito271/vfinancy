-- 0023_products.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_products_set_updated_at ON products;
DROP TRIGGER IF EXISTS trg_warehouses_set_updated_at ON warehouses;
DROP TRIGGER IF EXISTS trg_product_brands_set_updated_at ON product_brands;
DROP TRIGGER IF EXISTS trg_product_categories_set_updated_at ON product_categories;

DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS units_of_measure;
DROP TABLE IF EXISTS product_brands;
DROP TABLE IF EXISTS product_categories;
