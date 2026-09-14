import type { PaginationRequest, ShipmentDTO } from '../wails-types';
import { wailsClient } from '../bindings';
import { fetchAllPages } from '../paginate';

export type ShipmentStatus = ShipmentDTO['status'];

export const SHIPMENT_STATUSES: { value: ShipmentStatus; label: string }[] = [
  { value: 'pending', label: 'Pendiente' },
  { value: 'shipped', label: 'Enviado' },
  { value: 'delivered', label: 'Entregado' },
];

export interface ShipmentQuery {
  search?: string;
  status?: string;
}

export interface ShipmentInput {
  customerId?: string;
  saleId?: string;
  description: string;
  notes?: string;
}

export const shipmentService = {
  async list(q: ShipmentQuery = {}): Promise<ShipmentDTO[]> {
    return fetchAllPages((page, pageSize) =>
      wailsClient.listShipments({ page, pageSize } as PaginationRequest, q.search ?? '', q.status ?? ''),
    );
  },
  async get(id: string): Promise<ShipmentDTO> {
    return wailsClient.getShipment(id);
  },
  async create(input: ShipmentInput): Promise<ShipmentDTO> {
    return wailsClient.createShipment({
      ...input,
      customerId: input.customerId ?? '',
      saleId: input.saleId ?? '',
      status: 'pending',
      notes: input.notes ?? '',
    });
  },
  async update(id: string, input: ShipmentInput, status: ShipmentStatus): Promise<ShipmentDTO> {
    return wailsClient.updateShipment({
      ...input,
      id,
      customerId: input.customerId ?? '',
      saleId: input.saleId ?? '',
      status,
      notes: input.notes ?? '',
    });
  },
  async remove(id: string): Promise<void> {
    await wailsClient.deleteShipment(id);
  },
};