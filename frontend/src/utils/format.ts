import { Currencies, type CurrencyCode } from '@/constants/currencies';

type LanguageCode = 'es-PE' | 'es-MX' | 'es-CO' | 'es-AR' | 'en-US' | 'pt-BR';
const locale: LanguageCode = 'es-PE';

const currencyFormatters = new Map<string, Intl.NumberFormat>();

function getCurrencyFormatter(currency: string) {
  let f = currencyFormatters.get(currency);
  if (!f) {
    const cur = Currencies[currency as CurrencyCode] ?? Currencies.PEN;
    f = new Intl.NumberFormat(cur.locale, {
      style: 'currency',
      currency: cur.code,
      minimumFractionDigits: cur.decimals,
      maximumFractionDigits: cur.decimals,
    });
    currencyFormatters.set(currency, f);
  }
  return f;
}

export function formatCurrency(value: number, currency: string = 'PEN'): string {
  return getCurrencyFormatter(currency).format(value);
}

const dateFormatter = new Intl.DateTimeFormat(locale, {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});

// Matches a plain "YYYY-MM-DD" date string with no time component.
const isoDateOnly = /^\d{4}-\d{2}-\d{2}$/;

function toDate(value: string | Date | null | undefined): Date | null {
  if (value === null || value === undefined) return null;
  // Date-only strings are treated as LOCAL calendar dates: JS parses
  // "YYYY-MM-DD" as UTC midnight, which shifts the displayed date one
  // day back in negative-UTC-offset zones (e.g. UTC-5). Build the Date
  // from the local components instead so the selected day is preserved.
  if (typeof value === 'string' && isoDateOnly.test(value)) {
    const [y, m, d] = value.split('-').map(Number);
    const date = new Date(y, m - 1, d);
    return Number.isNaN(date.getTime()) ? null : date;
  }
  const d = typeof value === 'string' ? new Date(value) : value;
  return Number.isNaN(d.getTime()) ? null : d;
}

export function formatDate(value: string | Date | null | undefined, fallback = '—'): string {
  const d = toDate(value);
  return d ? dateFormatter.format(d) : fallback;
}

export function formatNumber(value: number, decimals = 0): string {
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value);
}

export function daysBetween(from: Date | string, to: Date | string): number {
  const a = toDate(from);
  const b = toDate(to);
  if (!a || !b) return 0;
  return Math.floor((b.getTime() - a.getTime()) / (1000 * 60 * 60 * 24));
}
