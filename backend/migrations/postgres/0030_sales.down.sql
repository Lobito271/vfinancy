-- 0030_sales.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_sales_set_updated_at ON sales;
DROP TABLE IF EXISTS sale_items;
DROP TABLE IF EXISTS sales;
