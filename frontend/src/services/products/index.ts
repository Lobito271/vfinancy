import type { ProductDTO, SaveProductRequest } from '../wails-types';
import { wailsClient } from '../bindings';

export interface ProductQuery {
  search?: string;
  page?: number;
  pageSize?: number;
}

export type ProductInput = Omit<SaveProductRequest, 'id'>;

export const productsService = {
  async list(q: ProductQuery = {}): Promise<{ items: ProductDTO[]; total: number }> {
    const res = await wailsClient.listProducts(
      { page: q.page ?? 1, pageSize: q.pageSize ?? 200 },
      q.search ?? '',
    );
    return { items: res.items as ProductDTO[], total: res.total };
  },

  async get(id: string): Promise<ProductDTO> {
    return wailsClient.getProduct(id);
  },

  async create(input: ProductInput): Promise<ProductDTO> {
    return wailsClient.createProduct(input);
  },

  async update(id: string, input: ProductInput): Promise<ProductDTO> {
    return wailsClient.updateProduct({ ...input, id });
  },

  async remove(id: string): Promise<void> {
    await wailsClient.removeProduct(id);
  },

  async getStock(id: string): Promise<number> {
    return wailsClient.getProductStock(id);
  },

  async getOptions(): Promise<ProductDTO[]> {
    return wailsClient.productOptions();
  },
};
