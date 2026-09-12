import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  productsService,
  type ProductInput,
  type ProductQuery,
} from '@/services/products';
import { queryKeys } from '@/services/queryKeys';

export function useProducts(query: ProductQuery = {}) {
  return useQuery({
    queryKey: queryKeys.products.list(query),
    queryFn: () => productsService.list(query),
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ProductInput) => productsService.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.products.all }),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...input }: { id: string } & ProductInput) => productsService.update(id, input),
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

export function useSetProductActive() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) => productsService.setActive(id, active),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.products.all }),
  });
}
