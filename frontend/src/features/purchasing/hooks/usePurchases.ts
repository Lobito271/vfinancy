import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { purchasingService, type CustomerOrderCreateInput, type PurchaseCreateInput } from '@/services/purchasing';
import { queryKeys } from '@/services/queryKeys';

export interface MarkPaidInput {
  paymentDate: string;
  method: string;
  creditCardId: string;
  reference: string;
  notes: string;
}

export function usePurchases(search?: string) {
  return useQuery({
    queryKey: [...queryKeys.purchasing.list, search ?? ''],
    queryFn: () => purchasingService.list(search),
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

export function useMarkPurchaseReceived() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, arrivalDate }: { id: string; arrivalDate: string }) =>
      purchasingService.markReceived(id, { arrivalDate }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}

export function useMarkPurchaseFaulty() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: { arrivalDate: string; reason: string } }) =>
      purchasingService.markFaulty(id, input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}

export function useCustomerOrders(search?: string) {
  return useQuery({
    queryKey: [...queryKeys.purchasing.customerOrders, search ?? ''],
    queryFn: () => purchasingService.listCustomerOrders(search),
  });
}

export function useCustomerOrder(id?: string) {
  return useQuery({
    queryKey: queryKeys.purchasing.customerOrder(id ?? ''),
    queryFn: () => purchasingService.getCustomerOrder(id ?? ''),
    enabled: Boolean(id),
  });
}

export function useCreateCustomerOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CustomerOrderCreateInput) => purchasingService.createCustomerOrder(input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}

export function useRegisterCustomerOrderPayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      input,
    }: {
      id: string;
      input: { paymentDate: string; amount: number; method: string; reference: string; notes: string };
    }) => purchasingService.registerCustomerOrderPayment(id, input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}

export function useMarkCustomerOrderFaulty() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: { arrivalDate: string; reason: string } }) =>
      purchasingService.markCustomerOrderFaulty(id, input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}

export function useCancelCustomerOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => purchasingService.cancelCustomerOrder(id, reason),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.purchasing.all }),
  });
}
