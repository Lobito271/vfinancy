export type SaleStatus = 'paid' | 'pending' | 'partial' | 'cancelled';

export type CustomerStatus = 'active' | 'inactive' | 'blocked';

export type DocumentType = '' | 'DNI' | 'RUC';

export interface Customer {
  id: string;
  documentType: DocumentType;
  documentNumber: string;
  businessName: string;
  phone?: string;
  email?: string;
  address?: string;
  currentDebt: number;
  status: CustomerStatus;
}

export interface Product {
  id: string;
  sku: string;
  description: string;
  unitCode: string;
  costUsd: number;
  salePrice: number;
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
  dueDate: string;
  status: SaleStatus;
  saleType: string;
  total: number;
  costTotal: number;
  profit: number;
  paid: number;
  balance: number;
}

export interface Purchase {
  id: string;
  number: string;
  date: string;
  status: string;
  currencyCode: string;
  orderType: string;
  customerId: string;
  creditCardId: string;
  costUsd: number;
  salePricePen: number;
  realCostPen: number;
  refundAmount: number;
  faulty: boolean;
  faultyReason: string;
  arrivalDate: string;
  notes: string;
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
