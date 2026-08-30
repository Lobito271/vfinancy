import type { ReactNode } from 'react';

export interface Column<T> {
  id: string;
  header: string | ReactNode;
  cell: (row: T, info: { rowIndex: number; value: unknown }) => ReactNode;
  accessor?: (row: T) => unknown;
  sortable?: boolean;
  filterable?: boolean;
  align?: 'left' | 'right' | 'center';
  width?: number | string;
  minWidth?: number;
  resizable?: boolean;
  sticky?: boolean | 'left' | 'right';
  defaultVisible?: boolean;
  exportValue?: (row: T) => string | number;
  className?: string;
  headerClassName?: string;
  exportable?: boolean;
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
  visibleColumns: string[];
  search: string;
  page: number;
  pageSize: number;
  selected: string[];
}

export interface DataTablePreferences {
  visibleColumns?: string[];
  pageSize?: number;
  sort?: SortState | null;
  columnWidths?: Record<string, number>;
}

export const DataTableDefaults = {
  pageSize: 25,
  pageSizeOptions: [10, 25, 50, 100] as const,
  stickyHeader: true,
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
