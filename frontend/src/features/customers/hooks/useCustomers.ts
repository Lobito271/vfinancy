import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { customersService, type CustomerInput, type CustomerQuery } from '@/services/customers';
import { queryKeys } from '@/services/queryKeys';

export function useCustomers(query: CustomerQuery = {}) {
  return useQuery({
    queryKey: queryKeys.customers.list(query),
    queryFn: () => customersService.list(query),
  });
}

function useInvalidateCustomers() {
  const qc = useQueryClient();
  return () => void qc.invalidateQueries({ queryKey: queryKeys.customers.all });
}

export function useCreateCustomer() {
  const invalidate = useInvalidateCustomers();
  return useMutation({
    mutationFn: (input: CustomerInput) => customersService.create(input),
    onSuccess: invalidate,
  });
}

export function useUpdateCustomer() {
  const invalidate = useInvalidateCustomers();
  return useMutation({
    mutationFn: ({ id, ...input }: { id: string } & CustomerInput) =>
      customersService.update(id, input),
    onSuccess: invalidate,
  });
}

export function useDeleteCustomer() {
  const invalidate = useInvalidateCustomers();
  return useMutation({
    mutationFn: (id: string) => customersService.remove(id),
    onSuccess: invalidate,
  });
}
