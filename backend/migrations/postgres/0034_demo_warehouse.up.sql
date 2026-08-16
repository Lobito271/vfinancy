-- 0034_demo_warehouse.up.sql (PostgreSQL)
-- Seed a default warehouse for the demo company so stock receipt
-- (ReceiveStock) has a valid warehouse to reference. Idempotent.

INSERT INTO warehouses (id, company_id, branch_id, code, name, address, is_default, allows_clearance, is_active)
SELECT '00000000-0000-0000-0000-0000000000c1', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'ALM-01', 'Almacén Principal', NULL, TRUE, FALSE, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM warehouses
    WHERE company_id = '00000000-0000-0000-0000-000000000001'
      AND code = 'ALM-01'
      AND deleted_at IS NULL
);
