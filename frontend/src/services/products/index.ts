import type { ProductDTO, SaveProductRequest } from '../wails-types';
import { wailsClient } from '../bindings';
import { fetchAllPages } from '../paginate';

export interface ProductQuery {
  search?: string;
  page?: number;
  pageSize?: number;
}

export type ProductInput = Omit<SaveProductRequest, 'id'>;

export const productsService = {
  async list(q: ProductQuery = {}): Promise<{ items: ProductDTO[]; total: number }> {
    const items = await fetchAllPages((page, pageSize) =>
      wailsClient.listProducts({ page, pageSize }, q.search ?? ''),
    );
    return { items, total: items.length };
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

  async setActive(id: string, active: boolean): Promise<void> {
    await wailsClient.setProductActive(id, active);
  },

  async getStock(id: string): Promise<number> {
    return wailsClient.getProductStock(id);
  },

  async getOptions(): Promise<ProductDTO[]> {
    return wailsClient.productOptions();
  },
};
