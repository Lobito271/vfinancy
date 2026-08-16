import { useMemo, useState } from 'react';
import { Truck } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { CustomerStatusBadge } from '@/components/badge';
import { Input } from '@/components/input';
import { Button } from '@/components/button';
import { EmptyState } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { Can } from '@/components/auth';
import { Icons } from '@/design-system/icons';
import { Permissions } from '@/constants/permissions';
import { usePermission } from '@/hooks/usePermission';
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
      <div className="flex flex-col">
        <span className="font-medium">{row.businessName}</span>
        <span className="text-xs text-muted-foreground">
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
    align: 'right',
    sortable: true,
    cell: (row) => (
      <span className={row.currentDebt > 0 ? 'font-medium tabular-nums' : 'tabular-nums text-muted-foreground'}>
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
  const canEdit = usePermission(Permissions.Suppliers.Edit);
  const canDelete = usePermission(Permissions.Suppliers.Delete);

  const suppliers = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalDebt = suppliers.reduce((s, c) => s + c.currentDebt, 0);
  const active = suppliers.filter((s) => s.status === 'active').length;

  const tableColumns = useMemo<Column<Supplier>[]>(() => {
    if (!canEdit && !canDelete) return columns;
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 88,
        exportable: false,
        cell: (row) => (
          <div className="flex items-center justify-end gap-1">
            {canEdit && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Editar ${row.businessName}`}
                onClick={() => {
                  setEditing(row);
                  setFormOpen(true);
                }}
              >
                <Icons.Action.Edit />
              </Button>
            )}
            {canDelete && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Eliminar ${row.businessName}`}
                onClick={() => setDeleteTarget(row)}
              >
                <Icons.Action.Delete />
              </Button>
            )}
          </div>
        ),
      },
    ];
  }, [canEdit, canDelete]);

  return (
    <PageContainer>
      <PageHeader
        title="Proveedores"
        subtitle="Gestión de proveedores y cuentas por pagar"
        actions={
          <div className="flex items-center gap-2">
            <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Buscar proveedor…" className="w-64" aria-label="Buscar proveedor" />
            <Can permission={Permissions.Suppliers.Create}>
              <Button
                onClick={() => {
                  setEditing(null);
                  setFormOpen(true);
                }}
              >
                <Icons.Action.Create /> Nuevo proveedor
              </Button>
            </Can>
          </div>
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
        globalSearch={false}
        exportFilename="proveedores.csv"
        toolbarLeft={
          <Select value={status} onValueChange={(v) => setStatus(v === 'all' ? '' : v)}>
            <SelectTrigger className="w-44" aria-label="Filtrar por estado">
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
            title="No hay proveedores registrados"
            description="Cuando se registren proveedores, aparecerán aquí con sus cuentas por pagar."
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
