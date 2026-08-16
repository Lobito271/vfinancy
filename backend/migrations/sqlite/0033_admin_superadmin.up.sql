-- 0033_admin_superadmin.up.sql (SQLite)
-- Superadministrador:
--   1. Adds the "*.*" wildcard permission to the catalog.
--   2. Re-syncs the system "admin" role with the full permission catalog
--      (covers pre-existing databases where the seed never ran or newer
--      permission codes were added later).
--   3. Ensures the seeded demo user "victorfinancy" holds the admin role,
--      so existing installs get full CRUD access too.
--   4. Seeds the default units of measure required to create products.

INSERT INTO permissions (code, module, action, description)
SELECT '*.*', 'system', 'superadmin', 'Superadministrador: acceso total al sistema'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = '*.*');

INSERT OR IGNORE INTO role_permissions (role_id, permission_code)
SELECT r.id, p.code
FROM roles r, permissions p
WHERE r.company_id = '00000000-0000-0000-0000-000000000001'
  AND r.code = 'admin'
  AND r.deleted_at IS NULL;

INSERT OR IGNORE INTO user_roles (user_id, role_id, branch_id)
SELECT u.id, r.id, NULL
FROM users u
JOIN roles r ON r.company_id = u.company_id
WHERE u.company_id = '00000000-0000-0000-0000-000000000001'
  AND u.username = 'victorfinancy'
  AND u.deleted_at IS NULL
  AND r.code = 'admin'
  AND r.deleted_at IS NULL;

INSERT INTO units_of_measure (id, company_id, code, name, symbol, allows_decimals)
SELECT '00000000-0000-0000-0000-0000000000b1', '00000000-0000-0000-0000-000000000001', 'UND', 'Unidad', 'und', TRUE
WHERE NOT EXISTS (SELECT 1 FROM units_of_measure WHERE company_id = '00000000-0000-0000-0000-000000000001' AND code = 'UND');

INSERT INTO units_of_measure (id, company_id, code, name, symbol, allows_decimals)
SELECT '00000000-0000-0000-0000-0000000000b2', '00000000-0000-0000-0000-000000000001', 'KG', 'Kilogramo', 'kg', TRUE
WHERE NOT EXISTS (SELECT 1 FROM units_of_measure WHERE company_id = '00000000-0000-0000-0000-000000000001' AND code = 'KG');

INSERT INTO units_of_measure (id, company_id, code, name, symbol, allows_decimals)
SELECT '00000000-0000-0000-0000-0000000000b3', '00000000-0000-0000-0000-000000000001', 'L', 'Litro', 'l', TRUE
WHERE NOT EXISTS (SELECT 1 FROM units_of_measure WHERE company_id = '00000000-0000-0000-0000-000000000001' AND code = 'L');

INSERT INTO units_of_measure (id, company_id, code, name, symbol, allows_decimals)
SELECT '00000000-0000-0000-0000-0000000000b4', '00000000-0000-0000-0000-000000000001', 'CAJA', 'Caja', 'caja', FALSE
WHERE NOT EXISTS (SELECT 1 FROM units_of_measure WHERE company_id = '00000000-0000-0000-0000-000000000001' AND code = 'CAJA');
