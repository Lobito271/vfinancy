-- 0024_fiscal_periods.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_fiscal_periods_set_updated_at ON fiscal_periods;
DROP TABLE IF EXISTS fiscal_periods;
