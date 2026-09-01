export type SaleStatus = 'pending' | 'paid' | 'partial' | 'cancelled';

export type CustomerStatus = 'active' | 'inactive' | 'blocked';

type DocumentType = 'DNI' | 'RUC' | 'CE' | 'PASSPORT';

export interface Customer {
  id: string;
  documentType: DocumentType;
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
  address?: string;
  creditLimit: number;
  currentDebt: number;
  status: CustomerStatus;
  totalPurchases: number;
  createdAt: string;
}

export interface Supplier {
  id: string;
  documentType?: DocumentType;
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
  address?: string;
  paymentTermDays?: number;
  currentDebt: number;
  status: 'active' | 'inactive';
}

export interface Product {
  id: string;
  sku: string;
  barcode?: string;
  description: string;
  categoryId?: string;
  brandId?: string;
  category: string;
  brand: string;
  unit: string;
  unitId?: string;
  taxId?: string;
  costUSD: number;
  salePrice: number;
  minStock: number;
  maxStock: number;
  isActive: boolean;
}

export interface InventoryItem {
  id: string;
  productId: string;
  productSku: string;
  productDescription: string;
  warehouse: string;
  quantity: number;
  unitCost: number;
  currencyCode: string;
  arrivalDate: string;
  maxSaleDate: string;
  ageDays: number;
  daysRemaining: number;
  isClearance: boolean;
  status: string;
}

export interface Sale {
  id: string;
  number: string;
  customerId: string;
  customerName: string;
  date: string;
  status: SaleStatus;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  cost: number;
  profit: number;
}

export interface Purchase {
  id: string;
  number: string;
  supplierId: string;
  supplierName: string;
  date: string;
  status: string;
  total: number;
  creditCardId: string;
  costUSD: number;
  realCostPEN: number;
  projectedProfitPEN: number;
  arrivalDate: string;
  supplierOrderNumber: string;
  faulty: boolean;
  faultyReason: string;
  refundedAmount: number;
}

export interface CustomerOrderPayment {
  id: string;
  purchaseOrderId: string;
  number: string;
  paymentDate: string;
  amount: number;
  method: string;
  currencyCode: string;
  exchangeRate: number;
  reference: string;
  notes: string;
  status: 'active' | 'refunded';
  refundedAmount: number;
  refundedAt: string;
  refundReason: string;
}

export interface CustomerOrderItem {
  id: string;
  productId: string;
  lineNumber: number;
  quantity: number;
  unitPrice: number;
  discountPercent: number;
  discountAmount: number;
  taxRate: number;
  taxAmount: number;
  description: string;
}

export interface CustomerOrder {
  id: string;
  number: string;
  supplierId: string;
  customerId: string;
  customerName: string;
  creditCardId: string;
  orderType: string;
  date: string;
  status: string;
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
  paid: number;
  costUSD: number;
  salePricePEN: number;
  realCostPEN: number;
  projectedProfitPEN: number;
  anticipo: number;
  anticipoDate: string;
  porCobrar: number;
  supplierOrderNumber: string;
  faulty: boolean;
  faultyReason: string;
  refundedAmount: number;
  arrivalDate: string;
  notes: string;
  items: CustomerOrderItem[];
  payments: CustomerOrderPayment[];
}

export interface ActivityItem {
  id: string;
  type: 'sale' | 'purchase' | 'payment' | 'customer' | 'product';
  description: string;
  amount?: number;
  date: string;
  user: string;
}

export interface ChartPoint {
  label: string;
  value: number;
}
