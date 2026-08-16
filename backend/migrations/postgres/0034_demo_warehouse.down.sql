-- 0034_demo_warehouse.down.sql (PostgreSQL)

DELETE FROM warehouses
WHERE id = '00000000-0000-0000-0000-0000000000c1'
  AND company_id = '00000000-0000-0000-0000-000000000001';
