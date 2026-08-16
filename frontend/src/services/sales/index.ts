import type { Sale, SaleStatus } from '@/types/domain';
import type { SaleDTO } from '../wails-types';
import { wailsClient } from '../bindings';

export interface SaleLineInput {
  productId: string;
  quantity: number;
  unitPrice: number;
  discountPercent: number;
  discountAmount: number;
  taxRate: number;
  taxAmount: number;
  costSnapshot: number;
  description: string;
}

export interface SaleCreateInput {
  customerId: string;
  date: string;
  dueDate?: string;
  notes?: string;
  items: SaleLineInput[];
}

function toSale(dto: SaleDTO): Sale {
  return {
    id: dto.id,
    number: dto.number,
    customerId: dto.customerId,
    customerName: dto.customerName,
    date: dto.date,
    status: dto.status as SaleStatus,
    subtotal: Number(dto.subtotal),
    tax: Number(dto.tax),
    discount: Number(dto.discount),
    total: Number(dto.total),
    cost: Number(dto.cost),
    profit: Number(dto.profit),
  };
}

export const salesService = {
  async list(): Promise<Sale[]> {
    const res = await wailsClient.listSales({ customerId: '', status: '', page: 1, pageSize: 200 });
    return res.items.map(toSale);
  },
  async get(id: string): Promise<Sale | null> {
    try {
      return toSale(await wailsClient.getSale(id));
    } catch {
      return null;
    }
  },
  async create(input: SaleCreateInput): Promise<Sale> {
    const created = await wailsClient.createSale({
      customerId: input.customerId,
      currencyCode: 'PEN',
      exchangeRate: '1.000000',
      dueDate: input.dueDate ?? input.date,
      notes: input.notes ?? '',
      items: input.items.map((it) => ({
        productId: it.productId,
        quantity: it.quantity.toFixed(4),
        unitPrice: it.unitPrice.toFixed(2),
        discountPercent: it.discountPercent.toFixed(2),
        discountAmount: it.discountAmount.toFixed(2),
        taxRate: it.taxRate.toFixed(2),
        taxAmount: it.taxAmount.toFixed(2),
        costSnapshot: it.costSnapshot.toFixed(2),
        description: it.description,
      })),
    });
    return toSale(created);
  },
  async cancel(id: string, reason: string): Promise<Sale> {
    return toSale(await wailsClient.cancelSale({ id, reason }));
  },

  async collectPayment(id: string, input: { paymentDate: string; method: string; reference: string; notes: string }): Promise<Sale> {
    return toSale(
      await wailsClient.registerSalePayment({
        id,
        paymentDate: input.paymentDate,
        method: input.method,
        reference: input.reference,
        notes: input.notes,
      }),
    );
  },
};
