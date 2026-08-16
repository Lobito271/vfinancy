-- 0037_normalize_uuids.up.sql (SQLite)
-- Normalize non-canonical (32-hex, hyphen-less) UUIDs left behind by the
-- DEFAULT (lower(hex(randomblob(16)))) id generators used in the seed
-- migrations (taxes, application_settings, user_roles).
--
-- Why: the Go backend binds uuid.UUID params as canonical hyphenated
-- strings (github.com/google/uuid). Any stored id that lacks hyphens never
-- matches those params, so FK inserts fail, e.g.
--   INSERT INTO products (...) VALUES (..., tax_id) ->
--   FOREIGN KEY constraint failed  (products.tax_id -> taxes.id)
--
-- The rewrite is deterministic and injective:
--   '3d41ac8d0a468e17dd45d24c15cd30cc' -> '3d41ac8d-0a46-8e17-dd45-d24c15cd30cc'
--
-- The runner executes this file inside a single transaction on one
-- connection, so deferring FK enforcement until COMMIT keeps the
-- products.tax_id <-> taxes.id relationship valid at every point.

PRAGMA defer_foreign_keys = ON;

UPDATE taxes
SET id = lower(substr(id, 1, 8) || '-' || substr(id, 9, 4) || '-' || substr(id, 13, 4) || '-' || substr(id, 17, 4) || '-' || substr(id, 21, 12))
WHERE id NOT LIKE '%-%';

UPDATE products
SET tax_id = lower(substr(tax_id, 1, 8) || '-' || substr(tax_id, 9, 4) || '-' || substr(tax_id, 13, 4) || '-' || substr(tax_id, 17, 4) || '-' || substr(tax_id, 21, 12))
WHERE tax_id IS NOT NULL AND tax_id NOT LIKE '%-%';

UPDATE user_roles
SET id = lower(substr(id, 1, 8) || '-' || substr(id, 9, 4) || '-' || substr(id, 13, 4) || '-' || substr(id, 17, 4) || '-' || substr(id, 21, 12))
WHERE id NOT LIKE '%-%';

UPDATE application_settings
SET id = lower(substr(id, 1, 8) || '-' || substr(id, 9, 4) || '-' || substr(id, 13, 4) || '-' || substr(id, 17, 4) || '-' || substr(id, 21, 12))
WHERE id NOT LIKE '%-%';
