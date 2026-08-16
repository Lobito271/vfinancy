import type { Supplier } from '@/types/domain';
import type { SupplierDTO } from '../wails-types';
import { wailsClient } from '../bindings';

export interface SupplierQuery {
  search?: string;
  status?: string;
  page?: number;
  pageSize?: number;
}

export interface SupplierCreateInput {
  documentType: string;
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
  address?: string;
  defaultCurrency?: string;
  paymentTermDays?: number;
}

export interface SupplierUpdateInput {
  id: string;
  businessName?: string;
  contactName?: string;
  phone?: string;
  email?: string;
  address?: string;
  paymentTermDays?: number;
  status?: 'active' | 'inactive';
}

function toSupplier(dto: SupplierDTO): Supplier {
  return {
    id: dto.id,
    documentType: (dto.documentType as Supplier['documentType']) || undefined,
    documentNumber: dto.documentNumber,
    businessName: dto.businessName,
    contactName: dto.tradeName || undefined,
    phone: dto.phone || undefined,
    email: dto.email || undefined,
    address: dto.address || undefined,
    paymentTermDays: dto.paymentTermDays || undefined,
    currentDebt: Number(dto.currentDebt),
    status: (dto.status === 'active' ? 'active' : 'inactive') as Supplier['status'],
  };
}

export const suppliersService = {
  async list(q: SupplierQuery = {}): Promise<{ items: Supplier[]; total: number }> {
    const res = await wailsClient.listSuppliers({
      search: q.search ?? '',
      status: q.status ?? '',
      page: q.page ?? 1,
      pageSize: q.pageSize ?? 200,
    });
    return { items: res.items.map(toSupplier), total: res.total };
  },

  async get(id: string): Promise<Supplier | null> {
    try {
      return toSupplier(await wailsClient.getSupplier(id));
    } catch {
      return null;
    }
  },

  async create(input: SupplierCreateInput): Promise<Supplier> {
    const created = await wailsClient.createSupplier({
      documentType: input.documentType,
      documentNumber: input.documentNumber,
      businessName: input.businessName,
      tradeName: input.contactName ?? '',
      taxId: '',
      isInternational: false,
      defaultCurrency: input.defaultCurrency ?? 'PEN',
      paymentTermDays: input.paymentTermDays ?? 0,
      email: input.email ?? '',
      phone: input.phone ?? '',
      address: input.address ?? '',
    });
    return toSupplier(created);
  },

  async update(input: SupplierUpdateInput): Promise<Supplier> {
    const updated = await wailsClient.updateSupplier({
      id: input.id,
      businessName: input.businessName ?? '',
      tradeName: input.contactName ?? '',
      taxId: '',
      paymentTermDays: input.paymentTermDays ?? 0,
      email: input.email ?? '',
      phone: input.phone ?? '',
      address: input.address ?? '',
      isActive: input.status === 'inactive' ? false : input.status === 'active' ? true : null,
    });
    return toSupplier(updated);
  },

  async remove(id: string): Promise<void> {
    await wailsClient.removeSupplier(id);
  },

  async getOptions(): Promise<Array<{ value: string; label: string }>> {
    const res = await wailsClient.listSuppliers({ search: '', status: '', page: 1, pageSize: 200 });
    return res.items.map((s) => ({ value: s.id, label: s.businessName }));
  },
};
