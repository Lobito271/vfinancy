import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { catalogService } from '@/services/catalog';
import { queryKeys } from '@/services/queryKeys';
import type {
  CreateBrandRequest,
  CreateCategoryRequest,
  UpdateBrandRequest,
  UpdateCategoryRequest,
} from '@/services/wails-types';

export function useCatalogCategories() {
  return useQuery({
    queryKey: queryKeys.catalog.categoryList,
    queryFn: () => catalogService.listCategories(),
  });
}

export function useCatalogBrands() {
  return useQuery({
    queryKey: queryKeys.catalog.brandList,
    queryFn: () => catalogService.listBrands(),
  });
}

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateCategoryRequest) => catalogService.createCategory(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.catalog.categoryList });
      qc.invalidateQueries({ queryKey: queryKeys.catalog.categories });
      qc.invalidateQueries({ queryKey: queryKeys.products.all });
    },
  });
}

export function useUpdateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateCategoryRequest) => catalogService.updateCategory(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.catalog.categoryList });
      qc.invalidateQueries({ queryKey: queryKeys.catalog.categories });
      qc.invalidateQueries({ queryKey: queryKeys.products.all });
    },
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => catalogService.deleteCategory(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.catalog.categoryList });
      qc.invalidateQueries({ queryKey: queryKeys.catalog.categories });
      qc.invalidateQueries({ queryKey: queryKeys.products.all });
    },
  });
}

export function useCreateBrand() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateBrandRequest) => catalogService.createBrand(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.catalog.brandList });
      qc.invalidateQueries({ queryKey: queryKeys.catalog.brands });
      qc.invalidateQueries({ queryKey: queryKeys.products.all });
    },
  });
}

export function useUpdateBrand() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateBrandRequest) => catalogService.updateBrand(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.catalog.brandList });
      qc.invalidateQueries({ queryKey: queryKeys.catalog.brands });
      qc.invalidateQueries({ queryKey: queryKeys.products.all });
    },
  });
}

export function useDeleteBrand() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => catalogService.deleteBrand(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.catalog.brandList });
      qc.invalidateQueries({ queryKey: queryKeys.catalog.brands });
      qc.invalidateQueries({ queryKey: queryKeys.products.all });
    },
  });
}
