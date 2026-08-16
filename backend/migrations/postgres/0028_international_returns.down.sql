-- 0028_international_returns.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_international_returns_set_updated_at ON international_returns;
DROP TABLE IF EXISTS international_return_items;
DROP TABLE IF EXISTS international_returns;
