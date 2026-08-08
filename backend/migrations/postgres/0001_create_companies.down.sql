-- 0001_create_companies.down.sql
-- Reverse 0001: drop trigger, then the table (drops indexes and constraints
-- automatically).

DROP TRIGGER IF EXISTS trg_companies_set_updated_at ON companies;
DROP TABLE IF EXISTS companies;
