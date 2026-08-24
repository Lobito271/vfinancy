-- 0044_force_seed_cards_and_warehouse.up.sql (PostgreSQL)
-- Force-seeds the two corporate credit cards and the default warehouse
-- for every active company. Pure additive semantics: rows are inserted
-- only when their natural key is missing; nothing is updated or
-- removed. Only references columns the application already reads and
-- writes on every query.

INSERT INTO chart_of_accounts (id, company_id, code, name, type, depth, allows_movement)
SELECT gen_random_uuid(), c.id, '165.01',
       'Tarjetas de crédito por pagar', 'asset', 2, TRUE
  FROM companies c
 WHERE c.code <> 'SETUP'
   AND NOT EXISTS (
       SELECT 1 FROM chart_of_accounts g
        WHERE g.company_id = c.id AND g.code = '165.01'
   );

INSERT INTO credit_cards
    (id, company_id, issuer, last_four, card_holder,
     expiration_month, expiration_year, credit_limit, current_balance,
     cut_off_day, payment_due_day, currency_code, gl_account_id)
SELECT gen_random_uuid(), c.id, s.issuer, s.last_four,
       'Titular corporativo', 12, 2050, '0.00', '0.00',
       s.cut_off_day, s.payment_due_day, 'USD', g.id
  FROM companies c
  CROSS JOIN (
      SELECT 'visa' AS issuer, '0001' AS last_four, 25 AS cut_off_day, 20 AS payment_due_day
      UNION ALL
      SELECT 'diners', '0002', 3, 20
  ) s
  JOIN chart_of_accounts g
    ON g.company_id = c.id AND g.code = '165.01'
 WHERE c.code <> 'SETUP'
   AND NOT EXISTS (
       SELECT 1 FROM credit_cards cc
        WHERE cc.company_id = c.id
          AND cc.issuer = s.issuer
          AND cc.last_four = s.last_four
   );

INSERT INTO branches (id, company_id, code, name, is_default)
SELECT gen_random_uuid(), c.id, 'SUC-01', 'Sucursal Principal', TRUE
  FROM companies c
 WHERE c.code <> 'SETUP'
   AND NOT EXISTS (
       SELECT 1 FROM branches b WHERE b.company_id = c.id
   );

INSERT INTO warehouses
    (id, company_id, branch_id, code, name, is_default)
SELECT gen_random_uuid(), c.id,
       COALESCE(
           (SELECT b.id FROM branches b
             WHERE b.company_id = c.id AND b.is_default = TRUE
             ORDER BY b.created_at LIMIT 1),
           (SELECT b.id FROM branches b
             WHERE b.company_id = c.id
             ORDER BY b.created_at LIMIT 1)
       ),
       'ALM-01', 'Almacén Principal', TRUE
  FROM companies c
 WHERE c.code <> 'SETUP'
   AND EXISTS (
       SELECT 1 FROM branches b WHERE b.company_id = c.id
   )
   AND NOT EXISTS (
       SELECT 1 FROM warehouses w WHERE w.company_id = c.id
   );

UPDATE warehouses
   SET is_default = TRUE
 WHERE id = (
       SELECT w2.id FROM warehouses w2
        WHERE w2.company_id = warehouses.company_id
        ORDER BY w2.created_at
        LIMIT 1
   )
   AND NOT EXISTS (
       SELECT 1 FROM warehouses w3
        WHERE w3.company_id = warehouses.company_id
          AND w3.is_default = TRUE
   );
