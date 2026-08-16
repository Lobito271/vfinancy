import type { Purchase } from '@/types/domain';
import type { PurchaseOrderDTO } from '../wails-types';
import { wailsClient } from '../bindings';
import { suppliersService } from '../suppliers';

export interface PurchaseLineInput {
  productId: string;
  quantity: number;
  unitPrice: number;
  discountPercent: number;
  discountAmount: number;
  taxRate: number;
  taxAmount: number;
  description: string;
}

export interface PurchaseCreateInput {
  supplierId: string;
  orderDate: string;
  notes?: string;
  items: PurchaseLineInput[];
}

export const purchasingService = {
  async list(): Promise<Purchase[]> {
    const [res, { items: suppliers }] = await Promise.all([
      wailsClient.listPurchaseOrders({ status: '', page: 1, pageSize: 200 }),
      suppliersService.list(),
    ]);
    const nameById = new Map(suppliers.map((s) => [s.id, s.businessName]));
    return res.items.map((dto: PurchaseOrderDTO) => ({
      id: dto.id,
      number: dto.number,
      supplierId: dto.supplierId,
      supplierName: nameById.get(dto.supplierId) ?? '',
      date: dto.orderDate,
      status: dto.status,
      total: Number(dto.total),
    }));
  },

  async get(id: string): Promise<Purchase | null> {
    try {
      const dto = await wailsClient.getPurchaseOrder(id);
      const supplier = await suppliersService.get(dto.supplierId);
      return {
        id: dto.id,
        number: dto.number,
        supplierId: dto.supplierId,
        supplierName: supplier?.businessName ?? '',
        date: dto.orderDate,
        status: dto.status,
        total: Number(dto.total),
      };
    } catch {
      return null;
    }
  },

  async create(input: PurchaseCreateInput): Promise<Purchase> {
    const created = await wailsClient.createPurchaseOrder({
      supplierId: input.supplierId,
      currencyCode: 'PEN',
      exchangeRate: '1.000000',
      orderDate: input.orderDate,
      notes: input.notes ?? '',
      items: input.items.map((it) => ({
        productId: it.productId,
        quantity: it.quantity.toFixed(4),
        unitPrice: it.unitPrice.toFixed(2),
        discountPercent: it.discountPercent.toFixed(2),
        discountAmount: it.discountAmount.toFixed(2),
        taxRate: it.taxRate.toFixed(2),
        taxAmount: it.taxAmount.toFixed(2),
        description: it.description,
      })),
    });
    const supplier = await suppliersService.get(created.supplierId);
    return {
      id: created.id,
      number: created.number,
      supplierId: created.supplierId,
      supplierName: supplier?.businessName ?? '',
      date: created.orderDate,
      status: created.status,
      total: Number(created.total),
    };
  },

  async cancel(id: string, reason: string): Promise<void> {
    await wailsClient.cancelPurchaseOrder({ id, reason });
  },

  async markPaid(id: string, input: { paymentDate: string; method: string; reference: string; notes: string }): Promise<Purchase> {
    const dto = await wailsClient.registerPurchasePayment({
      id,
      paymentDate: input.paymentDate,
      method: input.method,
      reference: input.reference,
      notes: input.notes,
    });
    const supplier = await suppliersService.get(dto.supplierId);
    return {
      id: dto.id,
      number: dto.number,
      supplierId: dto.supplierId,
      supplierName: supplier?.businessName ?? '',
      date: dto.orderDate,
      status: dto.status,
      total: Number(dto.total),
    };
  },
};
