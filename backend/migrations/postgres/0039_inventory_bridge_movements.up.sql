-- 0039_inventory_bridge_movements.up.sql (PostgreSQL)
-- Module: Inventory
-- Adds the automated transaction-bridge movement types:
--   purchase_receipt (inbound, goods received from a purchase order),
--   void_sale        (inbound, stock returned when a sale is voided),
--   void_purchase    (outbound, stock deducted when a purchase is voided).

ALTER TABLE inventory_movements
    DROP CONSTRAINT inventory_movements_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_type_check
        CHECK (type IN ('purchase', 'purchase_receipt', 'sale', 'void_sale', 'void_purchase', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out', 'void_out'));
