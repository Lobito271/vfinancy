-- 0039_inventory_bridge_movements.down.sql (PostgreSQL)
-- Reverts to the pre-bridge movement type CHECK.

ALTER TABLE inventory_movements
    DROP CONSTRAINT inventory_movements_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_type_check
        CHECK (type IN ('purchase', 'sale', 'transfer_in', 'transfer_out', 'adjustment_in', 'adjustment_out', 'return_in', 'return_out', 'damage_out', 'void_out'));
