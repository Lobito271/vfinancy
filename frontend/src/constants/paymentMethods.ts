const PaymentMethods = {
  Cash: { code: 'cash', label: 'Efectivo' },
  BankTransfer: { code: 'bank_transfer', label: 'Transferencia bancaria' },
  Check: { code: 'check', label: 'Cheque' },
  Card: { code: 'card', label: 'Tarjeta' },
  Credit: { code: 'credit', label: 'Crédito' },
  Other: { code: 'other', label: 'Otro' },
} as const;

export const PaymentMethodOptions = Object.values(PaymentMethods).map((m) => ({
  value: m.code,
  label: m.label,
}));
