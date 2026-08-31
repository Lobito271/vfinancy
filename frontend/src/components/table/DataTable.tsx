import { useState, useMemo, useCallback, type ReactNode } from 'react';
import {
  ChevronsUpDown,
  ChevronUp,
  ChevronDown,
  X as XIcon,
} from 'lucide-react';
import { Input } from '@/components/input';
import { EmptyState, ErrorState } from '@/components/feedback';
import { TablePagination } from './TablePagination';
import { cx } from '@/utils/cx';
import { writeJSON, readJSON } from '@/utils/storage';
import {
  type Column,
  type SortState,
  type DataTableState,
  type DataTablePreferences,
  DataTableDefaults,
  getCellAlign,
  resolveWidth,
} from './types';

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyField: keyof T;
  loading?: boolean;
  error?: Error | null;
  empty?: ReactNode;
  onRetry?: () => void;
  onRowClick?: (row: T) => void;
  rowClassName?: (row: T) => string | undefined;
  state?: Partial<DataTableState>;
  preferencesKey?: string;
  defaultPreferences?: DataTablePreferences;
  toolbarLeft?: ReactNode;
  toolbarRight?: ReactNode;
  globalSearch?: boolean;
  stickyFirstColumn?: boolean;
  className?: string;
  ariaLabel?: string;
}

function getValue<T>(row: T, col: Column<T>): unknown {
  if (col.accessor) return col.accessor(row);
  return (row as Record<string, unknown>)[col.id];
}

function compare(a: unknown, b: unknown): number {
  if (a == null && b == null) return 0;
  if (a == null) return -1;
  if (b == null) return 1;
  if (typeof a === 'number' && typeof b === 'number') return a - b;
  if (a instanceof Date && b instanceof Date) return a.getTime() - b.getTime();
  return String(a).localeCompare(String(b), 'es', { numeric: true, sensitivity: 'base' });
}

export function DataTable<T>({
  columns,
  data,
  keyField,
  loading = false,
  error = null,
  empty,
  onRetry,
  onRowClick,
  rowClassName,
  state: externalState,
  preferencesKey,
  defaultPreferences,
  toolbarLeft,
  toolbarRight,
  globalSearch = true,
  stickyFirstColumn = DataTableDefaults.stickyFirstColumn,
  className,
  ariaLabel,
}: DataTableProps<T>) {
  const initial = useMemo<DataTableState>(() => {
    let base: DataTableState = {
      sort: defaultPreferences?.sort ?? null,
      filters: [],
      search: '',
      page: 1,
      pageSize: defaultPreferences?.pageSize ?? DataTableDefaults.pageSize,
      ...externalState,
    };
    if (preferencesKey) {
      const saved = readJSON<DataTablePreferences>(`vfinancy.dt.${preferencesKey}`);
      if (saved) {
        base = {
          ...base,
          pageSize: saved.pageSize ?? base.pageSize,
          sort: saved.sort ?? base.sort,
        };
      }
    }
    return base;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [state, setStateInternal] = useState<DataTableState>(initial);

  const update = useCallback(
    (patch: Partial<DataTableState>) => {
      setStateInternal((s) => ({ ...s, ...patch }));
    },
    [],
  );

  const columnsById = useMemo(() => {
    const map = new Map<string, Column<T>>();
    for (const c of columns) map.set(c.id, c);
    return map;
  }, [columns]);

  const filteredData = useMemo(() => {
    let rows = data;
    if (state.search && globalSearch) {
      const s = state.search.toLowerCase();
      rows = rows.filter((row) =>
        columns.some((col) => {
          const v = getValue(row, col);
          return v != null && String(v).toLowerCase().includes(s);
        }),
      );
    }
    for (const f of state.filters) {
      const col = columnsById.get(f.id);
      if (!col) continue;
      rows = rows.filter((row) => {
        const v = getValue(row, col);
        if (v == null) return false;
        if (Array.isArray(f.value) && f.value.length > 0) {
          return (f.value as unknown[]).includes(v);
        }
        if (typeof f.value === 'string' && f.value) {
          return String(v).toLowerCase().includes(String(f.value).toLowerCase());
        }
        return true;
      });
    }
    if (state.sort) {
      const col = columnsById.get(state.sort.id);
      if (col) {
        const dir = state.sort.direction === 'asc' ? 1 : -1;
        rows = [...rows].sort((a, b) => compare(getValue(a, col), getValue(b, col)) * dir);
      }
    }
    return rows;
  }, [data, state.search, state.filters, state.sort, columnsById, columns, globalSearch]);

  const total = filteredData.length;
  const pageStart = (state.page - 1) * state.pageSize;
  const pageRows = useMemo(() => filteredData.slice(pageStart, pageStart + state.pageSize), [filteredData, pageStart, state.pageSize]);

  const handleSort = useCallback(
    (id: string) => {
      const next: SortState | null = !state.sort || state.sort.id !== id
        ? { id, direction: 'asc' }
        : state.sort.direction === 'asc'
          ? { id, direction: 'desc' }
          : null;
      update({ sort: next });
      if (preferencesKey) {
        const saved = readJSON<DataTablePreferences>(`vfinancy.dt.${preferencesKey}`);
        writeJSON(`vfinancy.dt.${preferencesKey}`, { ...saved, sort: next });
      }
    },
    [state.sort, update, preferencesKey],
  );

  if (error) {
    return <ErrorState title="Error al cargar" description={error.message} onRetry={onRetry} />;
  }

  return (
    <div className={cx('datatable', className)} aria-label={ariaLabel}>
      {(globalSearch || toolbarLeft || toolbarRight) && (
        <div className="datatable-toolbar">
          <div className="datatable-toolbar__left">
            {toolbarLeft}
            {globalSearch && (
              <Input
                value={state.search}
                onChange={(e) => update({ search: e.target.value, page: 1 })}
                placeholder="Buscar…"
                className="datatable-search"
                aria-label="Buscar en la tabla"
              />
            )}
            {state.search && (
              <button
                type="button"
                className="btn btn--ghost btn--icon-sm"
                onClick={() => update({ search: '' })}
                aria-label="Limpiar búsqueda"
              >
                <XIcon />
              </button>
            )}
          </div>
          <div className="datatable-toolbar__right">{toolbarRight}</div>
        </div>
      )}

      <div className="datatable-scroll">
        <table className="datatable-table">
          <thead>
            <tr>
              {visibleHeaders(columns, stickyFirstColumn).map(({ col, sticky }) => {
                const isSorted = state.sort?.id === col.id;
                return (
                  <th
                    key={col.id}
                    style={{
                      width: resolveWidth(col.width),
                      minWidth: col.minWidth,
                      ...(sticky ? { position: 'sticky', left: 0, zIndex: 3 } : {}),
                    }}
                    className={cx(
                      getCellAlign(col.align),
                      col.sortable && 'sortable',
                      sticky && 'sticky-cell',
                      col.headerClassName,
                    )}
                    tabIndex={col.sortable ? 0 : undefined}
                    role={col.sortable ? 'button' : undefined}
                    onClick={() => col.sortable && handleSort(col.id)}
                    onKeyDown={(e) => {
                      if (col.sortable && (e.key === 'Enter' || e.key === ' ')) {
                        e.preventDefault();
                        handleSort(col.id);
                      }
                    }}
                  >
                    <span className="th-head">
                      {col.header}
                      {col.sortable && (
                        <span className="th-sort-icon">
                          {!isSorted && <ChevronsUpDown />}
                          {isSorted && state.sort?.direction === 'asc' && <ChevronUp />}
                          {isSorted && state.sort?.direction === 'desc' && <ChevronDown />}
                        </span>
                      )}
                    </span>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            <DataTableBody
              loading={loading}
              pageRows={pageRows}
              columns={visibleHeaders(columns, stickyFirstColumn)}
              keyField={keyField}
              onRowClick={onRowClick}
              rowClassName={rowClassName}
              empty={empty}
            />
          </tbody>
        </table>
      </div>

      {total > 0 && (
        <TablePagination
          page={state.page}
          pageSize={state.pageSize}
          total={total}
          onPageChange={(p) => update({ page: p })}
          onPageSizeChange={(n) => {
            update({ pageSize: n, page: 1 });
            if (preferencesKey) {
              const saved = readJSON<DataTablePreferences>(`vfinancy.dt.${preferencesKey}`);
              writeJSON(`vfinancy.dt.${preferencesKey}`, { ...saved, pageSize: n });
            }
          }}
        />
      )}
    </div>
  );
}

interface ResolvedColumn<T> {
  col: Column<T>;
  sticky: boolean;
}

function visibleHeaders<T>(columns: Column<T>[], stickyFirstColumn: boolean): ResolvedColumn<T>[] {
  return columns.map((col, i) => ({
    col,
    sticky: col.sticky === true || (stickyFirstColumn && col.sticky !== false && i === 0),
  }));
}

function DataTableBody<T>({
  loading,
  pageRows,
  columns,
  keyField,
  onRowClick,
  rowClassName,
  empty,
}: {
  loading: boolean;
  pageRows: T[];
  columns: ResolvedColumn<T>[];
  keyField: keyof T;
  onRowClick?: (row: T) => void;
  rowClassName?: (row: T) => string | undefined;
  empty?: ReactNode;
}) {
  if (loading) {
    return (
      <>
        {Array.from({ length: 5 }).map((_, i) => (
          <tr key={`skel-${i}`}>
            {columns.map(({ col }) => (
              <td key={col.id}><div className="skel-cell" /></td>
            ))}
          </tr>
        ))}
      </>
    );
  }

  if (pageRows.length === 0) {
    return (
      <tr>
        <td colSpan={columns.length} style={{ padding: 0 }}>
          {empty ?? <EmptyState title="Sin resultados" description="Ajusta los filtros o la búsqueda." />}
        </td>
      </tr>
    );
  }

  return (
    <>
      {pageRows.map((row, rowIndex) => {
        const id = String(row[keyField]);
        return (
          <tr
            key={id}
            onClick={onRowClick ? () => onRowClick(row) : undefined}
            onKeyDown={onRowClick ? (e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onRowClick(row);
              }
            } : undefined}
            tabIndex={onRowClick ? 0 : undefined}
            role={onRowClick ? 'button' : undefined}
            className={cx(
              onRowClick && 'clickable',
              rowClassName?.(row),
            )}
          >
            {columns.map(({ col, sticky }) => {
              return (
                <td
                  key={col.id}
                  style={sticky ? { position: 'sticky', left: 0, zIndex: 1, background: 'inherit' } : undefined}
                  className={cx(
                    getCellAlign(col.align),
                    sticky && 'sticky-cell',
                    col.className,
                  )}
                >
                  {col.cell(row, { rowIndex, value: getValue(row, col) })}
                </td>
              );
            })}
          </tr>
        );
      })}
    </>
  );
}
