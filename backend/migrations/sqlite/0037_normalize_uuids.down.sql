-- 0037_normalize_uuids.down.sql (SQLite)
-- Reverse 0037: restore hyphen-less 32-hex ids.

PRAGMA defer_foreign_keys = ON;

UPDATE taxes
SET id = replace(id, '-', '')
WHERE id LIKE '%-%';

UPDATE products
SET tax_id = replace(tax_id, '-', '')
WHERE tax_id IS NOT NULL AND tax_id LIKE '%-%';

UPDATE user_roles
SET id = replace(id, '-', '')
WHERE id LIKE '%-%';

UPDATE application_settings
SET id = replace(id, '-', '')
WHERE id LIKE '%-%';
