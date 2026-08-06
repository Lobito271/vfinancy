import type { Supplier } from '@/data/mock';
import { suppliers as mockSuppliers } from '@/data/mock';
import { sleep, generateId } from '@/utils';

export interface SupplierCreateInput {
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
}

export interface SupplierUpdateInput extends Partial<SupplierCreateInput> {
  id: string;
  status?: Supplier['status'];
}

let store: Supplier[] = [...mockSuppliers];

export const suppliersService = {
  async list(): Promise<Supplier[]> {
    await sleep(150);
    return [...store];
  },
  async get(id: string): Promise<Supplier | null> {
    await sleep(100);
    return store.find((s) => s.id === id) ?? null;
  },
  async create(input: SupplierCreateInput): Promise<Supplier> {
    await sleep(200);
    const created: Supplier = {
      id: generateId('s'),
      ...input,
      currentDebt: 0,
      status: 'active',
    };
    store = [created, ...store];
    return created;
  },
  async update(input: SupplierUpdateInput): Promise<Supplier> {
    await sleep(200);
    const idx = store.findIndex((s) => s.id === input.id);
    if (idx < 0) throw new Error('Supplier not found');
    const updated: Supplier = { ...store[idx], ...input };
    store = [...store.slice(0, idx), updated, ...store.slice(idx + 1)];
    return updated;
  },
  async remove(id: string): Promise<void> {
    await sleep(150);
    store = store.filter((s) => s.id !== id);
  },
  async getOptions(): Promise<Array<{ value: string; label: string }>> {
    return store.map((s) => ({ value: s.id, label: s.businessName }));
  },
};
