-- 0037_normalize_uuids.up.sql (PostgreSQL)
-- Mirror of the SQLite migration. The Postgres seed migrations use
-- gen_random_uuid() defaults, so every id is already a canonical UUID;
-- these statements are no-ops kept to keep the pair in sync.

UPDATE taxes
SET id = id::text::uuid
WHERE id::text NOT LIKE '%-%';

UPDATE products
SET tax_id = tax_id::text::uuid
WHERE tax_id IS NOT NULL AND tax_id::text NOT LIKE '%-%';

UPDATE user_roles
SET id = id::text::uuid
WHERE id::text NOT LIKE '%-%';

UPDATE application_settings
SET id = id::text::uuid
WHERE id::text NOT LIKE '%-%';
