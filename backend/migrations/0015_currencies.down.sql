-- 0015_currencies.down.sql

DROP TRIGGER IF EXISTS trg_currencies_set_updated_at ON currencies;
DROP TABLE IF EXISTS currencies;
