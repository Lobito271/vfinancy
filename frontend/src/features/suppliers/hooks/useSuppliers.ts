import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  suppliersService,
  type SupplierCreateInput,
  type SupplierQuery,
  type SupplierUpdateInput,
} from '@/services/suppliers';
import { queryKeys } from '@/services/queryKeys';

export function useSuppliers(query: SupplierQuery = {}) {
  return useQuery({
    queryKey: queryKeys.suppliers.list(query),
    queryFn: () => suppliersService.list(query),
  });
}

export function useSupplier(id: string | undefined) {
  return useQuery({
    queryKey: ['suppliers', 'detail', id],
    queryFn: () => suppliersService.get(id!),
    enabled: !!id,
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: SupplierCreateInput) => suppliersService.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.suppliers.all }),
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: SupplierUpdateInput) => suppliersService.update(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.suppliers.all }),
  });
}

export function useDeleteSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => suppliersService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.suppliers.all }),
  });
}
