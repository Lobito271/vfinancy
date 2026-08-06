import type { Customer } from '@/data/mock';
import { customers as mockCustomers } from '@/data/mock';
import { sleep, generateId } from '@/utils';

export interface CustomerQuery {
  search?: string;
  status?: Customer['status'];
  page?: number;
  pageSize?: number;
  sortBy?: 'businessName' | 'documentNumber' | 'currentDebt' | 'createdAt';
  sortDir?: 'asc' | 'desc';
}

export interface CustomerCreateInput {
  documentType: Customer['documentType'];
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
  address?: string;
  creditLimit: number;
}

export interface CustomerUpdateInput extends Partial<CustomerCreateInput> {
  id: string;
  status?: Customer['status'];
}

let store: Customer[] = [...mockCustomers];

export const customersService = {
  async list(q: CustomerQuery = {}): Promise<{ items: Customer[]; total: number }> {
    await sleep(150);
    let rows = [...store];
    if (q.search) {
      const s = q.search.toLowerCase();
      rows = rows.filter(
        (c) =>
          c.businessName.toLowerCase().includes(s) ||
          c.documentNumber.includes(s) ||
          (c.contactName?.toLowerCase().includes(s) ?? false),
      );
    }
    if (q.status) rows = rows.filter((c) => c.status === q.status);
    const total = rows.length;
    if (q.sortBy) {
      rows.sort((a, b) => {
        const av = a[q.sortBy!] as string | number;
        const bv = b[q.sortBy!] as string | number;
        if (av < bv) return q.sortDir === 'desc' ? 1 : -1;
        if (av > bv) return q.sortDir === 'desc' ? -1 : 1;
        return 0;
      });
    }
    const start = ((q.page ?? 1) - 1) * (q.pageSize ?? 25);
    return { items: rows.slice(start, start + (q.pageSize ?? 25)), total };
  },

  async get(id: string): Promise<Customer | null> {
    await sleep(100);
    return store.find((c) => c.id === id) ?? null;
  },

  async create(input: CustomerCreateInput): Promise<Customer> {
    await sleep(200);
    const created: Customer = {
      id: generateId('c'),
      ...input,
      currentDebt: 0,
      totalPurchases: 0,
      status: 'active',
      createdAt: new Date().toISOString(),
    };
    store = [created, ...store];
    return created;
  },

  async update(input: CustomerUpdateInput): Promise<Customer> {
    await sleep(200);
    const idx = store.findIndex((c) => c.id === input.id);
    if (idx < 0) throw new Error('Customer not found');
    const updated: Customer = { ...store[idx], ...input };
    store = [...store.slice(0, idx), updated, ...store.slice(idx + 1)];
    return updated;
  },

  async remove(id: string): Promise<void> {
    await sleep(150);
    store = store.filter((c) => c.id !== id);
  },

  async getOptions(): Promise<Array<{ value: string; label: string }>> {
    return store.map((c) => ({ value: c.id, label: c.businessName }));
  },
};
