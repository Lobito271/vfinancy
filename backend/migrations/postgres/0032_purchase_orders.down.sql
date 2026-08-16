-- 0032_purchase_orders.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_supplier_payments_set_updated_at ON supplier_payments;
DROP TRIGGER IF EXISTS trg_purchase_orders_set_updated_at ON purchase_orders;
DROP TABLE IF EXISTS supplier_payment_allocations;
DROP TABLE IF EXISTS supplier_payments;
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;
