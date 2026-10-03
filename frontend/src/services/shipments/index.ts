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
  saleId: string;
  /** Sent on create only; the backend rejects changes afterwards. */
  securityCode?: string;
  location?: string;
  shipmentDate?: string;
  deliveredAt?: string;
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
      status: 'pending',
      notes: input.notes ?? '',
      location: input.location ?? '',
      shipmentDate: input.shipmentDate ?? '',
      deliveredAt: input.deliveredAt ?? '',
    });
  },
  async update(id: string, input: ShipmentInput, status: ShipmentStatus): Promise<ShipmentDTO> {
    return wailsClient.updateShipment({
      ...input,
      id,
      status,
      notes: input.notes ?? '',
      location: input.location ?? '',
      shipmentDate: input.shipmentDate ?? '',
      deliveredAt: input.deliveredAt ?? '',
    });
  },
  async remove(id: string): Promise<void> {
    await wailsClient.deleteShipment(id);
  },
};