-- 0038_inventory_void_status.down.sql (PostgreSQL)
-- Module: Inventory
-- Reverts the "voided" batch status and "void_out" movement type.

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
