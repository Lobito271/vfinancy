import { useState, useMemo, useCallback, useEffect, useRef, type ReactNode } from 'react';
import {
  ChevronsUpDown,
  ChevronUp,
  ChevronDown,
  Settings2,
  Download,
  X as XIcon,
} from 'lucide-react';
import { Button } from '@/components/button';
import { Input } from '@/components/input';
import { Checkbox } from '@/components/checkbox';
import { EmptyState, ErrorState } from '@/components/feedback';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/misc';
import { TablePagination } from './TablePagination';
import { downloadCSV } from '@/utils/download';
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

export interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyField: keyof T;
  loading?: boolean;
  error?: Error | null;
  empty?: ReactNode;
  onRetry?: () => void;
  onRowClick?: (row: T) => void;
  onSelectionChange?: (rows: T[]) => void;
  rowClassName?: (row: T) => string | undefined;
  state?: Partial<DataTableState>;
  onStateChange?: (state: DataTableState) => void;
  preferencesKey?: string;
  defaultPreferences?: DataTablePreferences;
  bulkActions?: (rows: T[]) => ReactNode;
  toolbarLeft?: ReactNode;
  toolbarRight?: ReactNode;
  globalSearch?: boolean;
  exportable?: boolean;
  exportFilename?: string;
  virtualized?: boolean;
  rowHeight?: number;
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
  onRowClick,
  onSelectionChange,
  rowClassName,
  state: externalState,
  onStateChange,
  preferencesKey,
  defaultPreferences,
  bulkActions,
  toolbarLeft,
  toolbarRight,
  globalSearch = true,
  exportable = true,
  exportFilename = 'export.csv',
  virtualized: _virtualized = false,
  rowHeight = 48,
  stickyFirstColumn = DataTableDefaults.stickyFirstColumn,
  className,
  ariaLabel,
  onRetry,
}: DataTableProps<T>) {
  const initial = useMemo<DataTableState>(
    () => ({
      sort: null,
      filters: [],
      visibleColumns: columns.filter((c) => c.defaultVisible !== false).map((c) => c.id),
      search: '',
      page: 1,
      pageSize: defaultPreferences?.pageSize ?? DataTableDefaults.pageSize,
      selected: [],
      ...externalState,
    }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const [state, setStateInternal] = useState<DataTableState>(initial);

  useEffect(() => {
    if (preferencesKey) {
      const saved = readJSON<DataTablePreferences>(`vfinancy.dt.${preferencesKey}`);
      if (saved) {
        setStateInternal((s) => ({
          ...s,
          pageSize: saved.pageSize ?? s.pageSize,
          sort: saved.sort ?? s.sort,
          visibleColumns: saved.visibleColumns ?? s.visibleColumns,
        }));
      }
    }
  }, [preferencesKey]);

  useEffect(() => {
    if (preferencesKey) {
      writeJSON(`vfinancy.dt.${preferencesKey}`, {
        pageSize: state.pageSize,
        sort: state.sort,
        visibleColumns: state.visibleColumns,
      });
    }
  }, [preferencesKey, state.pageSize, state.sort, state.visibleColumns]);

  useEffect(() => {
    if (externalState) {
      setStateInternal((s) => ({ ...s, ...externalState }));
    }
  }, [externalState]);

  const update = useCallback(
    (patch: Partial<DataTableState>) => {
      setStateInternal((s) => {
        const next = { ...s, ...patch };
        onStateChange?.(next);
        return next;
      });
    },
    [onStateChange],
  );

  const visibleColumns = useMemo(
    () => columns.filter((c) => state.visibleColumns.includes(c.id)),
    [columns, state.visibleColumns],
  );

  const filteredData = useMemo(() => {
    let rows = data;
    if (state.search && globalSearch) {
      const s = state.search.toLowerCase();
      rows = rows.filter((row) =>
        visibleColumns.some((col) => {
          const v = getValue(row, col);
          return v != null && String(v).toLowerCase().includes(s);
        }),
      );
    }
    for (const f of state.filters) {
      const col = columns.find((c) => c.id === f.id);
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
      const col = visibleColumns.find((c) => c.id === state.sort!.id);
      if (col) {
        const dir = state.sort.direction === 'asc' ? 1 : -1;
        rows = [...rows].sort((a, b) => compare(getValue(a, col), getValue(b, col)) * dir);
      }
    }
    return rows;
  }, [data, state.search, state.filters, state.sort, visibleColumns, columns, globalSearch]);

  const total = filteredData.length;
  const totalPages = Math.max(1, Math.ceil(total / state.pageSize));
  void totalPages;
  const pageStart = (state.page - 1) * state.pageSize;
  const pageRows = useMemo(() => filteredData.slice(pageStart, pageStart + state.pageSize), [filteredData, pageStart, state.pageSize]);

  const allSelected = pageRows.length > 0 && pageRows.every((r) => state.selected.includes(String(r[keyField])));
  const someSelected = pageRows.some((r) => state.selected.includes(String(r[keyField]))) && !allSelected;

  const toggleAll = () => {
    const ids = pageRows.map((r) => String(r[keyField]));
    const next = allSelected ? state.selected.filter((id) => !ids.includes(id)) : Array.from(new Set([...state.selected, ...ids]));
    update({ selected: next });
    onSelectionChange?.(pageRows.filter((r) => next.includes(String(r[keyField]))));
  };

  const toggleRow = (row: T) => {
    const id = String(row[keyField]);
    const next = state.selected.includes(id) ? state.selected.filter((x) => x !== id) : [...state.selected, id];
    update({ selected: next });
    onSelectionChange?.(pageRows.filter((r) => next.includes(String(r[keyField]))));
  };

  const handleSort = (id: string) => {
    const next: SortState | null = !state.sort || state.sort.id !== id
      ? { id, direction: 'asc' }
      : state.sort.direction === 'asc'
        ? { id, direction: 'desc' }
        : null;
    update({ sort: next });
  };

  const handleExport = () => {
    const cols = visibleColumns.filter((c) => c.exportable !== false);
    const rows = filteredData.map((row) => {
      const out: Record<string, unknown> = {};
      for (const c of cols) {
        out[c.id] = c.exportValue ? c.exportValue(row) : getValue(row, c);
      }
      return out;
    });
    downloadCSV(rows, exportFilename, cols.map((c) => ({ key: c.id, header: typeof c.header === 'string' ? c.header : c.id })));
  };

  const parentRef = useRef<HTMLDivElement>(null);
  void rowHeight;
  void _virtualized;

  if (error) {
    return <ErrorState title="Error al cargar" description={error.message} onRetry={onRetry} />;
  }

  const hasSelection = state.selected.length > 0;
  const selectionRows = data.filter((r) => state.selected.includes(String(r[keyField])));

  return (
    <div className={cx('datatable', className)} aria-label={ariaLabel}>
      {(globalSearch || toolbarLeft || toolbarRight || exportable) && (
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
              <Button variant="ghost" size="icon-sm" onClick={() => update({ search: '' })} aria-label="Limpiar búsqueda">
                <XIcon />
              </Button>
            )}
            {hasSelection && (
              <span className="datatable-count">
                {state.selected.length} seleccionado{state.selected.length > 1 ? 's' : ''}
              </span>
            )}
          </div>
          <div className="datatable-toolbar__right">
            {bulkActions && hasSelection && bulkActions(selectionRows)}
            {toolbarRight}
            {exportable && (
              <Button variant="outline" size="sm" onClick={handleExport}>
                <Download /> Exportar CSV
              </Button>
            )}
            <ColumnVisibilityMenu
              columns={columns}
              visible={state.visibleColumns}
              onChange={(v) => update({ visibleColumns: v })}
            />
          </div>
        </div>
      )}

      <div className="datatable-scroll" ref={parentRef}>
        <table className="datatable-table">
          <thead>
            <tr>
              {onSelectionChange && (
                <th style={{ width: '2.5rem' }}>
                  <Checkbox
                    checked={allSelected ? true : someSelected ? 'indeterminate' : false}
                    onCheckedChange={toggleAll}
                    aria-label="Seleccionar todas las filas"
                  />
                </th>
              )}
              {visibleColumns.map((col, i) => {
                const isSorted = state.sort?.id === col.id;
                const sticky = col.sticky === true || (stickyFirstColumn && col.sticky !== false && i === 0);
                return (
                  <th
                    key={col.id}
                    style={{
                      width: resolveWidth(col.width),
                      minWidth: col.minWidth,
                      ...(sticky ? { position: 'sticky', left: 0, zIndex: 1 } : {}),
                    }}
                    className={cx(
                      getCellAlign(col.align),
                      col.sortable && 'sortable',
                      sticky && 'sticky-cell',
                      col.headerClassName,
                    )}
                    onClick={() => col.sortable && handleSort(col.id)}
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
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <tr key={`skel-${i}`}>
                  {onSelectionChange && <td><div className="skel-check" /></td>}
                  {visibleColumns.map((col) => (
                    <td key={col.id}>
                      <div className="skel-cell" />
                    </td>
                  ))}
                </tr>
              ))
            ) : pageRows.length === 0 ? (
              <tr>
                <td colSpan={visibleColumns.length + (onSelectionChange ? 1 : 0)} style={{ padding: 0 }}>
                  {empty ?? <EmptyState title="Sin resultados" description="Ajusta los filtros o la búsqueda." />}
                </td>
              </tr>
            ) : (
              pageRows.map((row, rowIndex) => {
                const id = String(row[keyField]);
                const isSelected = state.selected.includes(id);
                return (
                  <tr
                    key={id}
                    onClick={onRowClick ? () => onRowClick(row) : undefined}
                    className={cx(
                      onRowClick && 'clickable',
                      isSelected && 'selected',
                      rowClassName?.(row),
                    )}
                  >
                    {onSelectionChange && (
                      <td onClick={(e) => e.stopPropagation()}>
                        <Checkbox
                          checked={isSelected}
                          onCheckedChange={() => toggleRow(row)}
                          aria-label="Seleccionar fila"
                        />
                      </td>
                    )}
                    {visibleColumns.map((col, colIndex) => {
                      const sticky = col.sticky === true || (stickyFirstColumn && col.sticky !== false && colIndex === 0);
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
              })
            )}
          </tbody>
        </table>
      </div>

      {total > 0 && (
        <TablePagination
          page={state.page}
          pageSize={state.pageSize}
          total={total}
          onPageChange={(p) => update({ page: p })}
          onPageSizeChange={(n) => update({ pageSize: n, page: 1 })}
        />
      )}
    </div>
  );
}

function ColumnVisibilityMenu<T>({
  columns,
  visible,
  onChange,
}: {
  columns: Column<T>[];
  visible: string[];
  onChange: (v: string[]) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" aria-label="Visibilidad de columnas">
          <Settings2 /> Columnas
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" style={{ width: '14rem' }}>
        <DropdownMenuLabel>Mostrar columnas</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {columns.map((col) => (
          <DropdownMenuCheckboxItem
            key={col.id}
            checked={visible.includes(col.id)}
            onCheckedChange={(checked) => {
              const next = checked ? Array.from(new Set([...visible, col.id])) : visible.filter((id) => id !== col.id);
              onChange(next);
            }}
          >
            {typeof col.header === 'string' ? col.header : col.id}
          </DropdownMenuCheckboxItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
