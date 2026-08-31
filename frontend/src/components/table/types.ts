import type { ReactNode } from 'react';

export interface Column<T> {
  id: string;
  header: string | ReactNode;
  cell: (row: T, info: { rowIndex: number; value: unknown }) => ReactNode;
  accessor?: (row: T) => unknown;
  sortable?: boolean;
  align?: 'left' | 'right' | 'center';
  width?: number | string;
  minWidth?: number;
  sticky?: boolean | 'left' | 'right';
  className?: string;
  headerClassName?: string;
}

type SortDirection = 'asc' | 'desc';

export interface SortState {
  id: string;
  direction: SortDirection;
}

interface FilterState {
  id: string;
  value: unknown;
}

export interface DataTableState {
  sort: SortState | null;
  filters: FilterState[];
  search: string;
  page: number;
  pageSize: number;
}

export interface DataTablePreferences {
  pageSize?: number;
  sort?: SortState | null;
}

export const DataTableDefaults = {
  pageSize: 25,
  pageSizeOptions: [10, 25, 50, 100] as const,
  stickyFirstColumn: true,
};

type CellAlign = NonNullable<Column<unknown>['align']>;

export function getCellAlign(align: CellAlign | undefined): string {
  if (align === 'right') return 'ta-right';
  if (align === 'center') return 'ta-center';
  return '';
}

export function resolveWidth(width: number | string | undefined): string | undefined {
  if (width === undefined) return undefined;
  return typeof width === 'number' ? `${width}px` : width;
}
