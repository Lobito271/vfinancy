-- 0000_init.down.sql
-- Drop the trigger function and the migrations bookkeeping table.
-- The pgcrypto extension is NOT dropped because other extensions
-- may depend on it; in a real environment the DBA decides whether
-- to drop it explicitly.

DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS schema_migrations;
