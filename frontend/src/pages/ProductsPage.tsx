import { useMemo, useState } from 'react';
import { Package, Pencil, Trash2, Plus } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { Input } from '@/components/input';
import { Button } from '@/components/button';
import { EmptyState } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { useProducts, useDeleteProduct } from '@/features/products/hooks/useProducts';
import { ProductFormDialog } from '@/features/products/components/ProductFormDialog';
import type { Product } from '@/types/domain';
import { formatCurrency, formatPercent } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const columns: Column<Product>[] = [
  {
    id: 'sku',
    header: 'SKU',
    sortable: true,
    sticky: true,
    cell: (row) => <span className="fw-medium tabular">{row.sku}</span>,
  },
  {
    id: 'description',
    header: 'Descripción',
    sortable: true,
    cell: (row) => (
      <div className="vstack">
        <span>{row.description}</span>
        {row.barcode && <span className="subtle-text">Código: {row.barcode}</span>}
      </div>
    ),
  },
  { id: 'category', header: 'Categoría', cell: (row) => row.category || '—' },
  { id: 'brand', header: 'Marca', cell: (row) => row.brand || '—' },
  { id: 'unit', header: 'Unidad', cell: (row) => row.unit || '—' },
  {
    id: 'costUSD',
    header: 'Costo (USD)',
    align: 'right',
    sortable: true,
    cell: (row) => <span className="tabular-nums">{formatCurrency(row.costUSD, 'USD')}</span>,
  },
  {
    id: 'salePrice',
    header: 'Precio venta',
    align: 'right',
    sortable: true,
    cell: (row) => <span className="tabular">{formatCurrency(row.salePrice)}</span>,
  },
  {
    id: 'margin',
    header: 'Margen',
    align: 'right',
    cell: (row) => (
      <span className="tabular-nums muted">
        {row.costUSD > 0 ? formatPercent((row.salePrice - row.costUSD) / row.costUSD) : '—'}
      </span>
    ),
  },
  {
    id: 'isActive',
    header: 'Estado',
    cell: (row) =>
      row.isActive ? <Badge variant="success">Activo</Badge> : <Badge variant="muted">Inactivo</Badge>,
  },
];

export function ProductsPage() {
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Product | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Product | null>(null);

  const { data, isLoading, isError, error, refetch } = useProducts({ search, status });
  const deleteMutation = useDeleteProduct();
  const push = useNotificationStore((s) => s.push);

  const products = data?.items ?? [];
  const total = data?.total ?? 0;
  const active = products.filter((p) => p.isActive).length;
  const avgMargin = products.length
    ? products.reduce((s, p) => s + (p.salePrice - p.costUSD) / Math.max(p.costUSD, 0.01), 0) / products.length
    : 0;

  const tableColumns = useMemo<Column<Product>[]>(() => [
    ...columns,
    {
      id: 'actions',
      header: '',
      width: 88,
      exportable: false,
      cell: (row) => (
        <div className="row-actions">
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Editar ${row.sku}`}
            onClick={() => {
              setEditing(row);
              setFormOpen(true);
            }}
          >
            <Pencil />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Eliminar ${row.sku}`}
            onClick={() => setDeleteTarget(row)}
          >
            <Trash2 />
          </Button>
        </div>
      ),
    },
  ], []);

  return (
    <PageContainer>
      <PageHeader
        title="Productos"
        subtitle="Catálogo de productos y servicios"
        actions={
          <div className="hstack hstack--sm">
            <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Buscar producto…" style={{ width: "16rem" }} aria-label="Buscar producto" />
            <Button
              onClick={() => {
                setEditing(null);
                setFormOpen(true);
              }}
            >
              <Plus /> Nuevo producto
            </Button>
          </div>
        }
      />

      <Grid cols={4}>
        <StatCard label="Total productos" value={String(total)} icon={Package} />
        <StatCard label="Productos activos" value={String(active)} />
        <StatCard label="Inactivos" value={String(total - active)} />
        <StatCard label="Margen promedio" value={formatPercent(avgMargin)} />
      </Grid>

      <DataTable
        columns={tableColumns}
        data={products}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        globalSearch={false}
        exportFilename="productos.csv"
        toolbarLeft={
          <Select value={status} onValueChange={(v) => setStatus(v === 'all' ? '' : v)}>
            <SelectTrigger style={{ width: "11rem" }} aria-label="Filtrar por estado">
              <SelectValue placeholder="Estado: todos" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Estado: todos</SelectItem>
              <SelectItem value="active">Activos</SelectItem>
              <SelectItem value="inactive">Inactivos</SelectItem>
            </SelectContent>
          </Select>
        }
        empty={
          <EmptyState
            title="No hay productos registrados"
            description="Cuando se creen productos, aparecerán aquí con su costo, precio y estado."
          />
        }
      />

      <ProductFormDialog open={formOpen} onOpenChange={setFormOpen} product={editing} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title="Eliminar producto"
        description={`¿Eliminar el producto ${deleteTarget?.sku ?? ''} — ${deleteTarget?.description ?? 'este producto'}? Esta acción no se puede deshacer.`}
        confirmLabel="Eliminar"
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              push({ title: 'Producto eliminado', variant: 'success' });
              setDeleteTarget(null);
            },
            onError: (err: unknown) => {
              push({ title: 'No se pudo eliminar el producto', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
              setDeleteTarget(null);
            },
          });
        }}
      />
    </PageContainer>
  );
}
