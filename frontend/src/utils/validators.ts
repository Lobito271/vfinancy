const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PHONE_PE_RE = /^9\d{8}$/;
const RUC_RE = /^(10|20)\d{9}$/;
const DNI_RE = /^\d{8}$/;
const URL_RE = /^https?:\/\/[\w.-]+/;

export const isPresent = (v: unknown): boolean => v !== null && v !== undefined && v !== '';
export const isBlank = (v: unknown): boolean => !isPresent(v);

export const isEmail = (v: string): boolean => EMAIL_RE.test(v);
export const isPhonePE = (v: string): boolean => PHONE_PE_RE.test(v);
export const isDNI = (v: string): boolean => DNI_RE.test(v);
export const isRUC = (v: string): boolean => RUC_RE.test(v);
export const isUrl = (v: string): boolean => URL_RE.test(v);

export const isPositive = (v: number): boolean => Number.isFinite(v) && v > 0;
export const isNonNegative = (v: number): boolean => Number.isFinite(v) && v >= 0;
export const isInteger = (v: number): boolean => Number.isInteger(v);

export const minLength = (n: number) => (v: string) => v.length >= n;
export const maxLength = (n: number) => (v: string) => v.length <= n;
export const between = (min: number, max: number) => (v: number) => v >= min && v <= max;

export const required = (v: unknown): boolean => isPresent(v);
export const requiredMsg = (msg = 'Requerido') => (v: unknown) => (isPresent(v) ? undefined : msg);

export function matchField(refName: string) {
  return (v: string, all: Record<string, unknown>) => {
    const ref = all[refName];
    return v === ref ? undefined : 'Los valores no coinciden';
  };
}

export const isOneOf = (allowed: readonly string[]) => (v: string) => allowed.includes(v);

export const validators = {
  required,
  isPresent,
  isBlank,
  isEmail,
  isPhonePE,
  isDNI,
  isRUC,
  isUrl,
  isPositive,
  isNonNegative,
  isInteger,
  minLength,
  maxLength,
  between,
  matchField,
  isOneOf,
} as const;
