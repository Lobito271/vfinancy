import { Currencies, type CurrencyCode } from '@/constants/currencies';
import { Languages, type LanguageCode, DefaultLanguage } from '@/constants/languages';

const locale: LanguageCode = DefaultLanguage;

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

export function formatMoney(value: number, currency: string = 'PEN'): string {
  return formatCurrency(value, currency);
}

const dateFormatter = new Intl.DateTimeFormat(locale, {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});

const dateTimeFormatter = new Intl.DateTimeFormat(locale, {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
});

const longDateFormatter = new Intl.DateTimeFormat(locale, {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

const timeFormatter = new Intl.DateTimeFormat(locale, {
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
});

function toDate(value: string | Date | null | undefined): Date | null {
  if (value === null || value === undefined) return null;
  const d = typeof value === 'string' ? new Date(value) : value;
  return Number.isNaN(d.getTime()) ? null : d;
}

export function formatDate(value: string | Date | null | undefined, fallback = '—'): string {
  const d = toDate(value);
  return d ? dateFormatter.format(d) : fallback;
}

export function formatDateTime(value: string | Date | null | undefined, fallback = '—'): string {
  const d = toDate(value);
  return d ? dateTimeFormatter.format(d) : fallback;
}

export function formatLongDate(value: string | Date | null | undefined, fallback = '—'): string {
  const d = toDate(value);
  return d ? longDateFormatter.format(d) : fallback;
}

export function formatTime(value: string | Date | null | undefined, fallback = '—'): string {
  const d = toDate(value);
  return d ? timeFormatter.format(d) : fallback;
}

export function formatNumber(value: number, decimals = 0): string {
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value);
}

export function formatPercent(value: number, decimals = 1): string {
  return new Intl.NumberFormat(locale, {
    style: 'percent',
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

export function addDays(value: Date | string, days: number): Date {
  const d = toDate(value);
  if (!d) return new Date();
  d.setDate(d.getDate() + days);
  return d;
}

export function formatRelative(value: string | Date | null | undefined): string {
  const d = toDate(value);
  if (!d) return '—';
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const sec = Math.floor(diffMs / 1000);
  const min = Math.floor(sec / 60);
  const hr = Math.floor(min / 60);
  const day = Math.floor(hr / 24);
  if (sec < 60) return 'hace un momento';
  if (min < 60) return `hace ${min} min`;
  if (hr < 24) return `hace ${hr} h`;
  if (day < 7) return `hace ${day} d`;
  return formatDate(d);
}

void Languages;
