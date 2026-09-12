import type { PageResult } from './wails-types';

const MAX_PAGES = 100;

export async function fetchAllPages<T>(
  fetchPage: (page: number, pageSize: number) => Promise<PageResult<T>>,
  pageSize = 200,
): Promise<T[]> {
  const all: T[] = [];
  for (let page = 1; page <= MAX_PAGES; page += 1) {
    const res = await fetchPage(page, pageSize);
    const items = (res.items ?? []) as T[];
    all.push(...items);
    if (items.length === 0 || all.length >= res.total) break;
  }
  return all;
}
