-- 0040_imports_orders.down.sql (SQLite)

DROP INDEX IF EXISTS idx_customer_order_payments_customer;
DROP INDEX IF EXISTS idx_customer_order_payments_order;
DROP INDEX IF EXISTS uq_customer_order_payments_number;
DROP TABLE IF EXISTS customer_order_payments;

DROP INDEX IF EXISTS idx_purchase_orders_customer;
DROP INDEX IF EXISTS idx_purchase_orders_order_type;

ALTER TABLE products DROP COLUMN details;

ALTER TABLE purchase_orders DROP COLUMN refunded_amount;
ALTER TABLE purchase_orders DROP COLUMN faulty_reason;
ALTER TABLE purchase_orders DROP COLUMN faulty;
ALTER TABLE purchase_orders DROP COLUMN anticipo_date;
ALTER TABLE purchase_orders DROP COLUMN anticipo;
ALTER TABLE purchase_orders DROP COLUMN projected_profit_pen;
ALTER TABLE purchase_orders DROP COLUMN real_cost_pen;
ALTER TABLE purchase_orders DROP COLUMN sale_price_pen;
ALTER TABLE purchase_orders DROP COLUMN cost_usd;
ALTER TABLE purchase_orders DROP COLUMN arrival_date;
ALTER TABLE purchase_orders DROP COLUMN customer_id;
ALTER TABLE purchase_orders DROP COLUMN order_type;
