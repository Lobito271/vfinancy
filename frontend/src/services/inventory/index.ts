import type { InventoryItem } from '@/types/domain';
import type { InventoryMovementTypeCode } from '@/constants/status';
import type { InventoryBatchDTO, InventoryMovementDTO } from '../wails-types';
import { wailsClient } from '../bindings';
import { productsService } from '../products';

export interface InventoryMovement {
  id: string;
  date: string;
  type: InventoryMovementTypeCode;
  productId: string;
  productDescription: string;
  warehouse: string;
  quantity: number;
  reference?: string;
  notes?: string;
}

function daysBetween(from: string, to: string): number {
  const ms = new Date(to).getTime() - new Date(from).getTime();
  return Math.floor(ms / 86_400_000);
}

function toInventoryItem(
  dto: InventoryBatchDTO,
  products: Map<string, { sku: string; description: string }>,
  warehouses: Map<string, string>,
): InventoryItem {
  const product = products.get(dto.productId);
  return {
    id: dto.id,
    productId: dto.productId,
    productSku: product?.sku ?? '',
    productDescription: product?.description ?? 'Producto',
    warehouse: warehouses.get(dto.warehouseId) ?? dto.warehouseId,
    quantity: Number(dto.currentQuantity),
    unitCost: Number(dto.unitCost) || 0,
    currencyCode: dto.currencyCode || 'PEN',
    arrivalDate: dto.arrivalDate,
    maxSaleDate: dto.maxSaleDate,
    ageDays: Math.max(0, daysBetween(dto.arrivalDate, new Date().toISOString())),
    daysRemaining: dto.maxSaleDate ? daysBetween(new Date().toISOString(), dto.maxSaleDate) : 0,
    isClearance: dto.isClearance,
    status: dto.status,
  };
}

async function productIndex(): Promise<Map<string, { sku: string; description: string }>> {
  const { items } = await productsService.list();
  return new Map(items.map((p) => [p.id, { sku: p.sku, description: p.description }]));
}

async function warehouseIndex(): Promise<Map<string, string>> {
  const warehouses = await wailsClient.listWarehouses();
  return new Map(warehouses.map((w) => [w.id, `${w.code} — ${w.name}`]));
}

async function fetchBatches(onlyClearance: boolean): Promise<InventoryItem[]> {
  const [res, products, warehouses] = await Promise.all([
    wailsClient.listInventoryBatches({ onlyClearance, page: 1, pageSize: 200 }),
    productIndex(),
    warehouseIndex(),
  ]);
  return res.items.map((dto) => toInventoryItem(dto, products, warehouses));
}

export const inventoryService = {
  async list(): Promise<InventoryItem[]> {
    return fetchBatches(false);
  },
  async getClearance(): Promise<InventoryItem[]> {
    return fetchBatches(true);
  },
  async getLowStock(): Promise<Array<InventoryItem & { minStock: number }>> {
    const [{ items: products }, batches] = await Promise.all([
      productsService.list(),
      this.list(),
    ]);
    const stockByProduct = new Map<string, number>();
    for (const batch of batches) {
      stockByProduct.set(batch.productId, (stockByProduct.get(batch.productId) ?? 0) + batch.quantity);
    }
    return products
      .filter((p) => (stockByProduct.get(p.id) ?? 0) < p.minStock)
      .map((p) => ({
        id: p.id,
        productId: p.id,
        productSku: p.sku,
        productDescription: p.description,
        warehouse: '',
        quantity: stockByProduct.get(p.id) ?? 0,
        unitCost: 0,
        currencyCode: 'PEN',
        arrivalDate: '',
        maxSaleDate: '',
        ageDays: 0,
        daysRemaining: 0,
        isClearance: false,
        status: 'active',
        minStock: p.minStock,
      }));
  },
  async getMovements(productId?: string): Promise<InventoryMovement[]> {
    const [res, products] = await Promise.all([
      wailsClient.listInventoryMovements({ productId: productId ?? '', page: 1, pageSize: 200 }),
      productIndex(),
    ]);
    return res.items.map((dto: InventoryMovementDTO) => ({
      id: dto.id,
      date: dto.occurredAt,
      type: dto.type as InventoryMovementTypeCode,
      productId: dto.productId,
      productDescription: products.get(dto.productId)?.description ?? 'Producto',
      warehouse: dto.warehouseId,
      quantity: Number(dto.quantity),
      notes: dto.notes || undefined,
    }));
  },

  async receive(input: InventoryReceiveInput): Promise<void> {
    await wailsClient.receiveStock({
      productId: input.productId,
      warehouseId: input.warehouseId,
      lotNumber: input.lotNumber,
      arrivalDate: input.arrivalDate,
      quantity: input.quantity.toFixed(4),
      unitCost: input.unitCost.toFixed(2),
      currencyCode: 'PEN',
    });
  },

  async adjust(batchId: string, delta: number, reason: string): Promise<void> {
    await wailsClient.adjustStock({ batchId, delta: delta.toFixed(4), reason });
  },

  async issue(batchId: string, quantity: number): Promise<void> {
    await wailsClient.issueStock({ batchId, quantity: quantity.toFixed(4) });
  },

  async void(batchId: string, reason?: string): Promise<void> {
    await wailsClient.voidStock({ batchId, reason: reason ?? '' });
  },
};

export interface InventoryReceiveInput {
  productId: string;
  warehouseId: string;
  lotNumber: string;
  arrivalDate: string;
  quantity: number;
  unitCost: number;
}
