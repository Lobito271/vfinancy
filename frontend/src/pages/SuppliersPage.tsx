import { useMemo, useState } from 'react';
import { Truck, Pencil, Trash2, Plus } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { CustomerStatusBadge } from '@/components/badge';
import { SearchInput } from '@/components/input';
import { Button } from '@/components/button';
import { EmptyState } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { RowActions } from '@/components/misc';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { useSuppliers, useDeleteSupplier } from '@/features/suppliers/hooks/useSuppliers';
import { SupplierFormDialog } from '@/features/suppliers/components/SupplierFormDialog';
import type { Supplier } from '@/types/domain';
import { formatCurrency } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const columns: Column<Supplier>[] = [
  {
    id: 'businessName',
    header: 'Proveedor',
    sortable: true,
    sticky: true,
    cell: (row) => (
      <div className="vstack">
        <span className="fw-medium">{row.businessName}</span>
        <span className="subtle-text">
          {row.documentType} · {row.documentNumber}
        </span>
      </div>
    ),
  },
  { id: 'contactName', header: 'Contacto', cell: (row) => row.contactName || '—' },
  { id: 'phone', header: 'Teléfono', cell: (row) => row.phone || '—' },
  { id: 'email', header: 'Correo', cell: (row) => row.email || '—' },
  {
    id: 'currentDebt',
    header: 'Cuentas por pagar',
    align: 'numeric',
    sortable: true,
    cell: (row) => (
      <span className={row.currentDebt > 0 ? 'fw-medium tabular' : 'tabular muted'}>
        {formatCurrency(row.currentDebt)}
      </span>
    ),
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => <CustomerStatusBadge status={row.status === 'active' ? 'active' : 'inactive'} />,
  },
];

export function SuppliersPage() {
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Supplier | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Supplier | null>(null);

  const { data, isLoading, isError, error, refetch } = useSuppliers({ search, status });
  const deleteMutation = useDeleteSupplier();
  const push = useNotificationStore((s) => s.push);

  const suppliers = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalDebt = suppliers.reduce((s, c) => s + c.currentDebt, 0);
  const active = suppliers.filter((s) => s.status === 'active').length;

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const tableColumns = useMemo<Column<Supplier>[]>(() => [
    ...columns,
    {
      id: 'actions',
      header: '',
      width: 72,
      cell: (row) => (
        <RowActions
          actions={[
            {
              label: 'Editar',
              icon: Pencil,
              onSelect: () => {
                setEditing(row);
                setFormOpen(true);
              },
            },
            {
              label: 'Eliminar',
              icon: Trash2,
              danger: true,
              onSelect: () => setDeleteTarget(row),
            },
          ]}
          label={`Acciones de ${row.businessName}`}
        />
      ),
    },
  ], []);

  return (
    <PageContainer>
      <PageHeader
        title="Proveedores"
        subtitle="Gestión de proveedores y cuentas por pagar"
        actions={
          <Button onClick={openCreate}>
            <Plus /> Nuevo proveedor
          </Button>
        }
      />

      <Grid cols={4}>
        <StatCard label="Total proveedores" value={String(total)} icon={Truck} />
        <StatCard label="Proveedores activos" value={String(active)} />
        <StatCard label="Con deuda" value={String(suppliers.filter((s) => s.currentDebt > 0).length)} />
        <StatCard label="Cuentas por pagar" value={formatCurrency(totalDebt)} />
      </Grid>

      <DataTable
        columns={tableColumns}
        data={suppliers}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        preferencesKey="suppliers"
        toolbarLeft={
          <>
            <SearchInput
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onClear={() => setSearch('')}
              placeholder="Buscar proveedor…"
              className="datatable-search"
              aria-label="Buscar proveedor"
            />
            <Select
              items={[
                { value: 'all', label: 'Estado: todos' },
                { value: 'active', label: 'Activos' },
                { value: 'inactive', label: 'Inactivos' },
              ]}
              value={status || 'all'}
              onValueChange={(v) => setStatus(v === 'all' ? '' : (v ?? ''))}
            >
              <SelectTrigger style={{ width: '11rem' }} aria-label="Filtrar por estado">
                <SelectValue placeholder="Estado: todos" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Estado: todos</SelectItem>
                <SelectItem value="active">Activos</SelectItem>
                <SelectItem value="inactive">Inactivos</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        empty={
          <EmptyState
            title="No hay proveedores registrados"
            description="Registra tu primer proveedor para comenzar a registrar compras y cuentas por pagar."
            action={{ label: 'Nuevo proveedor', onClick: openCreate }}
          />
        }
      />

      <SupplierFormDialog open={formOpen} onOpenChange={setFormOpen} supplier={editing} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title="Eliminar proveedor"
        description={`¿Eliminar a ${deleteTarget?.businessName ?? 'este proveedor'}? Esta acción no se puede deshacer.`}
        confirmLabel="Eliminar"
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              push({ title: 'Proveedor eliminado', variant: 'success' });
              setDeleteTarget(null);
            },
            onError: (err: unknown) => {
              push({ title: 'No se pudo eliminar el proveedor', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
              setDeleteTarget(null);
            },
          });
        }}
      />
    </PageContainer>
  );
}
