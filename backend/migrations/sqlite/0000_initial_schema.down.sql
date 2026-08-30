DROP TABLE IF EXISTS customer_order_payments;
DROP TABLE IF EXISTS supplier_payment_allocations;
DROP TABLE IF EXISTS supplier_payments;
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;


DROP TABLE IF EXISTS customer_advance_applications;
DROP TABLE IF EXISTS customer_advances;
DROP TABLE IF EXISTS customer_payment_allocations;
DROP TABLE IF EXISTS customer_payments;


DROP TABLE IF EXISTS sale_items;
DROP TABLE IF EXISTS sales;


DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory_batches;


DROP TABLE IF EXISTS international_return_items;
DROP TABLE IF EXISTS international_returns;


DROP TABLE IF EXISTS bank_transactions;
DROP TABLE IF EXISTS credit_cards;
DROP TABLE IF EXISTS bank_accounts;


DROP TABLE IF EXISTS journal_entry_lines;
DROP TABLE IF EXISTS journal_entries;


DROP TABLE IF EXISTS chart_of_accounts;


DROP TABLE IF EXISTS fiscal_periods;


DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS units_of_measure;
DROP TABLE IF EXISTS product_brands;
DROP TABLE IF EXISTS product_categories;


DROP TABLE IF EXISTS suppliers;


DROP TABLE IF EXISTS customers;


DROP TRIGGER IF EXISTS trg_countries_sync_delete;
DROP TRIGGER IF EXISTS trg_currencies_sync_delete;
DROP TRIGGER IF EXISTS trg_taxes_sync_delete;
DROP TRIGGER IF EXISTS trg_application_settings_sync_delete;
DROP TRIGGER IF EXISTS trg_branches_sync_delete;
DROP TRIGGER IF EXISTS trg_companies_sync_delete;

DROP TABLE IF EXISTS sync_tombstones;
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS sync_cursors;
DROP TABLE IF EXISTS sync_devices;


DROP TRIGGER IF EXISTS trg_audit_events_no_update;
DROP TRIGGER IF EXISTS trg_audit_events_no_delete;
DROP TABLE IF EXISTS audit_events;


DROP TABLE IF EXISTS exchange_rates;


DROP TABLE IF EXISTS taxes;


DROP TABLE IF EXISTS countries;


DROP TABLE IF EXISTS currencies;


DROP TABLE IF EXISTS application_settings;


DROP TRIGGER IF EXISTS trg_audit_logs_no_update;
DROP TRIGGER IF EXISTS trg_audit_logs_no_delete;
DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS local_profiles;


DROP TABLE IF EXISTS branches;


DROP TABLE IF EXISTS companies;


SELECT 1;
