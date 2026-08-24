-- 0040_imports_orders.down.sql (PostgreSQL)

DROP TABLE IF EXISTS customer_order_payments;

DROP INDEX IF EXISTS idx_purchase_orders_customer;
DROP INDEX IF EXISTS idx_purchase_orders_order_type;

ALTER TABLE products DROP COLUMN details;

ALTER TABLE purchase_orders
    DROP CONSTRAINT IF EXISTS fk_purchase_orders_customer,
    DROP COLUMN IF EXISTS order_type,
    DROP COLUMN IF EXISTS customer_id,
    DROP COLUMN IF EXISTS arrival_date,
    DROP COLUMN IF EXISTS cost_usd,
    DROP COLUMN IF EXISTS sale_price_pen,
    DROP COLUMN IF EXISTS real_cost_pen,
    DROP COLUMN IF EXISTS projected_profit_pen,
    DROP COLUMN IF EXISTS anticipo,
    DROP COLUMN IF EXISTS anticipo_date,
    DROP COLUMN IF EXISTS faulty,
    DROP COLUMN IF EXISTS faulty_reason,
    DROP COLUMN IF EXISTS refunded_amount;
