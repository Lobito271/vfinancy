import { Fragment, useState, useMemo, useCallback, useEffect, type ReactNode } from 'react';
import {
  ChevronsUpDown,
  ChevronUp,
  ChevronDown,
  Filter,
} from 'lucide-react';
import { EmptyState, ErrorState } from '@/components/feedback';
import { TablePagination } from './TablePagination';
import { Button } from '@/components/button';
import { Drawer } from '@/components/misc';
import { SearchInput } from '@/components/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import { cx } from '@/utils/cx';
import { writeJSON, readJSON } from '@/utils/storage';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
  type RowAction,
} from '@/components/misc';
import {
  type Column,
  type ColumnFilter,
  type SortState,
  type FilterState,
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
  rowActions?: (row: T) => RowAction[] | null;
  rowClassName?: (row: T) => string | undefined;
  state?: Partial<DataTableState>;
  preferencesKey?: string;
  defaultPreferences?: DataTablePreferences;
  toolbarLeft?: ReactNode;
  toolbarRight?: ReactNode;
  columnFilters?: ColumnFilter<T>[];
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

const ROW_HEIGHT = 36;
const TABLE_CHROME_HEIGHT = 320;

function estimatePageSize(): number {
  if (typeof window === 'undefined') return 20;
  return Math.max(8, Math.floor((window.innerHeight - TABLE_CHROME_HEIGHT) / ROW_HEIGHT));
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
  rowActions,
  rowClassName,
  state: externalState,
  preferencesKey,
  defaultPreferences,
  toolbarLeft,
  toolbarRight,
  stickyFirstColumn = DataTableDefaults.stickyFirstColumn,
  columnFilters,
  className,
  ariaLabel,
}: DataTableProps<T>) {
  const initial = useMemo<DataTableState>(() => {
    let base: DataTableState = {
      sort: defaultPreferences?.sort ?? null,
      filters: [],
      search: '',
      page: 1,
      ...externalState,
    };
    if (preferencesKey) {
      const saved = readJSON<DataTablePreferences>(`vfinancy.dt.${preferencesKey}`);
      if (saved) {
        base = {
          ...base,
          sort: saved.sort ?? base.sort,
        };
      }
    }
    return base;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [state, setStateInternal] = useState<DataTableState>(initial);
  const [pageSize, setPageSize] = useState(estimatePageSize);

  useEffect(() => {
    const onResize = () => {
      setPageSize(estimatePageSize());
      setStateInternal((s) => ({ ...s, page: 1 }));
    };
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

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
    for (const f of state.filters) {
      if (f.match) {
        rows = rows.filter((row) => f.match!(row, f.value));
        continue;
      }
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
  }, [data, state.filters, state.sort, columnsById, columns]);

  const total = filteredData.length;
  const pageStart = (state.page - 1) * pageSize;
  const pageRows = useMemo(() => filteredData.slice(pageStart, pageStart + pageSize), [filteredData, pageStart, pageSize]);

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

  const hasColumnFilters = (columnFilters ?? []).length > 0;
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [filterDraft, setFilterDraft] = useState<Record<string, unknown>>({});

  const activeFilterCount = useMemo(
    () => (columnFilters ?? []).filter((f) => {
      const v = filterDraft[f.columnId];
      return v !== undefined && v !== '' && (!(Array.isArray(v)) || v.length > 0);
    }).length,
    [columnFilters, filterDraft],
  );

  const applyFilters = () => {
    const filters: FilterState[] = (columnFilters ?? [])
      .map((cf) => ({
        id: cf.columnId,
        value: filterDraft[cf.columnId] ?? '',
        match: cf.match
          ? (_row: unknown, value: unknown) => cf.match!(_row as T, value)
          : undefined,
      }))
      .filter((f) => f.value !== '' && !(Array.isArray(f.value) && f.value.length === 0));
    update({ filters, page: 1 });
    setFiltersOpen(false);
  };

  const clearFilters = () => {
    setFilterDraft({});
    update({ filters: [], page: 1 });
    setFiltersOpen(false);
  };

  if (error) {
    return <ErrorState title="Error al cargar" description={error.message} onRetry={onRetry} />;
  }

  return (
    <div className={cx('datatable', className)} aria-label={ariaLabel}>
      {(toolbarLeft || toolbarRight || hasColumnFilters) && (
        <div className="datatable-toolbar">
          <div className="datatable-toolbar__left">{toolbarLeft}</div>
          <div className="datatable-toolbar__right">
            {toolbarRight}
            {hasColumnFilters && (
              <Button variant="outline" onClick={() => { setFilterDraft(
                Object.fromEntries((columnFilters ?? []).map((cf) => {
                  const current = state.filters.find((f) => f.id === cf.columnId);
                  return [cf.columnId, current?.value ?? ''];
                })),
              ); setFiltersOpen(true); }}
              >
                <Filter /> Filtros{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
              </Button>
            )}
          </div>
        </div>
      )}

      <Drawer
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        title="Filtros"
        description="Filtra los resultados por columna."
        footer={
          <div className="hstack hstack--sm">
            <Button variant="outline" onClick={clearFilters} disabled={activeFilterCount === 0}>
              Limpiar
            </Button>
            <Button onClick={applyFilters}>Aplicar</Button>
          </div>
        }
      >
        <div className="stack">
          {(columnFilters ?? []).map((cf) => (
            <div className="field" key={cf.columnId}>
              <label className="label">{cf.label}</label>
              {cf.type === 'select' ? (
                <Select
                  items={cf.options ?? []}
                  value={(filterDraft[cf.columnId] as string) ?? ''}
                  onValueChange={(v) => setFilterDraft((d) => ({ ...d, [cf.columnId]: v ?? '' }))}
                >
                  <SelectTrigger aria-label={cf.label}>
                    <SelectValue placeholder="Todos" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">Todos</SelectItem>
                    {(cf.options ?? []).map((o) => (
                      <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <SearchInput
                  value={(filterDraft[cf.columnId] as string) ?? ''}
                  onChange={(e) => setFilterDraft((d) => ({ ...d, [cf.columnId]: e.target.value }))}
                  onClear={() => setFilterDraft((d) => ({ ...d, [cf.columnId]: '' }))}
                  placeholder={`Buscar ${cf.label.toLowerCase()}…`}
                  aria-label={`Filtrar por ${cf.label}`}
                />
              )}
            </div>
          ))}
        </div>
      </Drawer>

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
                    data-active-sort={isSorted ? true : undefined}
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
                          {!isSorted && <ChevronsUpDown strokeWidth={2.5} />}
                          {isSorted && state.sort?.direction === 'asc' && <ChevronUp strokeWidth={2.5} />}
                          {isSorted && state.sort?.direction === 'desc' && <ChevronDown strokeWidth={2.5} />}
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
              rowActions={rowActions}
              rowClassName={rowClassName}
              empty={empty}
            />
          </tbody>
        </table>
      </div>

      {total > 0 && (
        <TablePagination
          page={state.page}
          pageSize={pageSize}
          total={total}
          onPageChange={(p) => update({ page: p })}
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
  rowActions,
  rowClassName,
  empty,
}: {
  loading: boolean;
  pageRows: T[];
  columns: ResolvedColumn<T>[];
  keyField: keyof T;
  onRowClick?: (row: T) => void;
  rowActions?: (row: T) => RowAction[] | null;
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
        const rowEl = (
          <tr
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
                  style={sticky ? { position: 'sticky', left: 0, zIndex: 1 } : undefined}
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
        const actions = rowActions?.(row);
        if (!actions || actions.length === 0) {
          return <Fragment key={id}>{rowEl}</Fragment>;
        }
        return (
          <ContextMenu key={`cm-${id}`}>
            <ContextMenuTrigger render={rowEl} />
            <ContextMenuContent>
              {actions.map((action) => (
                <ContextMenuItem
                  key={action.label}
                  danger={action.danger}
                  disabled={action.disabled}
                  onSelect={action.onSelect}
                >
                  {action.icon && <action.icon className="menu-item-icon" />}
                  {action.label}
                </ContextMenuItem>
              ))}
            </ContextMenuContent>
          </ContextMenu>
        );
      })}
    </>
  );
}
