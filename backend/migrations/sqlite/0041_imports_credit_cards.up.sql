-- 0041_imports_credit_cards.up.sql (SQLite)
-- Module: Purchasing / Imports ERP

ALTER TABLE purchase_orders ADD COLUMN credit_card_id TEXT REFERENCES credit_cards(id);