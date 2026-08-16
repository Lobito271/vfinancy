import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { purchasingService, type PurchaseCreateInput } from '@/services/purchasing';
import { queryKeys } from '@/services/queryKeys';

export interface MarkPaidInput {
  paymentDate: string;
  method: string;
  reference: string;
  notes: string;
}

export function usePurchases() {
  return useQuery({
    queryKey: queryKeys.purchasing.list,
    queryFn: () => purchasingService.list(),
  });
}

export function useCreatePurchase() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: PurchaseCreateInput) => purchasingService.create(input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}

export function useCancelPurchase() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => purchasingService.cancel(id, reason),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}

export function useMarkPurchasePaid() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: MarkPaidInput }) => purchasingService.markPaid(id, input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}
