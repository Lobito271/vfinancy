export const Taxes = {
  IGV: {
    code: 'IGV',
    name: 'IGV (Impuesto General a las Ventas)',
    shortName: 'IGV',
    rate: 0.18,
    country: 'PE',
    inclusive: false,
  },
  ISR: {
    code: 'ISR',
    name: 'Impuesto a la Renta',
    shortName: 'Renta',
    rate: 0.295,
    country: 'PE',
    inclusive: false,
  },
  IVAP: {
    code: 'IVAP',
    name: 'IVAP (Impuesto a la Venta de Arroz Pilado)',
    shortName: 'IVAP',
    rate: 0.04,
    country: 'PE',
    inclusive: false,
  },
  EXEMPT: {
    code: 'EXEMPT',
    name: 'Exonerado',
    shortName: 'Exonerado',
    rate: 0,
    country: '*',
    inclusive: false,
  },
} as const;

export type TaxCode = keyof typeof Taxes;
export const DefaultTax: TaxCode = 'IGV';

export function getTax(code: string) {
  return Taxes[code as TaxCode] ?? Taxes[DefaultTax];
}

export function calculateTax(base: number, code: TaxCode, inclusive = false): { base: number; tax: number; total: number } {
  const t = getTax(code);
  if (t.rate === 0) return { base, tax: 0, total: base };
  if (inclusive) {
    const tax = (base * t.rate) / (1 + t.rate);
    return { base: base - tax, tax, total: base };
  }
  const tax = base * t.rate;
  return { base, tax, total: base + tax };
}
