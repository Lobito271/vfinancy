import type { Product } from '@/data/mock';
import { products as mockProducts } from '@/data/mock';
import { sleep, generateId } from '@/utils';

export interface ProductCreateInput {
  sku: string;
  barcode?: string;
  description: string;
  category: string;
  brand: string;
  cost: number;
  salePrice: number;
  minStock: number;
  maxStock: number;
}

let store: Product[] = [...mockProducts];

export const productsService = {
  async list(): Promise<Product[]> {
    await sleep(150);
    return [...store];
  },
  async get(id: string): Promise<Product | null> {
    await sleep(100);
    return store.find((p) => p.id === id) ?? null;
  },
  async create(input: ProductCreateInput): Promise<Product> {
    await sleep(200);
    const created: Product = {
      id: generateId('p'),
      ...input,
      currentStock: 0,
    };
    store = [created, ...store];
    return created;
  },
  async update(id: string, input: Partial<ProductCreateInput>): Promise<Product> {
    await sleep(200);
    const idx = store.findIndex((p) => p.id === id);
    if (idx < 0) throw new Error('Product not found');
    const updated: Product = { ...store[idx], ...input };
    store = [...store.slice(0, idx), updated, ...store.slice(idx + 1)];
    return updated;
  },
  async remove(id: string): Promise<void> {
    await sleep(150);
    store = store.filter((p) => p.id !== id);
  },
  async getOptions(): Promise<Array<{ value: string; label: string }>> {
    return store.map((p) => ({ value: p.id, label: `${p.sku} — ${p.description}` }));
  },
};
