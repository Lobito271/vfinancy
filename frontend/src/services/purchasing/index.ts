import type { CustomerOrder, CustomerOrderPayment, Purchase } from '@/types/domain';
import type { PurchaseOrderDTO } from '../wails-types';
import { wailsClient } from '../bindings';
import { customersService } from '../customers';
import { suppliersService } from '../suppliers';

interface PurchaseLineInput {
  productId: string;
  quantity: number;
  unitPrice: number;
  discountPercent: number;
  discountAmount: number;
  taxRate: number;
  taxAmount: number;
  description: string;
}

export interface PurchaseCreateInput {
  supplierId: string;
  creditCardId: string;
  orderDate: string;
  supplierOrderNumber: string;
  exchangeRate?: number;
  notes?: string;
  items: PurchaseLineInput[];
}

export interface CustomerOrderCreateInput {
  supplierId: string;
  customerId: string;
  creditCardId: string;
  orderDate: string;
  supplierOrderNumber: string;
  costUSD: number;
  salePricePEN: number;
  anticipo: number;
  anticipoDate: string;
  exchangeRate?: number;
  notes?: string;
  items: PurchaseLineInput[];
}

export const purchasingService = {
  async list(search?: string): Promise<Purchase[]> {
    const [res, { items: suppliers }] = await Promise.all([
      wailsClient.listPurchaseOrders({ status: '', orderType: '', search: search ?? '', page: 1, pageSize: 200 }),
      suppliersService.list(),
    ]);
    const nameById = new Map(suppliers.map((s) => [s.id, s.businessName]));
    return res.items.map((dto: PurchaseOrderDTO) => mapPurchase(dto, nameById.get(dto.supplierId) ?? ''));
  },

  async get(id: string): Promise<Purchase | null> {
    try {
      const dto = await wailsClient.getPurchaseOrder(id);
      const supplier = await suppliersService.get(dto.supplierId);
      return mapPurchase(dto, supplier?.businessName ?? '');
    } catch {
      return null;
    }
  },

  async create(input: PurchaseCreateInput): Promise<Purchase> {
    const created = await wailsClient.createPurchaseOrder({
      supplierId: input.supplierId,
      customerId: '',
      creditCardId: input.creditCardId,
      orderType: '',
      currencyCode: 'USD',
      exchangeRate: (input.exchangeRate ?? 1).toFixed(6),
      orderDate: input.orderDate,
      supplierOrderNumber: input.supplierOrderNumber,
      costUSD: input.items.reduce(
        (sum, it) => sum + it.unitPrice * it.quantity - it.discountAmount,
        0,
      ).toFixed(2),
      salePricePEN: '0.00',
      anticipo: '0.00',
      anticipoDate: '',
      notes: input.notes ?? '',
      items: input.items.map((it) => ({
        productId: it.productId,
        quantity: it.quantity.toFixed(4),
        unitPrice: it.unitPrice.toFixed(2),
        discountPercent: it.discountPercent.toFixed(2),
        discountAmount: it.discountAmount.toFixed(2),
        taxRate: it.taxRate.toFixed(2),
        taxAmount: it.taxAmount.toFixed(2),
        description: it.description,
      })),
    });
    const supplier = await suppliersService.get(created.supplierId);
    return mapPurchase(created, supplier?.businessName ?? '');
  },

  async cancel(id: string, reason: string): Promise<void> {
    await wailsClient.cancelPurchaseOrder({ id, reason });
  },

  async markFaulty(id: string, input: { arrivalDate: string; reason: string }): Promise<Purchase> {
    const dto = await wailsClient.markPurchaseFaulty({
      id,
      arrivalDate: input.arrivalDate,
      reason: input.reason,
    });
    const supplier = await suppliersService.get(dto.supplierId);
    return mapPurchase(dto, supplier?.businessName ?? '');
  },

  async markReceived(id: string, input: { arrivalDate: string }): Promise<Purchase> {
    const dto = await wailsClient.markPurchaseReceived({
      id,
      arrivalDate: input.arrivalDate,
    });
    const supplier = await suppliersService.get(dto.supplierId);
    return mapPurchase(dto, supplier?.businessName ?? '');
  },

  async registerCustomerOrderPayment(
    purchaseId: string,
    input: { paymentDate: string; amount: number; method: string; reference: string; notes: string },
  ): Promise<void> {
    await wailsClient.registerCustomerOrderPayment({
      purchaseId,
      paymentDate: input.paymentDate,
      amount: input.amount.toFixed(2),
      currencyCode: 'PEN',
      exchangeRate: '1.000000',
      method: input.method,
      reference: input.reference,
      notes: input.notes,
    });
  },

  async markPaid(id: string, input: { paymentDate: string; method: string; creditCardId: string; reference: string; notes: string }): Promise<Purchase> {
    const dto = await wailsClient.registerPurchasePayment({
      id,
      paymentDate: input.paymentDate,
      method: input.creditCardId ? 'card' : input.method,
      reference: input.reference,
      notes: input.notes,
    });
    const supplier = await suppliersService.get(dto.supplierId);
    return mapPurchase(dto, supplier?.businessName ?? '');
  },

  async listCustomerOrders(search?: string): Promise<CustomerOrder[]> {
    const [res, { items: customers }] = await Promise.all([
      wailsClient.listPurchaseOrders({ status: '', orderType: 'customer', search: search ?? '', page: 1, pageSize: 200 }),
      customersService.list({ page: 1, pageSize: 200 }),
    ]);
    const nameById = new Map(customers.map((c) => [c.id, c.businessName]));
    return res.items.map((dto) => mapCustomerOrder(dto, nameById.get(dto.customerId) ?? ''));
  },

  async getCustomerOrder(id: string): Promise<CustomerOrder | null> {
    try {
      const dto = await wailsClient.getPurchaseOrder(id);
      let customerName = '';
      if (dto.customerId) {
        const customer = await customersService.get(dto.customerId);
        customerName = customer?.businessName ?? '';
      }
      return mapCustomerOrder(dto, customerName);
    } catch {
      return null;
    }
  },

  async createCustomerOrder(input: CustomerOrderCreateInput): Promise<CustomerOrder> {
    const dto = await wailsClient.createPurchaseOrder({
      supplierId: input.supplierId,
      customerId: input.customerId,
      creditCardId: input.creditCardId,
      orderType: 'customer',
      currencyCode: 'USD',
      exchangeRate: (input.exchangeRate ?? 1).toFixed(6),
      orderDate: input.orderDate,
      supplierOrderNumber: input.supplierOrderNumber,
      costUSD: input.costUSD.toFixed(2),
      salePricePEN: input.salePricePEN.toFixed(2),
      anticipo: input.anticipo.toFixed(2),
      anticipoDate: input.anticipoDate,
      notes: input.notes ?? '',
      items: input.items.map((it) => ({
        productId: it.productId,
        quantity: it.quantity.toFixed(4),
        unitPrice: it.unitPrice.toFixed(2),
        discountPercent: it.discountPercent.toFixed(2),
        discountAmount: it.discountAmount.toFixed(2),
        taxRate: it.taxRate.toFixed(2),
        taxAmount: it.taxAmount.toFixed(2),
        description: it.description,
      })),
    });
    return mapCustomerOrder(dto, '');
  },

  async markCustomerOrderFaulty(id: string, input: { arrivalDate: string; reason: string }): Promise<void> {
    await wailsClient.markPurchaseFaulty({ id, arrivalDate: input.arrivalDate, reason: input.reason });
  },

  async cancelCustomerOrder(id: string, reason: string): Promise<void> {
    await wailsClient.cancelPurchaseOrder({ id, reason });
  },
};

function mapCustomerOrder(dto: PurchaseOrderDTO, customerName: string): CustomerOrder {
  const payments: CustomerOrderPayment[] = (dto.payments ?? []).map((p) => ({
    id: p.id,
    purchaseOrderId: p.purchaseOrderId,
    number: p.number,
    paymentDate: p.paymentDate,
    amount: Number(p.amount),
    method: p.method,
    currencyCode: p.currencyCode,
    exchangeRate: Number(p.exchangeRate),
    reference: p.reference,
    notes: p.notes,
    status: p.status === 'refunded' ? 'refunded' : 'active',
    refundedAmount: Number(p.refundedAmount),
    refundedAt: p.refundedAt,
    refundReason: p.refundReason,
  }));
  return {
    id: dto.id,
    number: dto.number,
    supplierId: dto.supplierId,
    customerId: dto.customerId,
    customerName,
    creditCardId: dto.creditCardId,
    orderType: dto.orderType,
    date: dto.orderDate,
    status: dto.status,
    subtotal: Number(dto.subtotal),
    discount: Number(dto.discount),
    tax: Number(dto.tax),
    total: Number(dto.total),
    paid: Number(dto.paid),
    costUSD: Number(dto.costUSD),
    salePricePEN: Number(dto.salePricePEN),
    realCostPEN: Number(dto.realCostPEN),
    projectedProfitPEN: Number(dto.projectedProfitPEN),
    anticipo: Number(dto.anticipo),
    anticipoDate: dto.anticipoDate,
    porCobrar: Number(dto.porCobrar),
    supplierOrderNumber: dto.supplierOrderNumber,
    faulty: dto.faulty,
    faultyReason: dto.faultyReason,
    refundedAmount: Number(dto.refundedAmount),
    arrivalDate: dto.arrivalDate,
    notes: dto.notes,
    items: (dto.items ?? []).map((it) => ({
      id: it.id,
      productId: it.productId,
      lineNumber: it.lineNumber,
      quantity: Number(it.quantity),
      unitPrice: Number(it.unitPrice),
      discountPercent: Number(it.discountPercent),
      discountAmount: Number(it.discountAmount),
      taxRate: Number(it.taxRate),
      taxAmount: Number(it.taxAmount),
      description: it.description,
    })),
    payments,
  };
}

function mapPurchase(dto: PurchaseOrderDTO, supplierName: string): Purchase {
  return {
    id: dto.id,
    number: dto.number,
    supplierId: dto.supplierId,
    supplierName,
    date: dto.orderDate,
    status: dto.status,
    total: Number(dto.total),
    creditCardId: dto.creditCardId,
    costUSD: Number(dto.costUSD),
    realCostPEN: Number(dto.realCostPEN),
    projectedProfitPEN: Number(dto.projectedProfitPEN),
    arrivalDate: dto.arrivalDate,
    supplierOrderNumber: dto.supplierOrderNumber,
    faulty: dto.faulty,
    faultyReason: dto.faultyReason,
    refundedAmount: Number(dto.refundedAmount),
  };
}
