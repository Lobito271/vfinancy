import type {
  CreateSaleRequest,
  SaleDTO,
  SalePaymentRequest,
} from '../wails-types';
import { wailsClient } from '../bindings';
import { fetchAllPages } from '../paginate';
import { customersService } from '../customers';
import type { Sale } from '@/types/domain';

export interface SaleLineInput {
  productId: string;
  quantity: number;
  unitPrice: number;
}

export interface SaleCreateInput {
  customerId: string;
  saleType?: CreateSaleRequest['saleType'];
  date: string;
  dueDate?: string;
  paymentMethod?: CreateSaleRequest['paymentMethod'];
  notes?: string;
  initialPayment?: number;
  items: SaleLineInput[];
}

export interface SaleQuery {
  search?: string;
  status?: string;
  saleType?: string;
  customerId?: string;
  from?: string;
  to?: string;
  onlyUnpaid?: boolean;
  page?: number;
  pageSize?: number;
}

type SaleRequest = Omit<CreateSaleRequest, 'items'>;

const round2 = (v: number) => Math.round(v * 100) / 100;

function toSale(dto: SaleDTO, nameById: Map<string, string>): Sale {
  return {
    id: dto.id,
    number: dto.number,
    customerId: dto.customerId,
    customerName: nameById.get(dto.customerId) ?? '',
    date: dto.saleDate,
    dueDate: dto.dueDate,
    status: dto.status as Sale['status'],
    saleType: dto.saleType,
    total: dto.total,
    costTotal: dto.costTotal,
    profit: dto.profit,
    paid: dto.paidAmount,
    balance: round2(dto.total - dto.paidAmount),
  };
}

async function customerNames(): Promise<Map<string, string>> {
  const options = await customersService.getOptions();
  return new Map(options.map((c) => [c.id, c.businessName]));
}

function emptyRequest(input: SaleCreateInput, items: SaleLineInput[]): SaleRequest & { items: SaleLineInput[] } {
  return {
    customerId: input.customerId,
    saleType: input.saleType ?? 'stock',
    date: input.date,
    dueDate: input.dueDate ?? '',
    paymentMethod: input.paymentMethod ?? 'cash',
    notes: input.notes ?? '',
    initialPayment: input.initialPayment ?? 0,
    items: items.map((it) => ({ productId: it.productId, quantity: it.quantity, unitPrice: it.unitPrice })),
  };
}

export const salesService = {
  async list(q: SaleQuery = {}): Promise<Sale[]> {
    const [items, names] = await Promise.all([
      fetchAllPages((page, pageSize) =>
        wailsClient.listSales({
          page,
          pageSize,
          search: q.search ?? '',
          status: q.status ?? '',
          saleType: q.saleType ?? '',
          customerId: q.customerId ?? '',
          from: q.from ?? '',
          to: q.to ?? '',
          onlyUnpaid: q.onlyUnpaid ?? false,
        }),
      ),
      customerNames(),
    ]);
    return items.map((dto) => toSale(dto as SaleDTO, names));
  },

  async get(id: string): Promise<Sale> {
    const [dto, names] = await Promise.all([wailsClient.getSale(id), customerNames()]);
    return toSale(dto, names);
  },

  async create(input: SaleCreateInput): Promise<Sale> {
    const [dto, names] = await Promise.all([
      wailsClient.createSale(emptyRequest(input, input.items)),
      customerNames(),
    ]);
    return toSale(dto, names);
  },

  async cancel(id: string, reason: string): Promise<Sale> {
    const [dto, names] = await Promise.all([wailsClient.cancelSale({ id, reason }), customerNames()]);
    return toSale(dto, names);
  },

  async collectPayment(
    saleId: string,
    input: { amount: number; paymentMethod: SalePaymentRequest['paymentMethod']; reference?: string; date?: string },
  ): Promise<Sale> {
    const [dto, names] = await Promise.all([
      wailsClient.registerSalePayment({
        saleId,
        amount: input.amount,
        paymentMethod: input.paymentMethod,
        reference: input.reference ?? '',
        date: input.date ?? '',
      }),
      customerNames(),
    ]);
    return toSale(dto, names);
  },
};
