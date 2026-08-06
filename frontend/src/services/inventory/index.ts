import type { InventoryItem } from '@/data/mock';
import { inventory as mockInventory } from '@/data/mock';
import type { InventoryMovementTypeCode } from '@/constants/status';
import { sleep, generateId } from '@/utils';

export interface InventoryMovement {
  id: string;
  date: string;
  type: InventoryMovementTypeCode;
  productId: string;
  warehouse: string;
  quantity: number;
  reference?: string;
  notes?: string;
}

let stockStore: InventoryItem[] = [...mockInventory];
let movementLog: InventoryMovement[] = [];

export const inventoryService = {
  async list(): Promise<InventoryItem[]> {
    await sleep(150);
    return [...stockStore];
  },
  async getClearance(): Promise<InventoryItem[]> {
    await sleep(100);
    return stockStore.filter((i) => i.isClearance);
  },
  async getLowStock(): Promise<InventoryItem[]> {
    await sleep(100);
    return stockStore.filter((i) => i.daysRemaining === 0 || i.daysRemaining < 5);
  },
  async recordMovement(input: Omit<InventoryMovement, 'id' | 'date'>): Promise<InventoryMovement> {
    await sleep(200);
    const movement: InventoryMovement = {
      id: generateId('mv'),
      date: new Date().toISOString(),
      ...input,
    };
    movementLog = [movement, ...movementLog];
    return movement;
  },
  async getMovements(): Promise<InventoryMovement[]> {
    await sleep(100);
    return [...movementLog];
  },
};
