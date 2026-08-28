import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  productsService,
  type ProductCreateInput,
  type ProductQuery,
  type ProductUpdateInput,
} from '@/services/products';
import { catalogService } from '@/services/catalog';
import { queryKeys } from '@/services/queryKeys';

export function useProducts(query: ProductQuery = {}) {
  return useQuery({
    queryKey: queryKeys.products.list(query),
    queryFn: () => productsService.list(query),
  });
}

export function useCategories() {
  return useQuery({
    queryKey: ['catalog', 'categories'],
    queryFn: () => catalogService.getCategoryOptions(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useBrands() {
  return useQuery({
    queryKey: ['catalog', 'brands'],
    queryFn: () => catalogService.getBrandOptions(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ProductCreateInput) => productsService.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.products.all }),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ProductUpdateInput) => productsService.update(input.id, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.products.all }),
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => productsService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.products.all }),
  });
}
