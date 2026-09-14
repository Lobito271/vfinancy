import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { shipmentService, type ShipmentInput, type ShipmentQuery, type ShipmentStatus } from '@/services/shipments';
import { queryKeys } from '@/services/queryKeys';

export function useShipments(filters: ShipmentQuery = {}) {
  return useQuery({
    queryKey: queryKeys.shipments.list(filters),
    queryFn: () => shipmentService.list(filters),
  });
}

function invalidateShipments(qc: ReturnType<typeof useQueryClient>) {
  void qc.invalidateQueries({ queryKey: queryKeys.shipments.all });
}

export function useCreateShipment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ShipmentInput) => shipmentService.create(input),
    onSuccess: () => invalidateShipments(qc),
  });
}

export function useUpdateShipment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input, status }: { id: string; input: ShipmentInput; status: ShipmentStatus }) =>
      shipmentService.update(id, input, status),
    onSuccess: () => invalidateShipments(qc),
  });
}

export function useDeleteShipment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => shipmentService.remove(id),
    onSuccess: () => invalidateShipments(qc),
  });
}