-- 0037_normalize_uuids.down.sql (PostgreSQL)
-- Mirror of the SQLite down migration. Postgres ids are always canonical
-- UUIDs, so nothing to restore; kept as a no-op to keep the pair in sync.

UPDATE taxes
SET id = id::text::uuid
WHERE id::text LIKE '%-%';

UPDATE products
SET tax_id = tax_id::text::uuid
WHERE tax_id IS NOT NULL AND tax_id::text LIKE '%-%';

UPDATE user_roles
SET id = id::text::uuid
WHERE id::text LIKE '%-%';

UPDATE application_settings
SET id = id::text::uuid
WHERE id::text LIKE '%-%';
