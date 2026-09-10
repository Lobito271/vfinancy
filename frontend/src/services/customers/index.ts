import type { CustomerDTO, SaveCustomerRequest } from '../wails-types';
import { wailsClient } from '../bindings';
import { fetchAllPages } from '../paginate';

export interface CustomerQuery {
  search?: string;
  page?: number;
  pageSize?: number;
}

export type CustomerInput = Omit<SaveCustomerRequest, 'id'>;

export const customersService = {
  async list(q: CustomerQuery = {}): Promise<{ items: CustomerDTO[]; total: number }> {
    const items = await fetchAllPages((page, pageSize) =>
      wailsClient.listCustomers({ page, pageSize }, q.search ?? ''),
    );
    return { items, total: items.length };
  },

  async getOptions(): Promise<CustomerDTO[]> {
    return wailsClient.customerOptions();
  },

  async create(input: CustomerInput): Promise<CustomerDTO> {
    return wailsClient.createCustomer(input);
  },

  async update(id: string, input: CustomerInput): Promise<CustomerDTO> {
    return wailsClient.updateCustomer({ ...input, id });
  },

  async remove(id: string): Promise<void> {
    await wailsClient.removeCustomer(id);
  },

  async getByDocument(docType: string, docNumber: string): Promise<CustomerDTO> {
    return wailsClient.getCustomerByDocument(docType, docNumber);
  },
};
