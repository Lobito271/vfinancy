-- 0017_taxes.down.sql

DROP TRIGGER IF EXISTS trg_taxes_set_updated_at ON taxes;
DROP TABLE IF EXISTS taxes;
