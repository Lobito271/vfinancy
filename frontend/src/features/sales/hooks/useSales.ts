import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { salesService, type SaleCreateInput } from '@/services/sales';
import { queryKeys } from '@/services/queryKeys';

export function useSales(q: { search?: string } = {}) {
  return useQuery({
    queryKey: queryKeys.sales.list(q.search ?? null),
    queryFn: () => salesService.list({ search: q.search ?? '' }),
  });
}

export function useCreateSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: SaleCreateInput) => salesService.create(input),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.sales.all });
      void qc.invalidateQueries({ queryKey: queryKeys.dashboard.overview });
    },
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
      saleId,
      input,
    }: {
      saleId: string;
      input: { amount: number; paymentMethod: 'cash' | 'transfer' | 'other'; reference?: string; date?: string };
    }) => salesService.collectPayment(saleId, input),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.sales.all });
      void qc.invalidateQueries({ queryKey: queryKeys.dashboard.overview });
    },
  });
}
