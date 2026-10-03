export const moneyInputCurrencies = ['USD', 'PEN'] as const;
export type Currency = (typeof moneyInputCurrencies)[number];

export function nextMoneyCurrency(current: string): Currency {
  const index = moneyInputCurrencies.indexOf(current as Currency);
  return moneyInputCurrencies[index < 0 ? 0 : (index + 1) % moneyInputCurrencies.length];
}

export function parseAmount(value: string): number | undefined {
  const input = value.trim().replace(/\s/g, '');
  if (!input) return undefined;
  if (!/^-?[\d.,]+$/.test(input)) return undefined;
  if (!/\d/.test(input)) return undefined;

  const commaGroups = input.split(',');
  if (!input.includes('.') && commaGroups.length > 1 && commaGroups[0].replace('-', '').length <= 3 && commaGroups.slice(1).every((part) => part.length === 3)) {
    const amount = Number(input.replace(/,/g, ''));
    return Number.isFinite(amount) ? amount : undefined;
  }

  const decimalIndex = Math.max(input.lastIndexOf('.'), input.lastIndexOf(','));
  const integer = (decimalIndex < 0 ? input : input.slice(0, decimalIndex)).replace(/[.,]/g, '');
  const fraction = decimalIndex < 0 ? '' : input.slice(decimalIndex + 1).replace(/[.,]/g, '');
  const amount = Number(`${integer || '0'}${decimalIndex < 0 ? '' : `.${fraction}`}`);
  return Number.isFinite(amount) ? amount : undefined;
}

export function convertAmount(amount: number, from: Currency, to: Currency, usdToPen: number): number {
  if (from === to) return amount;
  return from === 'USD' ? amount * usdToPen : amount / usdToPen;
}

export function roundMoney(amount: number): number {
  return Math.round((amount + Number.EPSILON) * 100) / 100;
}
