-- 0029_inventory_batches.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_inventory_batches_set_updated_at ON inventory_batches;
DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory_batches;
