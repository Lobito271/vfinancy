export function cn(...args: Array<string | number | boolean | null | undefined>): string {
  return args.filter(Boolean).join(' ');
}

export function unique<T>(arr: T[]): T[] {
  return Array.from(new Set(arr));
}

export function groupBy<T, K extends string | number>(arr: T[], key: (item: T) => K): Record<K, T[]> {
  return arr.reduce(
    (acc, item) => {
      const k = key(item);
      if (!acc[k]) acc[k] = [] as T[];
      acc[k].push(item);
      return acc;
    },
    {} as Record<K, T[]>,
  );
}

export function sum<T>(arr: T[], selector: (item: T) => number): number {
  return arr.reduce((acc, item) => acc + selector(item), 0);
}

export function sortBy<T>(arr: T[], key: (item: T) => string | number, direction: 'asc' | 'desc' = 'asc'): T[] {
  const out = [...arr];
  out.sort((a, b) => {
    const va = key(a);
    const vb = key(b);
    if (va < vb) return direction === 'asc' ? -1 : 1;
    if (va > vb) return direction === 'asc' ? 1 : -1;
    return 0;
  });
  return out;
}

export function chunk<T>(arr: T[], size: number): T[][] {
  const out: T[][] = [];
  for (let i = 0; i < arr.length; i += size) {
    out.push(arr.slice(i, i + size));
  }
  return out;
}

export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export function pick<T extends object, K extends keyof T>(obj: T, keys: K[]): Pick<T, K> {
  const out = {} as Pick<T, K>;
  for (const k of keys) out[k] = obj[k];
  return out;
}

export function omit<T extends object, K extends keyof T>(obj: T, keys: K[]): Omit<T, K> {
  const out = { ...obj };
  for (const k of keys) delete out[k];
  return out;
}
