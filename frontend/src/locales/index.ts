import es, { type Translations } from './es';

type Path<T> = T extends object
  ? { [K in keyof T]: K extends string ? `${K}.${Path<T[K]>}` | K : never }[keyof T]
  : never;

export type TranslationKey = Path<Translations>;

type Vars = Record<string, string | number>;

export function t(key: TranslationKey, vars?: Vars): string {
  const parts = key.split('.');
  let cur: unknown = es;
  for (const p of parts) {
    if (cur && typeof cur === 'object' && p in (cur as Record<string, unknown>)) {
      cur = (cur as Record<string, unknown>)[p];
    } else {
      return key;
    }
  }
  if (typeof cur !== 'string') return key;
  if (!vars) return cur;
  return cur.replace(/\{(\w+)\}/g, (_, name) => String(vars[name] ?? `{${name}}`));
}

export const locale = 'es-PE';
export const translations = es;
