-- SQLite does not support ALTER TABLE DROP COLUMN; recreate without the column.
CREATE TABLE purchase_orders_new (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL,
  supplier_id TEXT NOT NULL,
  branch_id TEXT,
  number TEXT NOT NULL,
  order_date INTEGER NOT NULL,
  expected_date INTEGER,
  received_date INTEGER,
  status TEXT NOT NULL DEFAULT 'pending',
  subtotal TEXT NOT NULL DEFAULT '0.00',
  tax_amount TEXT NOT NULL DEFAULT '0.00',
  total TEXT NOT NULL DEFAULT '0.00',
  paid_amount TEXT NOT NULL DEFAULT '0.00',
  discount_amount TEXT NOT NULL DEFAULT '0.00',
  currency_code TEXT NOT NULL DEFAULT 'PEN',
  exchange_rate TEXT NOT NULL DEFAULT '1.000000',
  notes TEXT,
  order_type TEXT NOT NULL DEFAULT 'general',
  customer_id TEXT,
  credit_card_id TEXT,
  arrival_date INTEGER,
  cost_usd TEXT NOT NULL DEFAULT '0.00',
  sale_price_pen TEXT NOT NULL DEFAULT '0.00',
  real_cost_pen TEXT NOT NULL DEFAULT '0.00',
  projected_profit_pen TEXT NOT NULL DEFAULT '0.00',
  anticipo TEXT NOT NULL DEFAULT '0.00',
  anticipo_date INTEGER,
  faulty INTEGER NOT NULL DEFAULT 0,
  faulty_reason TEXT,
  refunded_amount TEXT NOT NULL DEFAULT '0.00',
  cancelled_at INTEGER,
  cancelled_reason TEXT,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  deleted_at INTEGER,
  created_by TEXT,
  updated_by TEXT
);

INSERT INTO purchase_orders_new (
  id, company_id, supplier_id, branch_id, number, order_date,
  expected_date, received_date, status, subtotal, tax_amount, total,
  paid_amount, discount_amount, currency_code, exchange_rate, notes,
  order_type, customer_id, credit_card_id, arrival_date, cost_usd, sale_price_pen,
  real_cost_pen, projected_profit_pen, anticipo, anticipo_date,
  faulty, faulty_reason, refunded_amount,
  cancelled_at, cancelled_reason, created_at, updated_at, deleted_at,
  created_by, updated_by
) SELECT
  id, company_id, supplier_id, branch_id, number, order_date,
  expected_date, received_date, status, subtotal, tax_amount, total,
  paid_amount, discount_amount, currency_code, exchange_rate, notes,
  order_type, customer_id, credit_card_id, arrival_date, cost_usd, sale_price_pen,
  real_cost_pen, projected_profit_pen, anticipo, anticipo_date,
  faulty, faulty_reason, refunded_amount,
  cancelled_at, cancelled_reason, created_at, updated_at, deleted_at,
  created_by, updated_by
FROM purchase_orders;

DROP TABLE purchase_orders;
ALTER TABLE purchase_orders_new RENAME TO purchase_orders;
