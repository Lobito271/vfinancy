import type { Sale, SaleStatus } from '@/data/mock';
import { sales as mockSales } from '@/data/mock';
import { sleep, generateId } from '@/utils';

export interface SaleCreateInput {
  customerId: string;
  date: string;
  status: SaleStatus;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  cost: number;
}

let store: Sale[] = [...mockSales];

export const salesService = {
  async list(): Promise<Sale[]> {
    await sleep(150);
    return [...store];
  },
  async get(id: string): Promise<Sale | null> {
    await sleep(100);
    return store.find((s) => s.id === id) ?? null;
  },
  async create(input: SaleCreateInput): Promise<Sale> {
    await sleep(200);
    const customer = (await import('@/services/customers')).customersService.get(input.customerId);
    const created: Sale = {
      id: generateId('sa'),
      number: `V-${new Date().getFullYear()}-${String(store.length + 1).padStart(5, '0')}`,
      customerName: (await customer)?.businessName ?? '',
      profit: input.total - input.cost - input.tax,
      ...input,
    };
    store = [created, ...store];
    return created;
  },
};
