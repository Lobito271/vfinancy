-- 0041_imports_credit_cards.down.sql (SQLite)

DELETE FROM credit_cards
WHERE company_id = '00000000-0000-0000-0000-000000000001'
  AND issuer IN ('visa', 'diners')
  AND deleted_at IS NULL;

DELETE FROM chart_of_accounts
WHERE company_id = '00000000-0000-0000-0000-000000000001'
  AND code = '165.01';

ALTER TABLE purchase_orders DROP COLUMN credit_card_id;
