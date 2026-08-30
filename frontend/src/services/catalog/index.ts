import type { BrandDTO, CategoryDTO, CreateBrandRequest, CreateCategoryRequest, UnitDTO, UpdateBrandRequest, UpdateCategoryRequest } from '../wails-types';
import { wailsClient } from '../bindings';

interface SelectOptionLike {
  value: string;
  label: string;
}

export const catalogService = {
  async listUnits(): Promise<UnitDTO[]> {
    return wailsClient.listUnits();
  },

  async getUnitOptions(): Promise<SelectOptionLike[]> {
    const units = await wailsClient.listUnits();
    return units.map((u) => ({
      value: u.id,
      label: u.symbol ? `${u.code} — ${u.name} (${u.symbol})` : `${u.code} — ${u.name}`,
    }));
  },

  async listCategories(): Promise<CategoryDTO[]> {
    return wailsClient.listCategories();
  },

  async getCategoryOptions(): Promise<SelectOptionLike[]> {
    const categories = await wailsClient.listCategories();
    return categories.map((c) => ({ value: c.id, label: c.name }));
  },

  async createCategory(req: CreateCategoryRequest): Promise<CategoryDTO> {
    return wailsClient.createCategory(req);
  },

  async updateCategory(req: UpdateCategoryRequest): Promise<CategoryDTO> {
    return wailsClient.updateCategory(req);
  },

  async deleteCategory(id: string): Promise<void> {
    return wailsClient.deleteCategory(id);
  },

  async listBrands(): Promise<BrandDTO[]> {
    return wailsClient.listBrands();
  },

  async getBrandOptions(): Promise<SelectOptionLike[]> {
    const brands = await wailsClient.listBrands();
    return brands.map((b) => ({ value: b.id, label: b.name }));
  },

  async createBrand(req: CreateBrandRequest): Promise<BrandDTO> {
    return wailsClient.createBrand(req);
  },

  async updateBrand(req: UpdateBrandRequest): Promise<BrandDTO> {
    return wailsClient.updateBrand(req);
  },

  async deleteBrand(id: string): Promise<void> {
    return wailsClient.deleteBrand(id);
  },
};
