import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  inventoryService,
  type InventoryReceiveInput,
} from '@/services/inventory';
import { queryKeys } from '@/services/queryKeys';

export function useInventory() {
  return useQuery({
    queryKey: queryKeys.inventory.list,
    queryFn: () => inventoryService.list(),
  });
}

function useInvalidateInventory() {
  const qc = useQueryClient();
  return () => {
    void qc.invalidateQueries({ queryKey: queryKeys.inventory.all });
    void qc.invalidateQueries({ queryKey: queryKeys.products.all });
  };
}

export function useReceiveStock() {
  const invalidate = useInvalidateInventory();
  return useMutation({
    mutationFn: (input: InventoryReceiveInput) => inventoryService.receive(input),
    onSuccess: invalidate,
  });
}

export function useAdjustStock() {
  const invalidate = useInvalidateInventory();
  return useMutation({
    mutationFn: ({ batchId, delta, reason }: { batchId: string; delta: number; reason: string }) =>
      inventoryService.adjust(batchId, delta, reason),
    onSuccess: invalidate,
  });
}

export function useVoidStock() {
  const invalidate = useInvalidateInventory();
  return useMutation({
    mutationFn: ({ batchId, reason }: { batchId: string; reason?: string }) =>
      inventoryService.void(batchId, reason),
    onSuccess: invalidate,
  });
}
