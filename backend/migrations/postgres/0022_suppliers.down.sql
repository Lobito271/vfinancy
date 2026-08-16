-- 0022_suppliers.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_suppliers_set_updated_at ON suppliers;
DROP TABLE IF EXISTS suppliers;
