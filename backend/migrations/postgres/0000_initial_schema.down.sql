

ALTER TABLE inventory_movements
    DROP CONSTRAINT inventory_movements_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_type_check
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out', 'void_out'));


ALTER TABLE inventory_batches
    DROP CONSTRAINT inventory_batches_status_check;

ALTER TABLE inventory_batches
    ADD CONSTRAINT inventory_batches_status_check
        CHECK (status IN ('active', 'depleted', 'written_off'));

ALTER TABLE inventory_movements
    DROP CONSTRAINT inventory_movements_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_type_check
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out'));


UPDATE taxes
SET id = id::text::uuid
WHERE id::text LIKE '%-%';

UPDATE products
SET tax_id = tax_id::text::uuid
WHERE tax_id IS NOT NULL AND tax_id::text LIKE '%-%';

UPDATE application_settings
SET id = id::text::uuid
WHERE id::text LIKE '%-%';


DELETE FROM warehouses
WHERE id = '00000000-0000-0000-0000-0000000000c1'
  AND company_id = '00000000-0000-0000-0000-000000000001';


DROP TRIGGER IF EXISTS trg_supplier_payments_set_updated_at ON supplier_payments;
DROP TRIGGER IF EXISTS trg_purchase_orders_set_updated_at ON purchase_orders;
DROP TABLE IF EXISTS supplier_payment_allocations;
DROP TABLE IF EXISTS supplier_payments;
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;


DROP TRIGGER IF EXISTS trg_customer_advances_set_updated_at ON customer_advances;
DROP TRIGGER IF EXISTS trg_customer_payments_set_updated_at ON customer_payments;
DROP TABLE IF EXISTS customer_advance_applications;
DROP TABLE IF EXISTS customer_advances;
DROP TABLE IF EXISTS customer_payment_allocations;
DROP TABLE IF EXISTS customer_payments;


DROP TRIGGER IF EXISTS trg_sales_set_updated_at ON sales;
DROP TABLE IF EXISTS sale_items;
DROP TABLE IF EXISTS sales;


DROP TRIGGER IF EXISTS trg_inventory_batches_set_updated_at ON inventory_batches;
DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory_batches;


DROP TRIGGER IF EXISTS trg_international_returns_set_updated_at ON international_returns;
DROP TABLE IF EXISTS international_return_items;
DROP TABLE IF EXISTS international_returns;


DROP TRIGGER IF EXISTS trg_bank_transactions_set_updated_at ON bank_transactions;
DROP TRIGGER IF EXISTS trg_credit_cards_set_updated_at ON credit_cards;
DROP TRIGGER IF EXISTS trg_bank_accounts_set_updated_at ON bank_accounts;
DROP TABLE IF EXISTS bank_transactions;
DROP TABLE IF EXISTS credit_cards;
DROP TABLE IF EXISTS bank_accounts;


DROP TRIGGER IF EXISTS trg_journal_entries_set_updated_at ON journal_entries;
DROP TABLE IF EXISTS journal_entry_lines;
DROP TABLE IF EXISTS journal_entries;


DROP TRIGGER IF EXISTS trg_chart_of_accounts_set_updated_at ON chart_of_accounts;
DROP TABLE IF EXISTS chart_of_accounts;


DROP TRIGGER IF EXISTS trg_fiscal_periods_set_updated_at ON fiscal_periods;
DROP TABLE IF EXISTS fiscal_periods;


DROP TRIGGER IF EXISTS trg_products_set_updated_at ON products;
DROP TRIGGER IF EXISTS trg_warehouses_set_updated_at ON warehouses;
DROP TRIGGER IF EXISTS trg_product_brands_set_updated_at ON product_brands;
DROP TRIGGER IF EXISTS trg_product_categories_set_updated_at ON product_categories;

DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS units_of_measure;
DROP TABLE IF EXISTS product_brands;
DROP TABLE IF EXISTS product_categories;


DROP TRIGGER IF EXISTS trg_suppliers_set_updated_at ON suppliers;
DROP TABLE IF EXISTS suppliers;


DROP TRIGGER IF EXISTS trg_customers_set_updated_at ON customers;
DROP TABLE IF EXISTS customers;


DROP TABLE IF EXISTS sync_tombstones;
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS sync_cursors;
DROP TABLE IF EXISTS sync_devices;


DROP TRIGGER IF EXISTS trg_audit_events_no_mutation ON audit_events;
DROP FUNCTION IF EXISTS reject_audit_events_mutation();
DROP TABLE IF EXISTS audit_events;


DROP TABLE IF EXISTS exchange_rates;


DROP TRIGGER IF EXISTS trg_taxes_set_updated_at ON taxes;
DROP TABLE IF EXISTS taxes;


DROP TABLE IF EXISTS countries;


DROP TRIGGER IF EXISTS trg_currencies_set_updated_at ON currencies;
DROP TABLE IF EXISTS currencies;


DROP TRIGGER IF EXISTS trg_settings_set_updated_at ON application_settings;
DROP TABLE IF EXISTS application_settings;


DROP TRIGGER IF EXISTS trg_audit_logs_no_update ON audit_logs;
DROP TRIGGER IF EXISTS trg_audit_logs_no_delete ON audit_logs;
DROP FUNCTION IF EXISTS audit_logs_forbid_mutation();
DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS local_profiles;


DROP TRIGGER IF EXISTS trg_branches_set_updated_at ON branches;
DROP TABLE IF EXISTS branches;


DROP TRIGGER IF EXISTS trg_companies_set_updated_at ON companies;
DROP TABLE IF EXISTS companies;


DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS schema_migrations;
