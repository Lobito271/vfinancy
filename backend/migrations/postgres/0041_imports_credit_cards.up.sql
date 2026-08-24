-- 0041_imports_credit_cards.up.sql (PostgreSQL)
-- Module: Purchasing / Imports ERP
--
-- Credit cards used to pay suppliers (100% upfront in USD):
--   1. purchase_orders.credit_card_id — the card that settled the order.
--   2. Seeds the company's two cards (Visa / Diners) with their billing
--      cycles so the UI can require a card and project card billing:
--        - Visa:   cut-off on the 25th, payment due on the 20th.
--        - Diners: cut-off on the 3rd,  payment due on the 20th.
-- A liability GL account (165.01) is created for the card payables.

ALTER TABLE purchase_orders ADD COLUMN credit_card_id UUID REFERENCES credit_cards(id) ON DELETE SET NULL;

INSERT INTO chart_of_accounts (id, company_id, code, name, type, depth, allows_movement, description)
SELECT '00000000-0000-0000-0000-00000000ca01', '00000000-0000-0000-0000-000000000001',
       '165.01', 'Tarjetas de crédito por pagar', 'liability', 2, TRUE,
       'Cuentas por pagar a tarjetas de crédito corporativas'
WHERE NOT EXISTS (
    SELECT 1 FROM chart_of_accounts
    WHERE company_id = '00000000-0000-0000-0000-000000000001'
      AND code = '165.01'
);

INSERT INTO credit_cards (
    id, company_id, branch_id, issuer, last_four, card_holder,
    expiration_month, expiration_year, credit_limit, current_balance,
    cut_off_day, payment_due_day, currency_code, gl_account_id, is_active
)
SELECT '00000000-0000-0000-0000-00000000ca11',
       '00000000-0000-0000-0000-000000000001',
       NULL,
       'visa', '0001', 'vfinancy S.A.C.',
       12, 2030, 50000.00, 0.00,
       25, 20, 'USD',
       (SELECT id FROM chart_of_accounts
         WHERE company_id = '00000000-0000-0000-0000-000000000001' AND code = '165.01' LIMIT 1),
       TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM credit_cards
    WHERE company_id = '00000000-0000-0000-0000-000000000001'
      AND issuer = 'visa'
      AND deleted_at IS NULL
);

INSERT INTO credit_cards (
    id, company_id, branch_id, issuer, last_four, card_holder,
    expiration_month, expiration_year, credit_limit, current_balance,
    cut_off_day, payment_due_day, currency_code, gl_account_id, is_active
)
SELECT '00000000-0000-0000-0000-00000000ca12',
       '00000000-0000-0000-0000-000000000001',
       NULL,
       'diners', '0002', 'vfinancy S.A.C.',
       12, 2030, 50000.00, 0.00,
       3, 20, 'USD',
       (SELECT id FROM chart_of_accounts
         WHERE company_id = '00000000-0000-0000-0000-000000000001' AND code = '165.01' LIMIT 1),
       TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM credit_cards
    WHERE company_id = '00000000-0000-0000-0000-000000000001'
      AND issuer = 'diners'
      AND deleted_at IS NULL
);
