export type PaymentMethodCode = 'cash' | 'transfer' | 'other';

const PaymentMethods: Record<PaymentMethodCode, string> = {
  cash: 'Efectivo',
  transfer: 'Transferencia',
  other: 'Otro',
};

export const PaymentMethodOptions = Object.entries(PaymentMethods).map(([value, label]) => ({
  value,
  label,
}));
