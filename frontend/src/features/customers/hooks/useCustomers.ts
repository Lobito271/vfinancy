import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  customersService,
  type CustomerCreateInput,
  type CustomerQuery,
  type CustomerUpdateInput,
} from '@/services/customers';
import { queryKeys } from '@/services/queryKeys';

export function useCustomers(query: CustomerQuery = {}) {
  return useQuery({
    queryKey: queryKeys.customers.list(query),
    queryFn: () => customersService.list(query),
  });
}

export function useCustomer(id: string | undefined) {
  return useQuery({
    queryKey: id ? queryKeys.customers.detail(id) : ['noop'],
    queryFn: () => customersService.get(id!),
    enabled: !!id,
  });
}

export function useCreateCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CustomerCreateInput) => customersService.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.customers.all }),
  });
}

export function useUpdateCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CustomerUpdateInput) => customersService.update(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.customers.all }),
  });
}

export function useDeleteCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => customersService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.customers.all }),
  });
}

export function useCustomerOptions() {
  return useQuery({
    queryKey: queryKeys.customers.options,
    queryFn: () => customersService.getOptions(),
    staleTime: 5 * 60 * 1000,
  });
}
