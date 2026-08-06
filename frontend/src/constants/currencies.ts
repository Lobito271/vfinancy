export const Currencies = {
  PEN: { code: 'PEN', symbol: 'S/', name: 'Sol peruano', decimals: 2, locale: 'es-PE' },
  USD: { code: 'USD', symbol: '$', name: 'Dólar estadounidense', decimals: 2, locale: 'en-US' },
  EUR: { code: 'EUR', symbol: '€', name: 'Euro', decimals: 2, locale: 'es-ES' },
  MXN: { code: 'MXN', symbol: '$', name: 'Peso mexicano', decimals: 2, locale: 'es-MX' },
  COP: { code: 'COP', symbol: '$', name: 'Peso colombiano', decimals: 2, locale: 'es-CO' },
  CLP: { code: 'CLP', symbol: '$', name: 'Peso chileno', decimals: 0, locale: 'es-CL' },
  BRL: { code: 'BRL', symbol: 'R$', name: 'Real brasileño', decimals: 2, locale: 'pt-BR' },
} as const;

export type CurrencyCode = keyof typeof Currencies;
export const DefaultCurrency: CurrencyCode = 'PEN';

export function getCurrency(code: string) {
  return Currencies[code as CurrencyCode] ?? Currencies[DefaultCurrency];
}
