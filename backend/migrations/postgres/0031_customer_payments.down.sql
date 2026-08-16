-- 0031_customer_payments.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_customer_advances_set_updated_at ON customer_advances;
DROP TRIGGER IF EXISTS trg_customer_payments_set_updated_at ON customer_payments;
DROP TABLE IF EXISTS customer_advance_applications;
DROP TABLE IF EXISTS customer_advances;
DROP TABLE IF EXISTS customer_payment_allocations;
DROP TABLE IF EXISTS customer_payments;
