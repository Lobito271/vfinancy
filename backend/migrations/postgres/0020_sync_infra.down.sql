-- 0020_sync_infra.down.sql (PostgreSQL)

DROP TABLE IF EXISTS sync_tombstones;
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS sync_cursors;
DROP TABLE IF EXISTS sync_devices;
