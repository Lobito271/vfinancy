-- 0021_customers.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_customers_set_updated_at ON customers;
DROP TABLE IF EXISTS customers;
