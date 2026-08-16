-- 0027_bank_accounts.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_bank_transactions_set_updated_at ON bank_transactions;
DROP TRIGGER IF EXISTS trg_credit_cards_set_updated_at ON credit_cards;
DROP TRIGGER IF EXISTS trg_bank_accounts_set_updated_at ON bank_accounts;
DROP TABLE IF EXISTS bank_transactions;
DROP TABLE IF EXISTS credit_cards;
DROP TABLE IF EXISTS bank_accounts;
