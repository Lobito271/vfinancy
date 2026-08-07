-- 0002_create_branches.down.sql

DROP TRIGGER IF EXISTS trg_branches_set_updated_at ON branches;
DROP TABLE IF EXISTS branches;
