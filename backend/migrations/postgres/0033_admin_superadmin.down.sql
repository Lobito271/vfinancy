-- 0033_admin_superadmin.down.sql (PostgreSQL)

DELETE FROM role_permissions WHERE permission_code = '*.*';
DELETE FROM permissions WHERE code = '*.*';

DELETE FROM units_of_measure
WHERE company_id = '00000000-0000-0000-0000-000000000001'
  AND code IN ('UND', 'KG', 'L', 'CAJA');
