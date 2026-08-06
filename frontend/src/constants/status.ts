export const SaleStatus = {
  Pending: 'pending',
  Partial: 'partial',
  Paid: 'paid',
  Cancelled: 'cancelled',
} as const;
export type SaleStatusCode = (typeof SaleStatus)[keyof typeof SaleStatus];

export const CustomerStatus = {
  Active: 'active',
  Inactive: 'inactive',
  Blocked: 'blocked',
} as const;
export type CustomerStatusCode = (typeof CustomerStatus)[keyof typeof CustomerStatus];

export const ProductStock = {
  InStock: 'in_stock',
  LowStock: 'low_stock',
  OutOfStock: 'out_of_stock',
  Clearance: 'clearance',
} as const;
export type ProductStockCode = (typeof ProductStock)[keyof typeof ProductStock];

export const InventoryMovementType = {
  Purchase: 'purchase',
  Sale: 'sale',
  Adjustment: 'adjustment',
  Transfer: 'transfer',
  Return: 'return',
  Damage: 'damage',
} as const;
export type InventoryMovementTypeCode = (typeof InventoryMovementType)[keyof typeof InventoryMovementType];

export const PurchaseStatus = {
  Pending: 'pending',
  Paid: 'paid',
  Reconciled: 'reconciled',
  Cancelled: 'cancelled',
} as const;
export type PurchaseStatusCode = (typeof PurchaseStatus)[keyof typeof PurchaseStatus];

export const StockRules = {
  MAX_AGE_DAYS: 25,
} as const;

export const PaymentStatus = {
  Pending: 'pending',
  Partial: 'partial',
  Paid: 'paid',
  Overdue: 'overdue',
} as const;
export type PaymentStatusCode = (typeof PaymentStatus)[keyof typeof PaymentStatus];
