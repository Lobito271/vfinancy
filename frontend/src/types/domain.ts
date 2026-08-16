export type SaleStatus = 'pending' | 'paid' | 'partial' | 'cancelled';

export type CustomerStatus = 'active' | 'inactive' | 'blocked';

export type DocumentType = 'DNI' | 'RUC' | 'CE' | 'PASSPORT';

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
  cost: number;
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
