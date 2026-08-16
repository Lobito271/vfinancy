-- 0025_chart_of_accounts.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_chart_of_accounts_set_updated_at ON chart_of_accounts;
DROP TABLE IF EXISTS chart_of_accounts;
