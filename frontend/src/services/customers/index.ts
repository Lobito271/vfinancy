import type { Customer } from '@/types/domain';
import type { CustomerDTO } from '../wails-types';
import { wailsClient } from '../bindings';

export interface CustomerQuery {
  search?: string;
  status?: Customer['status'];
  page?: number;
  pageSize?: number;
  sortBy?: 'businessName' | 'documentNumber' | 'currentDebt' | 'createdAt';
  sortDir?: 'asc' | 'desc';
}

export interface CustomerCreateInput {
  documentType: Customer['documentType'];
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
  address?: string;
  creditLimit: number;
}

export interface CustomerUpdateInput extends Partial<CustomerCreateInput> {
  id: string;
  status?: Customer['status'];
}

function toCustomer(dto: CustomerDTO): Customer {
  return {
    id: dto.id,
    documentType: dto.documentType as Customer['documentType'],
    documentNumber: dto.documentNumber,
    businessName: dto.businessName,
    contactName: dto.tradeName || undefined,
    phone: dto.phone || undefined,
    email: dto.email || undefined,
    address: dto.address || undefined,
    creditLimit: Number(dto.creditLimit),
    currentDebt: Number(dto.currentDebt),
    status: dto.status as Customer['status'],
    totalPurchases: 0,
    createdAt: dto.createdAt,
  };
}

function money(value: number): string {
  return value.toFixed(2);
}

export const customersService = {
  async list(q: CustomerQuery = {}): Promise<{ items: Customer[]; total: number }> {
    const res = await wailsClient.listCustomers({
      search: q.search ?? '',
      status: q.status ?? '',
      page: q.page ?? 1,
      pageSize: q.pageSize ?? 25,
    });
    let rows = res.items.map(toCustomer);
    if (q.sortBy) {
      rows.sort((a, b) => {
        const av = a[q.sortBy!] as string | number;
        const bv = b[q.sortBy!] as string | number;
        if (av < bv) return q.sortDir === 'desc' ? 1 : -1;
        if (av > bv) return q.sortDir === 'desc' ? -1 : 1;
        return 0;
      });
    }
    return { items: rows, total: res.total };
  },

  async get(id: string): Promise<Customer | null> {
    try {
      return toCustomer(await wailsClient.getCustomer(id));
    } catch {
      return null;
    }
  },

  async create(input: CustomerCreateInput): Promise<Customer> {
    const created = await wailsClient.createCustomer({
      documentType: input.documentType,
      documentNumber: input.documentNumber,
      businessName: input.businessName,
      tradeName: input.contactName ?? '',
      creditLimit: money(input.creditLimit),
      paymentTermDays: 0,
      email: input.email ?? '',
      phone: input.phone ?? '',
      address: input.address ?? '',
    });
    return toCustomer(created);
  },

  async update(input: CustomerUpdateInput): Promise<Customer> {
    const updated = await wailsClient.updateCustomer({
      id: input.id,
      businessName: input.businessName ?? '',
      tradeName: input.contactName ?? '',
      creditLimit: money(input.creditLimit ?? 0),
      paymentTermDays: 0,
      email: input.email ?? '',
      phone: input.phone ?? '',
      address: input.address ?? '',
      isActive: input.status === 'inactive' ? false : input.status === 'active' ? true : null,
    });
    return toCustomer(updated);
  },

  async remove(id: string): Promise<void> {
    await wailsClient.removeCustomer(id);
  },

  async getOptions(): Promise<Array<{ value: string; label: string }>> {
    const res = await wailsClient.listCustomers({ search: '', status: '', page: 1, pageSize: 200 });
    return res.items.map((c) => ({ value: c.id, label: c.businessName }));
  },
};
