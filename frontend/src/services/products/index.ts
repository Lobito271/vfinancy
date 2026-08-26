import type { Product } from '@/types/domain';
import type { ProductDTO } from '../wails-types';
import { wailsClient } from '../bindings';
import { toFixed2 } from '@/utils/format';

export interface ProductQuery {
  search?: string;
  status?: string;
  page?: number;
  pageSize?: number;
}

export interface ProductCreateInput {
  sku: string;
  barcode?: string;
  description: string;
  categoryId?: string;
  brandId?: string;
  unitId?: string;
  taxId?: string;
  costUSD: number;
  salePrice: number;
  minStock: number;
  maxStock: number;
}

export interface ProductUpdateInput {
  id: string;
  description?: string;
  categoryId?: string;
  brandId?: string;
  costUSD?: number;
  salePrice?: number;
  minStock?: number;
  maxStock?: number;
  isActive?: boolean;
}

function toProduct(dto: ProductDTO): Product {
  return {
    id: dto.id,
    sku: dto.sku,
    barcode: dto.barcode || undefined,
    description: dto.description,
    categoryId: dto.categoryId || undefined,
    brandId: dto.brandId || undefined,
    category: dto.category || dto.categoryId,
    brand: dto.brand || dto.brandId,
    unit: dto.unit || dto.unitId,
    unitId: dto.unitId || undefined,
    taxId: dto.taxId || undefined,
    costUSD: Number(dto.costUSD),
    salePrice: Number(dto.salePrice),
    minStock: Number(dto.minStock),
    maxStock: Number(dto.maxStock),
    isActive: dto.isActive,
  };
}

export const productsService = {
  async list(q: ProductQuery = {}): Promise<{ items: Product[]; total: number }> {
    const res = await wailsClient.listProducts({
      search: q.search ?? '',
      status: q.status ?? '',
      page: q.page ?? 1,
      pageSize: q.pageSize ?? 200,
    });
    return { items: res.items.map(toProduct), total: res.total };
  },

  async get(id: string): Promise<Product | null> {
    try {
      return toProduct(await wailsClient.getProduct(id));
    } catch {
      return null;
    }
  },

  async create(input: ProductCreateInput): Promise<Product> {
    const created = await wailsClient.createProduct({
      sku: input.sku,
      barcode: input.barcode ?? '',
      description: input.description,
      categoryId: input.categoryId ?? '',
      brandId: input.brandId ?? '',
      unitId: input.unitId ?? '',
      taxId: input.taxId ?? '',
      costUSD: toFixed2(input.costUSD),
      salePrice: toFixed2(input.salePrice),
      saleCurrency: 'PEN',
      minStock: (input.minStock ?? 0).toFixed(4),
      maxStock: (input.maxStock ?? 0).toFixed(4),
      weight: '0.0000',
      isService: false,
    });
    return toProduct(created);
  },

  async update(id: string, input: ProductUpdateInput): Promise<Product> {
    const updated = await wailsClient.updateProduct({
      id,
      description: input.description ?? '',
      categoryId: input.categoryId ?? '',
      brandId: input.brandId ?? '',
      unitId: '',
      costUSD: toFixed2(input.costUSD),
      salePrice: toFixed2(input.salePrice),
      minStock: input.minStock != null ? input.minStock.toFixed(4) : '',
      maxStock: input.maxStock != null ? input.maxStock.toFixed(4) : '',
      isActive: input.isActive ?? null,
    });
    return toProduct(updated);
  },

  async remove(id: string): Promise<void> {
    await wailsClient.removeProduct(id);
  },

  async getOptions(): Promise<Array<{ value: string; label: string }>> {
    const res = await wailsClient.listProducts({ search: '', status: '', page: 1, pageSize: 200 });
    return res.items.map((p) => ({ value: p.id, label: `${p.sku} — ${p.description}` }));
  },
};
