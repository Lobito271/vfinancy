import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { salesService, type SaleCreateInput } from '@/services/sales';
import { queryKeys } from '@/services/queryKeys';

export function useSales() {
  return useQuery({
    queryKey: queryKeys.sales.list,
    queryFn: () => salesService.list(),
  });
}

export function useSale(id: string | undefined) {
  return useQuery({
    queryKey: queryKeys.sales.detail(id!),
    queryFn: () => salesService.get(id!),
    enabled: !!id,
  });
}

export function useCreateSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: SaleCreateInput) => salesService.create(input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.sales.all }),
  });
}

export function useCancelSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => salesService.cancel(id, reason),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.sales.all }),
  });
}

export function useCollectSalePayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      input,
    }: {
      id: string;
      input: { paymentDate: string; method: string; reference: string; notes: string };
    }) => salesService.collectPayment(id, input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.sales.all }),
  });
}
